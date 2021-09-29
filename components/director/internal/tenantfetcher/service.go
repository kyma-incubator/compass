package tenantfetcher

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

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

// TenantFieldMapping missing godoc
type TenantFieldMapping struct {
	TotalPagesField   string `envconfig:"APP_TENANT_TOTAL_PAGES_FIELD"`
	TotalResultsField string `envconfig:"APP_TENANT_TOTAL_RESULTS_FIELD"`
	EventsField       string `envconfig:"APP_TENANT_EVENTS_FIELD"`

	NameField          string `envconfig:"default=name,APP_MAPPING_FIELD_NAME"`
	IDField            string `envconfig:"default=id,APP_MAPPING_FIELD_ID"`
	CustomerIDField    string `envconfig:"default=customerId,APP_MAPPING_FIELD_CUSTOMER_ID"`
	SubdomainField     string `envconfig:"default=subdomain,APP_MAPPING_FIELD_SUBDOMAIN"`
	DetailsField       string `envconfig:"default=details,APP_MAPPING_FIELD_DETAILS"`
	DiscriminatorField string `envconfig:"optional,APP_MAPPING_FIELD_DISCRIMINATOR"`
	DiscriminatorValue string `envconfig:"optional,APP_MAPPING_VALUE_DISCRIMINATOR"`

	RegionField     string `envconfig:"default=region,APP_MAPPING_FIELD_REGION"`
	EntityTypeField string `envconfig:"default=entityType,APP_MAPPING_FIELD_ENTITY_TYPE"`
	ParentIDField   string `envconfig:"default=entityType,APP_MAPPING_FIELD_PARENT_ID"`
}

// MovedRuntimeByLabelFieldMapping missing godoc
type MovedRuntimeByLabelFieldMapping struct {
	LabelValue   string `envconfig:"default=id,APP_MAPPING_FIELD_ID"`
	SourceTenant string `envconfig:"default=sourceTenant,APP_MOVED_RUNTIME_BY_LABEL_SOURCE_TENANT_FIELD"`
	TargetTenant string `envconfig:"default=targetTenant,APP_MOVED_RUNTIME_BY_LABEL_TARGET_TENANT_FIELD"`
}

// QueryConfig contains the name of query parameters fields and default/start values
type QueryConfig struct {
	PageNumField   string `envconfig:"default=pageNum,APP_QUERY_PAGE_NUM_FIELD"`
	PageSizeField  string `envconfig:"default=pageSize,APP_QUERY_PAGE_SIZE_FIELD"`
	TimestampField string `envconfig:"default=timestamp,APP_QUERY_TIMESTAMP_FIELD"`
	RegionField    string `envconfig:"default=region,APP_QUERY_REGION_FIELD"`
	PageStartValue string `envconfig:"default=0,APP_QUERY_PAGE_START"`
	PageSizeValue  string `envconfig:"default=150,APP_QUERY_PAGE_SIZE"`
}

type PageConfig struct {
	TotalPagesField   string
	TotalResultsField string
	PageStartValue    string // todo duplicated
	PageNumField      string
}

