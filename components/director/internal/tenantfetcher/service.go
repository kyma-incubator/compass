package tenantfetcher

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

type TenantFieldMapping struct {
	NameField          string `envconfig:"default=name,APP_MAPPING_FIELD_NAME"`
	IDField            string `envconfig:"default=id,APP_MAPPING_FIELD_ID"`
	DetailsField       string `envconfig:"default=details,APP_MAPPING_FIELD_DETAILS"`
	DiscriminatorField string `envconfig:"optional,APP_MAPPING_FIELD_DISCRIMINATOR"`
	DiscriminatorValue string `envconfig:"optional,APP_MAPPING_VALUE_DISCRIMINATOR"`
}

// QueryConfig contains the name of query parameters fields and default/start values
type QueryConfig struct {
	PageNumField   string `envconfig:"default=pageNum,APP_QUERY_PAGE_NUM_FIELD"`
	PageSizeField  string `envconfig:"default=pageSize,APP_QUERY_PAGE_SIZE_FIELD"`
	TimestampField string `envconfig:"default=timestamp,APP_QUERY_TIMESTAMP_FIELD"`
	PageStartValue string `envconfig:"default=0,APP_QUERY_PAGE_START"`
	PageSizeValue  string `envconfig:"default=150,APP_QUERY_PAGE_SIZE"`
}

//go:generate mockery -name=TenantStorageService -output=automock -outpkg=automock -case=underscore
type TenantStorageService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	CreateManyIfNotExists(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error
	DeleteMany(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error
}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	EventsToTenants(eventsType EventsType, events []Event) []model.BusinessTenantMappingInput
}

//go:generate mockery -name=EventAPIClient -output=automock -outpkg=automock -case=underscore
type EventAPIClient interface {
	FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (*TenantEventsResponse, error)
}

const (
	retryAttempts          = 7
	retryDelayMilliseconds = 100
)

type Service struct {
	queryConfig QueryConfig
	transact    persistence.Transactioner
	// converter            Converter
	eventAPIClient       EventAPIClient
	tenantStorageService TenantStorageService
	providerName         string
	fieldMapping         TenantFieldMapping

	retryAttempts uint
}

func NewService(queryConfig QueryConfig, transact persistence.Transactioner, fieldMapping TenantFieldMapping, providerName string, client EventAPIClient, tenantStorageService TenantStorageService) *Service {
	return &Service{
		transact:             transact,
		fieldMapping:         fieldMapping,
		eventAPIClient:       client,
		tenantStorageService: tenantStorageService,
		queryConfig:          queryConfig,

		retryAttempts: retryAttempts,
	}
}

