package tenantfetcher

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const SubaccountLabelKey = "global_subaccount_id"

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

type MovedSubaccountFieldMapping struct {
	IDField      string `envconfig:"default=id,APP_MAPPING_FIELD_ID"`
	SourceGlobal string `envconfig:"default=sourceGlobalAccountGUID,APP_MOVED_SUBACCOUNT_SOURCE_FIELD"`
	TargetGlobal string `envconfig:"default=targetGlobalAccountGUID,APP_MOVED_SUBACCOUNT_TARGET_FIELD"`
}

// QueryConfig contains the name of query parameters fields and default/start values
type QueryConfig struct {
	PageNumField   string `envconfig:"default=pageNum,APP_QUERY_PAGE_NUM_FIELD"`
	PageSizeField  string `envconfig:"default=pageSize,APP_QUERY_PAGE_SIZE_FIELD"`
	TimestampField string `envconfig:"default=timestamp,APP_QUERY_TIMESTAMP_FIELD"`
	PageStartValue string `envconfig:"default=0,APP_QUERY_PAGE_START"`
	PageSizeValue  string `envconfig:"default=150,APP_QUERY_PAGE_SIZE"`
}

//go:generate mockery --name=TenantStorageService --output=automock --outpkg=automock --case=underscore
type TenantStorageService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
	CreateManyIfNotExists(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error
	DeleteMany(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error
}

//go:generate mockery --name=LabelDefinitionService --output=automock --outpkg=automock --case=underscore
type LabelDefinitionService interface {
	Upsert(ctx context.Context, def model.LabelDefinition) error
}

//go:generate mockery --name=EventAPIClient --output=automock --outpkg=automock --case=underscore
type EventAPIClient interface {
	FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (TenantEventsResponse, error)
}

//go:generate mockery --name=RuntimeStorageService --output=automock --outpkg=automock --case=underscore
type RuntimeStorageService interface {
	GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	Update(ctx context.Context, id string, in model.RuntimeInput) error
	UpdateTenantID(ctx context.Context, runtimeID, newTenantID string) error
}

type syncTenantsAction struct {
	name string
	fn   func() error
}

const (
	retryAttempts          = 7
	retryDelayMilliseconds = 100
)

type Service struct {
	queryConfig                 QueryConfig
	transact                    persistence.Transactioner
	kubeClient                  KubeClient
	eventAPIClient              EventAPIClient
	tenantStorageService        TenantStorageService
	runtimeStorageService       RuntimeStorageService
	providerName                string
	fieldMapping                TenantFieldMapping
	movedSubaccountFieldMapping MovedSubaccountFieldMapping
	labelDefService             LabelDefinitionService

	retryAttempts uint
}

func NewService(queryConfig QueryConfig, transact persistence.Transactioner, kubeClient KubeClient, fieldMapping TenantFieldMapping, movSubAcc MovedSubaccountFieldMapping, providerName string, client EventAPIClient, tenantStorageService TenantStorageService, runtimeStorageService RuntimeStorageService, labelDefService LabelDefinitionService) *Service {
	return &Service{
		transact:                    transact,
		kubeClient:                  kubeClient,
		fieldMapping:                fieldMapping,
		providerName:                providerName,
		eventAPIClient:              client,
		tenantStorageService:        tenantStorageService,
		runtimeStorageService:       runtimeStorageService,
		queryConfig:                 queryConfig,
		movedSubaccountFieldMapping: movSubAcc,
		retryAttempts:               retryAttempts,
		labelDefService:             labelDefService,
	}
}

func (a syncTenantsAction) Execute() error {
	if err := a.fn(); err != nil {
		return errors.Wrap(err, a.name)
	}

	return nil
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
	//TODO: Check for created and then moved tenants
	tenantsToCreate = s.dedupeTenants(tenantsToCreate)

	tenantsToDelete, err := s.getTenantsToDelete(lastConsumedTenantTimestamp)
	if err != nil {
		return err
	}

	subAccountsToMove, err := s.getSubaccountsToMove(lastConsumedTenantTimestamp)
	if err != nil {
		return err
	}

	totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete) + len(subAccountsToMove)
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

	//TODO: Check whether GAs created by this transaction are viewed when querying
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var currentTenants []*model.BusinessTenantMapping = nil
	var currentTenantsMap map[string]bool = nil

	getCurrentTenants := func() (map[string]bool, error) {
		if currentTenantsMap != nil {
			return currentTenantsMap, nil
		}

		var listErr error = nil
		currentTenants, listErr = s.tenantStorageService.List(ctx)

		if listErr != nil {
			return nil, errors.Wrap(listErr, "while listing tenants")
		}

		currentTenantsMap = make(map[string]bool)
		for _, ct := range currentTenants {
			currentTenantsMap[ct.ExternalTenant] = true
		}

		return currentTenantsMap, nil
	}

	actions := make([]*syncTenantsAction, 0)
	if len(tenantsToCreate) == 0 && len(tenantsToDelete) == 0 {
		actions = append(actions, &syncTenantsAction{"while moving subaccounts", func() error { return s.moveSubAccounts(ctx, subAccountsToMove) }})
	} else {
		actions = append(actions,
			&syncTenantsAction{"while storing tenants", func() error { return s.createTenants(ctx, getCurrentTenants, tenantsToCreate) }},
			&syncTenantsAction{"while moving subaccounts", func() error { return s.moveSubAccounts(ctx, subAccountsToMove) }},
			&syncTenantsAction{"while deleting tenants", func() error { return s.deleteTenants(ctx, getCurrentTenants, tenantsToDelete) }})
	}

	for _, action := range actions {
		if err = action.Execute(); err != nil {
			return errors.Wrap(err, "while processing events")
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	if err = s.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixNanoString(startTime)); err != nil {
		return err
	}

	return nil
}

func (s Service) createTenants(ctx context.Context, getCurrentTenants func() (map[string]bool, error), eventsTenants []model.BusinessTenantMappingInput) error {
	currTenants, err := getCurrentTenants()
	if err != nil {
		return err
	}

	tenantsToCreate := make([]model.BusinessTenantMappingInput, 0)
	for i := len(eventsTenants) - 1; i >= 0; i-- {
		if currTenants[eventsTenants[i].ExternalTenant] {
			continue
		}
		tenantsToCreate = append(tenantsToCreate, eventsTenants[i])
	}

	if err = s.tenantStorageService.CreateManyIfNotExists(ctx, tenantsToCreate); err != nil {
		return errors.Wrap(err, "while storing new tenants")
	}

	return nil
}

func (s Service) deleteTenants(ctx context.Context, getCurrentTenants func() (map[string]bool, error), eventsTenants []model.BusinessTenantMappingInput) error {
	currTenants, err := getCurrentTenants()

	if err != nil {
		return err
	}

	tenantsToDelete := make([]model.BusinessTenantMappingInput, 0)
	for _, toDelete := range eventsTenants {
		if currTenants[toDelete.ExternalTenant] {
			tenantsToDelete = append(tenantsToDelete, toDelete)
		}
	}

	if err = s.tenantStorageService.DeleteMany(ctx, tenantsToDelete); err != nil {
		return errors.Wrap(err, "while removing tenants")
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

func (s Service) getSubaccountsToMove(fromTimestamp string) ([]model.MovedSubaccountMappingInput, error) {
	return s.fetchSubAccountsWithRetries(MovedSubAccountEventsType, fromTimestamp)
}

func (s Service) fetchTenantsWithRetries(eventsType EventsType, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	var tenants []model.BusinessTenantMappingInput
	err := s.fetchWithRetries(func() error {
		fetchedTenants, err := s.fetchTenants(eventsType, fromTimestamp)
		if err != nil {
			return err
		}
		tenants = fetchedTenants
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func (s Service) fetchSubAccountsWithRetries(eventsType EventsType, fromTimestamp string) ([]model.MovedSubaccountMappingInput, error) {
	var tenants []model.MovedSubaccountMappingInput
	err := s.fetchWithRetries(func() error {
		fetchedTenants, err := s.fetchMovedSubaccounts(eventsType, fromTimestamp)
		if err != nil {
			return err
		}
		tenants = fetchedTenants
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func (s Service) fetchWithRetries(applyFunc func() error) error {
	err := retry.Do(applyFunc, retry.Attempts(s.retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))

	if err != nil {
		return err
	}
	return nil
}

func (s Service) fetchTenants(eventsType EventsType, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	tenants := make([]model.BusinessTenantMappingInput, 0)

	err := s.walkThroughPages(eventsType, fromTimestamp, func(page *eventsPage) error {
		mappings, err := page.getTenantMappings(eventsType)
		if err != nil {
			return err
		}
		tenants = append(tenants, mappings...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func (s Service) fetchMovedSubaccounts(eventsType EventsType, fromTimestamp string) ([]model.MovedSubaccountMappingInput, error) {
	subaccMappings := make([]model.MovedSubaccountMappingInput, 0)

	err := s.walkThroughPages(eventsType, fromTimestamp, func(page *eventsPage) error {
		mappings, err := page.getMovedSubaccounts()
		if err != nil {
			return err
		}
		subaccMappings = append(subaccMappings, mappings...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return subaccMappings, nil
}

func (s Service) walkThroughPages(eventsType EventsType, fromTimestamp string, applyFunc func(*eventsPage) error) error {
	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
	}
	firstPage, err := s.eventAPIClient.FetchTenantEventsPage(eventsType, params)
	if err != nil {
		return errors.Wrap(err, "while fetching tenant events page")
	}
	if firstPage == nil {
		return nil
	}

	err = applyFunc(s.eventsPage(firstPage))
	if err != nil {
		return err
	}

	initialCount := gjson.GetBytes(firstPage, s.fieldMapping.TotalResultsField).Int()
	totalPages := gjson.GetBytes(firstPage, s.fieldMapping.TotalPagesField).Int()

	pageStart, err := strconv.ParseInt(s.queryConfig.PageStartValue, 10, 64)
	if err != nil {
		return err
	}

	for i := pageStart + 1; i <= totalPages; i++ {
		params[s.queryConfig.PageNumField] = strconv.FormatInt(i, 10)
		res, err := s.eventAPIClient.FetchTenantEventsPage(eventsType, params)
		if err != nil {
			return errors.Wrap(err, "while fetching tenant events page")
		}
		if res == nil {
			return apperrors.NewInternalError("next page was expected but response was empty")
		}
		if initialCount != gjson.GetBytes(res, s.fieldMapping.TotalResultsField).Int() {
			return apperrors.NewInternalError("total results number changed during fetching consecutive events pages")
		}

		if err = applyFunc(s.eventsPage(res)); err != nil {
			return err
		}
	}

	return nil
}

func (s Service) moveSubAccounts(ctx context.Context, subAccMappings []model.MovedSubaccountMappingInput) error {
	for _, subAcc := range subAccMappings {
		filters := []*labelfilter.LabelFilter{
			{
				Key:   SubaccountLabelKey,
				Query: str.Ptr(fmt.Sprintf("\"%s\"", subAcc.SubaccountID)),
			},
		}

		//TODO: change magic values
		runtime, err := s.runtimeStorageService.GetByFiltersGlobal(ctx, filters)

		if err != nil {
			if apperrors.IsNotFoundError(err) {
				log.D().Debugf("No runtime found for moved subaccount %s", subAcc.SubaccountID)
				continue
			}
			return errors.Wrapf(err, "while listing runtimes for subaccount")
		}

		targetInternalTenant, err := s.tenantStorageService.GetInternalTenant(ctx, subAcc.TargetGlobal)
		if err != nil {
			return errors.Wrapf(err, "while getting internal tenant ID for external tenant ID %s", subAcc.TargetGlobal)
		}

		labelDef := model.LabelDefinition{
			Tenant: targetInternalTenant,
			Key:    SubaccountLabelKey,
		}

		if err := s.labelDefService.Upsert(ctx, labelDef); err != nil {
			return errors.Errorf("while upserting label definition to internal tenant with ID %s", targetInternalTenant)
		}

		err = s.runtimeStorageService.UpdateTenantID(ctx, runtime.ID, targetInternalTenant)
		if err != nil {
			return errors.Errorf("while updating tenant ID runtime of subaccount with ID %s", subAcc.SubaccountID)
		}
	}

	return nil
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

func (s Service) eventsPage(payload []byte) *eventsPage {
	return &eventsPage{
		fieldMapping:                s.fieldMapping,
		movedSubaccountFieldMapping: s.movedSubaccountFieldMapping,
		payload:                     payload,
		providerName:                s.providerName,
	}
}

func convertTimeToUnixNanoString(timestamp time.Time) string {
	return strconv.FormatInt(timestamp.UnixNano()/int64(time.Millisecond), 10)
}

type eventsPage struct {
	fieldMapping                TenantFieldMapping
	movedSubaccountFieldMapping MovedSubaccountFieldMapping
	providerName                string
	payload                     []byte
}

func (ep eventsPage) getEventsDetails() [][]byte {
	tenantDetails := make([][]byte, 0)
	gjson.GetBytes(ep.payload, ep.fieldMapping.EventsField).ForEach(func(key gjson.Result, event gjson.Result) bool {
		detailsType := event.Get(ep.fieldMapping.DetailsField).Type
		var details []byte
		if detailsType == gjson.String {
			details = []byte(gjson.Parse(event.Get(ep.fieldMapping.DetailsField).String()).Raw)
		} else if detailsType == gjson.JSON {
			details = []byte(event.Get(ep.fieldMapping.DetailsField).Raw)
		} else {
			log.D().Warnf("Invalid event data format: %+v", event)
			return true
		}

		tenantDetails = append(tenantDetails, details)
		return true
	})
	return tenantDetails
}

func (ep eventsPage) getMovedSubaccounts() ([]model.MovedSubaccountMappingInput, error) {
	eds := ep.getEventsDetails()
	subaccMappings := make([]model.MovedSubaccountMappingInput, 0, len(eds))
	for _, detail := range eds {
		mapping, err := ep.eventDataToChangedSubAccount(detail)
		if err != nil {
			return nil, err
		}

		subaccMappings = append(subaccMappings, *mapping)
	}

	return subaccMappings, nil
}

func (ep eventsPage) getTenantMappings(eventsType EventsType) ([]model.BusinessTenantMappingInput, error) {
	eds := ep.getEventsDetails()
	tenants := make([]model.BusinessTenantMappingInput, 0, len(eds))
	for _, detail := range eds {
		mapping, err := ep.eventDataToTenant(eventsType, detail)
		if err != nil {
			return nil, err
		}

		tenants = append(tenants, *mapping)
	}

	return tenants, nil
}

func (ep eventsPage) eventDataToChangedSubAccount(eventData []byte) (*model.MovedSubaccountMappingInput, error) {
	id, ok := gjson.GetBytes(eventData, ep.movedSubaccountFieldMapping.IDField).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedSubaccountFieldMapping.IDField)
	}

	source, ok := gjson.GetBytes(eventData, ep.movedSubaccountFieldMapping.SourceGlobal).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedSubaccountFieldMapping.SourceGlobal)
	}

	target, ok := gjson.GetBytes(eventData, ep.movedSubaccountFieldMapping.TargetGlobal).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.movedSubaccountFieldMapping.TargetGlobal)
	}

	return &model.MovedSubaccountMappingInput{
		SubaccountID: id,
		SourceGlobal: source,
		TargetGlobal: target,
	}, nil
}

func (ep eventsPage) eventDataToTenant(eventType EventsType, eventData []byte) (*model.BusinessTenantMappingInput, error) {
	if eventType == CreatedEventsType && ep.fieldMapping.DiscriminatorField != "" {
		discriminator, ok := gjson.GetBytes(eventData, ep.fieldMapping.DiscriminatorField).Value().(string)
		if !ok {
			return nil, errors.Errorf("invalid format of %s field", ep.fieldMapping.DiscriminatorField)
		}

		if discriminator != ep.fieldMapping.DiscriminatorValue {
			return nil, nil
		}
	}

	id, ok := gjson.GetBytes(eventData, ep.fieldMapping.IDField).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.fieldMapping.IDField)
	}

	name, ok := gjson.GetBytes(eventData, ep.fieldMapping.NameField).Value().(string)
	if !ok {
		return nil, errors.Errorf("invalid format of %s field", ep.fieldMapping.NameField)
	}

	return &model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: id,
		Provider:       ep.providerName,
	}, nil
}