// TenantStorageService missing godoc
//go:generate mockery --name=TenantStorageService --output=automock --outpkg=automock --case=underscore --unroll-variadic=False
type TenantStorageService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
	UpsertMany(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	DeleteMany(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error
}

// LabelDefinitionService missing godoc
//go:generate mockery --name=LabelDefinitionService --output=automock --outpkg=automock --case=underscore
type LabelDefinitionService interface {
	Upsert(ctx context.Context, def model.LabelDefinition) error
}

// EventAPIClient missing godoc
//go:generate mockery --name=EventAPIClient --output=automock --outpkg=automock --case=underscore
type EventAPIClient interface {
	FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (TenantEventsResponse, error)
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	Update(ctx context.Context, id string, in model.RuntimeInput) error
	UpdateTenantID(ctx context.Context, runtimeID, newTenantID string) error
}

// TenantSyncService missing godoc
//go:generate mockery --name=TenantSyncService --output=automock --outpkg=automock --case=underscore
type TenantSyncService interface {
	SyncTenants() error
}

const (
	retryAttempts          = 7
	retryDelayMilliseconds = 100
)

// GlobalAccountService missing godoc
type GlobalAccountService struct {
	queryConfig          QueryConfig
	transact             persistence.Transactioner
	kubeClient           KubeClient
	eventAPIClient       EventAPIClient
	tenantStorageService TenantStorageService
	providerName         string
	tenantsRegion        string
	fieldMapping         TenantFieldMapping
	retryAttempts        uint
	fullResyncInterval   time.Duration
	toEventsPage         func([]byte) *eventsPage
}

// SubaccountService missing godoc
type SubaccountService struct {
	queryConfig                     QueryConfig
	transact                        persistence.Transactioner
	kubeClient                      KubeClient
	eventAPIClient                  EventAPIClient
	tenantStorageService            TenantStorageService
	runtimeStorageService           RuntimeService
	providerName                    string
	tenantsRegions                  []string
	fieldMapping                    TenantFieldMapping
	movedRuntimeByLabelFieldMapping MovedRuntimeByLabelFieldMapping
	labelDefService                 LabelDefinitionService
	retryAttempts                   uint
	movedRuntimeLabelKey            string
	fullResyncInterval              time.Duration
	toEventsPage                    func([]byte) *eventsPage
}

// NewGAService missing godoc
func NewGAService(queryConfig QueryConfig,
	transact persistence.Transactioner,
	kubeClient KubeClient,
	fieldMapping TenantFieldMapping,
	providerName string, regionName string, client EventAPIClient,
	tenantStorageService TenantStorageService,
	fullResyncInterval time.Duration) *GlobalAccountService {
	return &GlobalAccountService{
		transact:             transact,
		kubeClient:           kubeClient,
		fieldMapping:         fieldMapping,
		providerName:         providerName,
		tenantsRegion:        regionName,
		eventAPIClient:       client,
		tenantStorageService: tenantStorageService,
		queryConfig:          queryConfig,
		retryAttempts:        retryAttempts,
		fullResyncInterval:   fullResyncInterval,
		toEventsPage: func(bytes []byte) *eventsPage {
			return &eventsPage{
				fieldMapping: fieldMapping,
				payload:      bytes,
				providerName: providerName,
			}
		},
	}
}

// NewGAService missing godoc
func NewSubaccountService(queryConfig QueryConfig,
	transact persistence.Transactioner,
	kubeClient KubeClient,
	fieldMapping TenantFieldMapping,
	movRuntime MovedRuntimeByLabelFieldMapping,
	providerName string, regionNames []string, client EventAPIClient,
	tenantStorageService TenantStorageService,
	runtimeStorageService RuntimeService,
	labelDefService LabelDefinitionService,
	movedRuntimeLabelKey string,
	fullResyncInterval time.Duration) *SubaccountService {
	return &SubaccountService{
		transact:                        transact,
		kubeClient:                      kubeClient,
		fieldMapping:                    fieldMapping,
		providerName:                    providerName,
		tenantsRegions:                  regionNames,
		eventAPIClient:                  client,
		tenantStorageService:            tenantStorageService,
		runtimeStorageService:           runtimeStorageService,
		queryConfig:                     queryConfig,
		movedRuntimeByLabelFieldMapping: movRuntime,
		retryAttempts:                   retryAttempts,
		labelDefService:                 labelDefService,
		movedRuntimeLabelKey:            movedRuntimeLabelKey,
		fullResyncInterval:              fullResyncInterval,
		toEventsPage: func(bytes []byte) *eventsPage {
			return &eventsPage{
				fieldMapping:                    fieldMapping,
				movedRuntimeByLabelFieldMapping: movRuntime,
				payload:                         bytes,
				providerName:                    providerName,
			}
		},
	}
}

func (s SubaccountService) SyncTenants() error {
	ctx := context.Background()
	startTime := time.Now()

	lastConsumedTenantTimestamp, lastResyncTimestamp, err := s.kubeClient.GetTenantFetcherConfigMapData(ctx)
	if err != nil {
		return err
	}

	shouldFullResync, err := shouldFullResync(lastResyncTimestamp, s.fullResyncInterval)
	if err != nil {
		return err
	}

	newLastResyncTimestamp := lastResyncTimestamp
	if shouldFullResync {
		log.C(ctx).Infof("Last full resync was %s ago. Will perform a full resync.", s.fullResyncInterval)
		lastConsumedTenantTimestamp = "1"
		newLastResyncTimestamp = convertTimeToUnixMilliSecondString(startTime)
	}

	for _, region := range s.tenantsRegions {
		tenantsToCreate, err := s.getSubaccountsToCreateForRegion(lastConsumedTenantTimestamp, region)
		if err != nil {
			return err
		}

		tenantsToDelete, err := s.getSubaccountsToDeleteForRegion(lastConsumedTenantTimestamp, region)
		if err != nil {
			return err
		}

		runtimesToMove, err := s.getRuntimesToMoveByLabel(lastConsumedTenantTimestamp)
		if err != nil {
			return err
		}

		tenantsToCreate = dedupeTenants(tenantsToCreate)
		tenantsToCreate = excludeTenants(tenantsToCreate, tenantsToDelete)

		totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete) + len(runtimesToMove)
		log.C(ctx).Printf("Amount of new events: %d", totalNewEvents)
		if totalNewEvents == 0 {
			return nil
		}

		tx, err := s.transact.Begin()
		if err != nil {
			return err
		}
		defer s.transact.RollbackUnlessCommitted(ctx, tx)
		ctx = persistence.SaveToContext(ctx, tx)

		currentTenants := make(map[string]string)
		if len(tenantsToCreate) > 0 || len(tenantsToDelete) > 0 {
			currentTenants, err = getCurrentTenants(ctx, s.tenantStorageService)
			if err != nil {
				return err
			}
		}

		// Order of event processing matters
		if len(tenantsToCreate) > 0 {
			if err := createTenants(ctx, s.tenantStorageService, currentTenants, tenantsToCreate, region, s.providerName); err != nil {
				return errors.Wrap(err, "while storing tenant")
			}
		}
		if len(runtimesToMove) > 0 {
			if err := s.moveRuntimesByLabel(ctx, runtimesToMove); err != nil {
				return errors.Wrap(err, "moving runtimes by label")
			}
		}
		if len(tenantsToDelete) > 0 {
			if err := deleteTenants(ctx, s.tenantStorageService, currentTenants, tenantsToDelete); err != nil {
				return errors.Wrap(err, "moving deleting runtimes")
			}
		}

		if err = tx.Commit(); err != nil {
			return err
		}
		log.C(ctx).Printf("Processed new events for region: %s", region)
	}

	if err = s.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixMilliSecondString(startTime), newLastResyncTimestamp); err != nil {
		return err
	}

	return nil
}