func (s Service) SyncTenants() error {
	tenantsToCreate, err := s.getTenantsToCreate()
	if err != nil {
		return err
	}
	tenantsToCreate = s.dedupeTenants(tenantsToCreate)

	tenantsToDelete, err := s.getTenantsToDelete()
	if err != nil {
		return err
	}

	deleteTenantsMap := make(map[string]model.BusinessTenantMappingInput)
	for _, ct := range tenantsToDelete {
		deleteTenantsMap[ct.ExternalTenant] = ct
	}

	for i := len(tenantsToCreate) - 1; i >= 0; i-- {
		if _, found := deleteTenantsMap[tenantsToCreate[i].ExternalTenant]; found {
			tenantsToCreate = append(tenantsToCreate[:i], tenantsToCreate[i+1:]...)
		}
	}

	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(tx)
	ctx := context.Background()
	ctx = persistence.SaveToContext(ctx, tx)

	currentTenants, err := s.tenantStorageService.List(ctx)
	if err != nil {
		return errors.Wrap(err, "while listing tenants")
	}

	currentTenantsMap := make(map[string]bool)
	for _, ct := range currentTenants {
		currentTenantsMap[ct.ExternalTenant] = true
	}

	for i := len(tenantsToCreate) - 1; i >= 0; i-- {
		if currentTenantsMap[tenantsToCreate[i].ExternalTenant] {
			tenantsToCreate = append(tenantsToCreate[:i], tenantsToCreate[i+1:]...)
		}
	}

	tenantsToDelete = make([]model.BusinessTenantMappingInput, 0)
	for _, toDelete := range deleteTenantsMap {
		if currentTenantsMap[toDelete.ExternalTenant] {
			tenantsToDelete = append(tenantsToDelete, toDelete)
		}
	}

	err = s.tenantStorageService.CreateManyIfNotExists(ctx, tenantsToCreate)
	if err != nil {
		return errors.Wrap(err, "while storing new tenants")
	}
	err = s.tenantStorageService.DeleteMany(ctx, tenantsToDelete)
	if err != nil {
		return errors.Wrap(err, "while removing tenants")
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s Service) getTenantsToCreate() ([]model.BusinessTenantMappingInput, error) {
	var tenantsToCreate []model.BusinessTenantMappingInput

	createdTenants, err := s.fetchTenantsWithRetries(CreatedEventsType)
	if err != nil {
		return nil, err
	}
	tenantsToCreate = append(tenantsToCreate, createdTenants...)

	updatedTenants, err := s.fetchTenantsWithRetries(UpdatedEventsType)
	if err != nil {
		return nil, err
	}
	tenantsToCreate = append(tenantsToCreate, updatedTenants...)

	return tenantsToCreate, nil
}

func (s Service) getTenantsToDelete() ([]model.BusinessTenantMappingInput, error) {
	return s.fetchTenantsWithRetries(DeletedEventsType)
}

func (s Service) fetchTenantsWithRetries(eventsType EventsType) ([]model.BusinessTenantMappingInput, error) {
	var tenants []model.BusinessTenantMappingInput
	err := retry.Do(func() error {
		fetchedTenants, err := s.fetchTenants(eventsType)
		if err != nil {
			return err
		}
		tenants = fetchedTenants
		return nil
	}, retry.Attempts(s.retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s Service) fetchTenants(eventsType EventsType) ([]model.BusinessTenantMappingInput, error) {
	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: strconv.FormatInt(1, 10),
	}
	firstPage, err := s.eventAPIClient.FetchTenantEventsPage(eventsType, params)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching tenant events page")
	}
	if firstPage == nil {
		return nil, nil
	}

	tenants := make([]model.BusinessTenantMappingInput, 0)
	tenants = append(tenants, s.convert(eventsType, firstPage.Events)...)
	// events := firstPage.Events
	// events = append(events, firstPage.Events)

	initialCount := firstPage.TotalResults
	totalPages := firstPage.TotalPages

	pageStart, err := strconv.Atoi(s.queryConfig.PageStartValue)
	if err != nil {
		return nil, err
	}
	for i := pageStart + 1; i <= totalPages; i++ {
		params[s.queryConfig.PageNumField] = strconv.Itoa(i)
		res, err := s.eventAPIClient.FetchTenantEventsPage(eventsType, params)
		if err != nil {
			return nil, errors.Wrap(err, "while fetching tenant events page")
		}
		if res == nil {
			return nil, apperrors.NewInternalError("next page was expected but response was empty")
		}
		if initialCount != res.TotalResults {
			return nil, apperrors.NewInternalError("total results number changed during fetching consecutive events pages")
		}
		// events = append(events, res.Events)
		tenants = append(tenants, s.convert(eventsType, res.Events)...)
	}

	return nil, nil
	// return s.converter.EventsToTenants(eventsType, events), nil
}

func (s Service) convert(eventType EventsType, eventsJSON []byte) []model.BusinessTenantMappingInput {
	events := make([]model.BusinessTenantMappingInput, 0)
	gjson.GetBytes(eventsJSON, "").ForEach(func(key gjson.Result, event gjson.Result) bool {
		var eventData map[string]interface{}
		details := event.Get(s.fieldMapping.DetailsField).Value()
		switch details.(type) {
		case string:
			eventDataJSON, ok := details.(string)
			if !ok {
				log.Warnf("Invalid event data format: %+v", event)
				return true
			}
			err := json.Unmarshal([]byte(eventDataJSON), &eventData)
			if err != nil {
				log.Warnf("Could not unmarshal event data", event)
				return true
			}
		case map[string]interface{}:
			var ok bool
			eventData, ok = details.(map[string]interface{})
			if !ok {
				log.Warnf("Invalid event data format: %+v", event)
				return true
			}
			tenant, err := s.eventDataToTenant(eventType, eventData)
			if err != nil {
				log.Warnf("Could not convert tenant: %s", err.Error())
				return true
			}
			events = append(events, *tenant)
		default:
			log.Warnf("Invalid event data format: %+v", event)
			return true
		}
		return true
	})
	return events
}

func (s Service) eventDataToTenant(eventType EventsType, eventData map[string]interface{}) (*model.BusinessTenantMappingInput, error) {
	if eventType == CreatedEventsType && s.fieldMapping.DiscriminatorField != "" {
		discriminator, ok := eventData[s.fieldMapping.DiscriminatorField].(string)
		if !ok {
			return nil, errors.Errorf("invalid format of %s field", s.fieldMapping.DiscriminatorField)
		}

		if discriminator != s.fieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}

	id, ok := eventData[s.fieldMapping.IDField].(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", s.fieldMapping.IDField)
	}

	name, ok := eventData[s.fieldMapping.NameField].(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", s.fieldMapping.NameField)
	}

	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: id,
		Provider:       s.providerName,
	}, nil
}

func (s Service) dedupeTenants(tenants []model.BusinessTenantMappingInput) []model.BusinessTenantMappingInput {
	elms := make(map[string]model.BusinessTenantMappingInput)
	for _, tc := range tenants {
		elms[tc.ExternalTenant] = tc
	}
	tenants = make([]model.BusinessTenantMappingInput, 0, len(elms))
	for _, t := range elms {
		tenants = append(tenants, t)
	}
	return tenants
}
