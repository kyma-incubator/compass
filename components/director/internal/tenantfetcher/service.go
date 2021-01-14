package tenantfetcher

import (
	"context"
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type TenantFieldMapping struct {
	TotalPagesField   string `envconfig:"APP_TENANT_TOTAL_PAGES_FIELD"`
	TotalResultsField string `envconfig:"APP_TENANT_TOTAL_RESULTS_FIELD"`
	EventsField       string `envconfig:"APP_TENANT_EVENTS_FIELD"`

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

//go:generate mockery -name=EventAPIClient -output=automock -outpkg=automock -case=underscore
type EventAPIClient interface {
	FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (TenantEventsResponse, error)
}

const (
	retryAttempts          = 7
	retryDelayMilliseconds = 100
)

type Service struct {
	queryConfig          QueryConfig
	transact             persistence.Transactioner
	kubeClient           KubeClient
	eventAPIClient       EventAPIClient
	tenantStorageService TenantStorageService
	providerName         string
	fieldMapping         TenantFieldMapping

	retryAttempts uint
}

func NewService(queryConfig QueryConfig, transact persistence.Transactioner, kubeClient KubeClient, fieldMapping TenantFieldMapping, providerName string, client EventAPIClient, tenantStorageService TenantStorageService) *Service {
	return &Service{
		transact:             transact,
		kubeClient:           kubeClient,
		fieldMapping:         fieldMapping,
		providerName:         providerName,
		eventAPIClient:       client,
		tenantStorageService: tenantStorageService,
		queryConfig:          queryConfig,

		retryAttempts: retryAttempts,
	}
}

func (s Service) SyncTenants() error {
	ctx := context.Background()
	startTime := time.Now()

	lastConsumedTenantTimestamp, err := s.kubeClient.GetTenantFetcherConfigMapData(ctx)
	if err != nil {
		return err
	}

	tenantsToCreate, err := s.getTenantsToCreate(lastConsumedTenantTimestamp)

	if err != nil {
		return err
	}
	tenantsToCreate = s.dedupeTenants(tenantsToCreate)

	tenantsToDelete, err := s.getTenantsToDelete(lastConsumedTenantTimestamp)
	if err != nil {
		return err
	}

	totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete)
	log.C(ctx).Printf("Amount of new events: %d", totalNewEvents)
	if totalNewEvents == 0 {
		return nil
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
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

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

	err = s.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixNanoString(startTime))
	if err != nil {
		return err
	}

	return nil
}

func (s Service) getTenantsToCreate(fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	var tenantsToCreate []model.BusinessTenantMappingInput

	createdTenants, err := s.fetchTenantsWithRetries(CreatedEventsType, fromTimestamp)
	if err != nil {
		return nil, err
	}
	tenantsToCreate = append(tenantsToCreate, createdTenants...)

	updatedTenants, err := s.fetchTenantsWithRetries(UpdatedEventsType, fromTimestamp)
	if err != nil {
		return nil, err
	}
	tenantsToCreate = append(tenantsToCreate, updatedTenants...)

	return tenantsToCreate, nil
}

func (s Service) getTenantsToDelete(fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	return s.fetchTenantsWithRetries(DeletedEventsType, fromTimestamp)
}

func (s Service) fetchTenantsWithRetries(eventsType EventsType, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	var tenants []model.BusinessTenantMappingInput
	err := retry.Do(func() error {
		fetchedTenants, err := s.fetchTenants(eventsType, fromTimestamp)
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

func (s Service) fetchTenants(eventsType EventsType, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
	}
	firstPage, err := s.eventAPIClient.FetchTenantEventsPage(eventsType, params)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching tenant events page")
	}
	if firstPage == nil {
		return nil, nil
	}

	tenants := make([]model.BusinessTenantMappingInput, 0)
	tenants = append(tenants, s.extractTenantMappings(eventsType, firstPage)...)
	initialCount := gjson.GetBytes(firstPage, s.fieldMapping.TotalResultsField).Int()
	totalPages := gjson.GetBytes(firstPage, s.fieldMapping.TotalPagesField).Int()

	pageStart, err := strconv.ParseInt(s.queryConfig.PageStartValue, 10, 64)
	if err != nil {
		return nil, err
	}
	for i := pageStart + 1; i <= totalPages; i++ {
		params[s.queryConfig.PageNumField] = strconv.FormatInt(i, 10)
		res, err := s.eventAPIClient.FetchTenantEventsPage(eventsType, params)
		if err != nil {
			return nil, errors.Wrap(err, "while fetching tenant events page")
		}
		if res == nil {
			return nil, apperrors.NewInternalError("next page was expected but response was empty")
		}
		if initialCount != gjson.GetBytes(res, s.fieldMapping.TotalResultsField).Int() {
			return nil, apperrors.NewInternalError("total results number changed during fetching consecutive events pages")
		}
		tenants = append(tenants, s.extractTenantMappings(eventsType, res)...)
	}

	return tenants, nil
}

func (s Service) extractTenantMappings(eventType EventsType, eventsJSON []byte) []model.BusinessTenantMappingInput {
	bussinessTenantMappings := make([]model.BusinessTenantMappingInput, 0)
	gjson.GetBytes(eventsJSON, s.fieldMapping.EventsField).ForEach(func(key gjson.Result, event gjson.Result) bool {
		detailsType := event.Get(s.fieldMapping.DetailsField).Type
		var details []byte
		if detailsType == gjson.String {
			details = []byte(gjson.Parse(event.Get(s.fieldMapping.DetailsField).String()).Raw)
		} else if detailsType == gjson.JSON {
			details = []byte(event.Get(s.fieldMapping.DetailsField).Raw)
		} else {
			log.D().Warnf("Invalid event data format: %+v", event)
			return true
		}

		tenant, err := s.eventDataToTenant(eventType, details)
		if err != nil {
			log.D().Warnf("Error: %s. Could not convert tenant: %s", err.Error(), string(details))
			return true
		}
		bussinessTenantMappings = append(bussinessTenantMappings, *tenant)
		return true
	})
	return bussinessTenantMappings
}

func (s Service) eventDataToTenant(eventType EventsType, eventData []byte) (*model.BusinessTenantMappingInput, error) {
	if eventType == CreatedEventsType && s.fieldMapping.DiscriminatorField != "" {
		discriminator, ok := gjson.GetBytes(eventData, s.fieldMapping.DiscriminatorField).Value().(string)
		if !ok {
			return nil, errors.Errorf("invalid format of %s field", s.fieldMapping.DiscriminatorField)
		}

		if discriminator != s.fieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}

	id, ok := gjson.GetBytes(eventData, s.fieldMapping.IDField).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", s.fieldMapping.IDField)
	}

	name, ok := gjson.GetBytes(eventData, s.fieldMapping.NameField).Value().(string)
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

func convertTimeToUnixNanoString(timestamp time.Time) string {
	return strconv.FormatInt(timestamp.UnixNano()/int64(time.Millisecond), 10)
}