// SyncTenants missing godoc
func (s GlobalAccountService) SyncTenants() error {
	ctx := context.Background()
	startTime := time.Now()

	lastConsumedTenantTimestamp, lastResyncTimestamp, err := s.kubeClient.GetTenantFetcherConfigMapData(ctx)
	if err != nil {
		return err
	}

	shouldFullResync, err := shouldFullResync(lastResyncTimestamp, s.fullResyncInterval)
	if err != nil {
		return err
	}

	newLastResyncTimestamp := lastResyncTimestamp
	if shouldFullResync {
		log.C(ctx).Infof("Last full resync was %s ago. Will perform a full resync.", s.fullResyncInterval)
		lastConsumedTenantTimestamp = "1"
		newLastResyncTimestamp = convertTimeToUnixMilliSecondString(startTime)
	}

	tenantsToCreate, err := s.getAccountsToCreate(lastConsumedTenantTimestamp)
	if err != nil {
		return err
	}

	tenantsToDelete, err := s.getAccountsToDelete(lastConsumedTenantTimestamp)
	if err != nil {
		return err
	}

	tenantsToCreate = dedupeTenants(tenantsToCreate)
	tenantsToCreate = excludeTenants(tenantsToCreate, tenantsToDelete)

	totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete)
	log.C(ctx).Printf("Amount of new events: %d", totalNewEvents)
	if totalNewEvents == 0 {
		return nil
	}

	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	currentTenants := make(map[string]string)
	if len(tenantsToCreate) > 0 || len(tenantsToDelete) > 0 {
		currentTenants, err = getCurrentTenants(ctx, s.tenantStorageService)
		if err != nil {
			return err
		}
	}

	// Order of event processing matters
	if len(tenantsToCreate) > 0 {
		if err := createTenants(ctx, s.tenantStorageService, currentTenants, tenantsToCreate, s.tenantsRegion, s.providerName); err != nil {
			return errors.Wrap(err, "while storing tenant")
		}
	}
	if len(tenantsToDelete) > 0 {
		if err := deleteTenants(ctx, s.tenantStorageService, currentTenants, tenantsToDelete); err != nil {
			return errors.Wrap(err, "moving deleting runtimes")
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	if err = s.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixMilliSecondString(startTime), newLastResyncTimestamp); err != nil {
		return err
	}

	return nil
}

func createTenants(ctx context.Context, tenantStorageService TenantStorageService, currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput, region string, provider string) error {
	tenantsToCreate := parents(currTenants, eventsTenants, provider)
	for _, eventTenant := range eventsTenants {
		if parentGUID, ok := currTenants[eventTenant.Parent]; ok {
			eventTenant.Parent = parentGUID
		}
		eventTenant.Region = region
		tenantsToCreate = append(tenantsToCreate, eventTenant)
	}
	if len(tenantsToCreate) > 0 {
		if err := tenantStorageService.UpsertMany(ctx, tenantsToCreate...); err != nil {
			return errors.Wrap(err, "while storing new tenants")
		}
	}
	return nil
}

func parents(currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput, providerName string) []model.BusinessTenantMappingInput {
	parentsToCreate := make([]model.BusinessTenantMappingInput, 0)
	for _, eventTenant := range eventsTenants {
		if len(eventTenant.Parent) > 0 {
			if _, ok := currTenants[eventTenant.Parent]; !ok {
				parentTenant := model.BusinessTenantMappingInput{
					Name:           eventTenant.Parent,
					ExternalTenant: eventTenant.Parent,
					Parent:         "",
					Type:           getTenantParentType(eventTenant.Type),
					Provider:       providerName,
				}
				parentsToCreate = append(parentsToCreate, parentTenant)
			}
		}
	}

	return dedupeTenants(parentsToCreate)
}

func getTenantParentType(tenantType string) string {
	if tenantType == "Subaccount" {
		return tenant.TypeToStr(tenant.Account)
	}
	return tenant.TypeToStr(tenant.Customer)
}

func (s SubaccountService) moveRuntimesByLabel(ctx context.Context, movedRuntimeMappings []model.MovedRuntimeByLabelMappingInput) error {
	for _, mapping := range movedRuntimeMappings {
		filters := []*labelfilter.LabelFilter{
			{
				Key:   s.movedRuntimeLabelKey,                             //global_subaccount_id
				Query: str.Ptr(fmt.Sprintf("\"%s\"", mapping.LabelValue)), // guid value
			},
		}

		runtime, err := s.runtimeStorageService.GetByFiltersGlobal(ctx, filters)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				log.D().Debugf("No runtime found for label key %s with value %s", s.movedRuntimeLabelKey, mapping.LabelValue)
				continue
			}
			return errors.Wrapf(err, "while listing runtimes for label key %s", s.movedRuntimeLabelKey)
		}

		targetInternalTenant, err := s.tenantStorageService.GetInternalTenant(ctx, mapping.TargetTenant) // should move runtime here
		if err != nil {
			return errors.Wrapf(err, "while getting internal tenant ID for external tenant ID %s", mapping.TargetTenant)
		}

		labelDef := model.LabelDefinition{
			Tenant: targetInternalTenant,
			Key:    s.movedRuntimeLabelKey,
		}

		if err := s.labelDefService.Upsert(ctx, labelDef); err != nil {
			return errors.Wrapf(err, "while upserting label definition to internal tenant with ID %s", targetInternalTenant)
		}

		err = s.runtimeStorageService.UpdateTenantID(ctx, runtime.ID, targetInternalTenant)
		if err != nil {
			return errors.Wrapf(err, "while updating tenant ID of runtime with label key-value match %s-%s",
				s.movedRuntimeLabelKey, mapping.LabelValue)
		}
	}

	return nil
}

func deleteTenants(ctx context.Context, tenantStorageService TenantStorageService, currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput) error {
	tenantsToDelete := make([]model.BusinessTenantMappingInput, 0)
	for _, toDelete := range eventsTenants {
		if _, ok := currTenants[toDelete.ExternalTenant]; ok {
			tenantsToDelete = append(tenantsToDelete, toDelete)
		}
	}

	if err := tenantStorageService.DeleteMany(ctx, tenantsToDelete); err != nil {
		return errors.Wrap(err, "while removing tenants")
	}

	return nil
}

func (s GlobalAccountService) getAccountsToCreate(fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	var tenantsToCreate []model.BusinessTenantMappingInput

	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
	}

	pageConfig := PageConfig{
		TotalPagesField:   s.fieldMapping.TotalPagesField,
		TotalResultsField: s.fieldMapping.TotalResultsField,
		PageStartValue:    s.queryConfig.PageStartValue,
		PageNumField:      s.queryConfig.PageNumField,
	}

	createdTenants, err := fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, CreatedAccountType, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
	if err != nil {
		return nil, err
	}
	tenantsToCreate = append(tenantsToCreate, createdTenants...)

	updatedTenants, err := fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, UpdatedAccountType, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
	if err != nil {
		return nil, err
	}

	tenantsToCreate = append(tenantsToCreate, updatedTenants...)

	return tenantsToCreate, nil
}

func (s SubaccountService) getSubaccountsToCreateForRegion(fromTimestamp string, region string) ([]model.BusinessTenantMappingInput, error) {
	var tenantsToCreate []model.BusinessTenantMappingInput

	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
		s.queryConfig.RegionField:    region,
	}
	pageConfig := PageConfig{
		TotalPagesField:   s.fieldMapping.TotalPagesField,
		TotalResultsField: s.fieldMapping.TotalResultsField,
		PageStartValue:    s.queryConfig.PageStartValue,
		PageNumField:      s.queryConfig.PageNumField,
	}
	createdTenants, err := fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, CreatedSubaccountType, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
	if err != nil {
		return nil, err
	}
	tenantsToCreate = append(tenantsToCreate, createdTenants...)

	updatedTenants, err := fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, UpdatedSubaccountType, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
	if err != nil {
		return nil, err
	}

	tenantsToCreate = append(tenantsToCreate, updatedTenants...)

	return tenantsToCreate, nil
}

func (s GlobalAccountService) getAccountsToDelete(fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
	}

	pageConfig := PageConfig{
		TotalPagesField:   s.fieldMapping.TotalPagesField,
		TotalResultsField: s.fieldMapping.TotalResultsField,
		PageStartValue:    s.queryConfig.PageStartValue,
		PageNumField:      s.queryConfig.PageNumField,
	}
	return fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, DeletedAccountType, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
}

func (s SubaccountService) getSubaccountsToDeleteForRegion(fromTimestamp string, region string) ([]model.BusinessTenantMappingInput, error) {
	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
		s.queryConfig.RegionField:    region,
	}
	pageConfig := PageConfig{
		TotalPagesField:   s.fieldMapping.TotalPagesField,
		TotalResultsField: s.fieldMapping.TotalResultsField,
		PageStartValue:    s.queryConfig.PageStartValue,
		PageNumField:      s.queryConfig.PageNumField,
	}
	return fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, DeletedSubaccountType, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
}

func (s SubaccountService) getRuntimesToMoveByLabel(fromTimestamp string) ([]model.MovedRuntimeByLabelMappingInput, error) {
	params := QueryParams{
		s.queryConfig.PageNumField:   s.queryConfig.PageStartValue,
		s.queryConfig.PageSizeField:  s.queryConfig.PageSizeValue,
		s.queryConfig.TimestampField: fromTimestamp,
	}
	pageConfig := PageConfig{
		TotalPagesField:   s.fieldMapping.TotalPagesField,
		TotalResultsField: s.fieldMapping.TotalResultsField,
		PageStartValue:    s.queryConfig.PageStartValue,
		PageNumField:      s.queryConfig.PageNumField,
	}
	return fetchMovedRuntimesWithRetries(s.eventAPIClient, s.retryAttempts, func() (QueryParams, PageConfig) { return params, pageConfig }, s.toEventsPage)
}

func fetchTenantsWithRetries(eventAPIClient EventAPIClient, retryNumber uint, eventsType EventsType, f func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.BusinessTenantMappingInput, error) {
	var tenants []model.BusinessTenantMappingInput
	err := fetchWithRetries(retryNumber, func() error {
		fetchedTenants, err := fetchTenants(eventAPIClient, eventsType, f, toEventsPage)
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

func fetchMovedRuntimesWithRetries(eventAPIClient EventAPIClient, retryNumber uint, f func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.MovedRuntimeByLabelMappingInput, error) {
	var tenants []model.MovedRuntimeByLabelMappingInput
	err := fetchWithRetries(retryNumber, func() error {
		fetchedTenants, err := fetchMovedRuntimes(eventAPIClient, f, toEventsPage)
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

func fetchWithRetries(retryAttempts uint, applyFunc func() error) error {
	err := retry.Do(applyFunc, retry.Attempts(retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))

	if err != nil {
		return err
	}
	return nil
}

func fetchTenants(eventAPIClient EventAPIClient, eventsType EventsType, f func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.BusinessTenantMappingInput, error) {
	tenants := make([]model.BusinessTenantMappingInput, 0)
	err := walkThroughPages(eventAPIClient, eventsType, f, toEventsPage, func(page *eventsPage) error {
		mappings := page.getTenantMappings(eventsType)
		tenants = append(tenants, mappings...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func fetchMovedRuntimes(eventAPIClient EventAPIClient, f func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.MovedRuntimeByLabelMappingInput, error) {
	allMappings := make([]model.MovedRuntimeByLabelMappingInput, 0)

	err := walkThroughPages(eventAPIClient, MovedSubaccountType, f, toEventsPage, func(page *eventsPage) error {
		mappings := page.getMovedRuntimes()
		allMappings = append(allMappings, mappings...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allMappings, nil
}

func walkThroughPages(eventAPIClient EventAPIClient, eventsType EventsType, f func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage, applyFunc func(*eventsPage) error) error {
	params, pageConfig := f()
	firstPage, err := eventAPIClient.FetchTenantEventsPage(eventsType, params)
	if err != nil {
		return errors.Wrap(err, "while fetching tenant events page")
	}
	if firstPage == nil {
		return nil
	}

	err = applyFunc(toEventsPage(firstPage))
	if err != nil {
		return err
	}

	initialCount := gjson.GetBytes(firstPage, pageConfig.TotalResultsField).Int()
	totalPages := gjson.GetBytes(firstPage, pageConfig.TotalPagesField).Int()

	pageStart, err := strconv.ParseInt(pageConfig.PageStartValue, 10, 64)
	if err != nil {
		return err
	}

	for i := pageStart + 1; i <= totalPages; i++ {
		params[pageConfig.PageNumField] = strconv.FormatInt(i, 10)
		res, err := eventAPIClient.FetchTenantEventsPage(eventsType, params)
		if err != nil {
			return errors.Wrap(err, "while fetching tenant events page")
		}
		if res == nil {
			return apperrors.NewInternalError("next page was expected but response was empty")
		}
		if initialCount != gjson.GetBytes(res, pageConfig.TotalResultsField).Int() {
			return apperrors.NewInternalError("total results number changed during fetching consecutive events pages")
		}

		if err = applyFunc(toEventsPage(res)); err != nil {
			return err
		}
	}

	return nil
}

func dedupeTenants(tenants []model.BusinessTenantMappingInput) []model.BusinessTenantMappingInput {
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

func excludeTenants(source, target []model.BusinessTenantMappingInput) []model.BusinessTenantMappingInput {
	deleteTenantsMap := make(map[string]model.BusinessTenantMappingInput)
	for _, ct := range target {
		deleteTenantsMap[ct.ExternalTenant] = ct
	}

	result := append([]model.BusinessTenantMappingInput{}, source...)

	for i := len(result) - 1; i >= 0; i-- {
		if _, found := deleteTenantsMap[result[i].ExternalTenant]; found {
			result = append(result[:i], result[i+1:]...)
		}
	}

	return result
}

func getCurrentTenants(ctx context.Context, tenantStorage TenantStorageService) (map[string]string, error) { // map externalTenant <--> internalTenant
	currentTenants, listErr := tenantStorage.List(ctx)
	if listErr != nil {
		return nil, errors.Wrap(listErr, "while listing tenants")
	}

	currentTenantsMap := make(map[string]string)
	for _, ct := range currentTenants {
		currentTenantsMap[ct.ExternalTenant] = ct.ID
	}

	return currentTenantsMap, nil
}

func shouldFullResync(lastFullResyncTimestamp string, fullResyncInterval time.Duration) (bool, error) {
	i, err := strconv.ParseInt(lastFullResyncTimestamp, 10, 64)
	if err != nil {
		return false, err
	}
	ts := time.Unix(i/1000, 0)
	return time.Now().After(ts.Add(fullResyncInterval)), nil
}

func convertTimeToUnixMilliSecondString(timestamp time.Time) string {
	return strconv.FormatInt(timestamp.UnixNano()/int64(time.Millisecond), 10)
}
