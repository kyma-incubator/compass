package tenantfetchersvc

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	// DefaultScenario is the name of the default scenario
	DefaultScenario = "DEFAULT"
	// TenantOnDemandProvider is the name of the business tenant mapping provider used when the tenant is not found in the events service
	TenantOnDemandProvider = "lazily-tenant-fetcher"
)

// TenantFieldMapping missing godoc
type TenantFieldMapping struct {
	TotalPagesField   string `envconfig:"TENANT_TOTAL_PAGES_FIELD"`
	TotalResultsField string `envconfig:"TENANT_TOTAL_RESULTS_FIELD"`
	EventsField       string `envconfig:"TENANT_EVENTS_FIELD"`

	NameField              string `envconfig:"MAPPING_FIELD_NAME"`
	IDField                string `envconfig:"MAPPING_FIELD_ID"`
	GlobalAccountGUIDField string `envconfig:"MAPPING_FIELD_GLOBAL_ACCOUNT_GUID" default:"globalAccountGUID"`
	SubaccountIDField      string `envconfig:"MAPPING_FIELD_SUBACCOUNT_ID"`
	SubaccountGUIDField    string `envconfig:"MAPPING_FIELD_SUBACCOUNT_GUID"`
	CustomerIDField        string `envconfig:"MAPPING_FIELD_CUSTOMER_ID"`
	SubdomainField         string `envconfig:"MAPPING_FIELD_SUBDOMAIN"`
	DetailsField           string `envconfig:"MAPPING_FIELD_DETAILS"`
	DiscriminatorField     string `envconfig:"MAPPING_FIELD_DISCRIMINATOR"`
	DiscriminatorValue     string `envconfig:"MAPPING_VALUE_DISCRIMINATOR"`

	RegionField     string `envconfig:"MAPPING_FIELD_REGION"`
	EntityIDField   string `envconfig:"MAPPING_FIELD_ENTITY_ID"`
	EntityTypeField string `envconfig:"MAPPING_FIELD_ENTITY_TYPE"`

	// This is not a value from the actual event but the key under which the GlobalAccountGUIDField will be stored to avoid collisions
	GlobalAccountKey string `envconfig:"default=gaID,GLOBAL_ACCOUNT_KEY"`
}

// MovedSubaccountsFieldMapping missing godoc
type MovedSubaccountsFieldMapping struct {
	LabelValue   string `envconfig:"MAPPING_FIELD_ID"`
	SourceTenant string `envconfig:"MOVED_SUBACCOUNT_SOURCE_TENANT_FIELD"`
	TargetTenant string `envconfig:"MOVED_SUBACCOUNT_TARGET_TENANT_FIELD"`
}

// QueryConfig contains the name of query parameters fields and default/start values
type QueryConfig struct {
	PageNumField    string `envconfig:"QUERY_PAGE_NUM_FIELD"`
	PageSizeField   string `envconfig:"QUERY_PAGE_SIZE_FIELD"`
	TimestampField  string `envconfig:"QUERY_TIMESTAMP_FIELD"`
	RegionField     string `envconfig:"QUERY_REGION_FIELD"`
	PageStartValue  string `envconfig:"QUERY_PAGE_START"`
	PageSizeValue   string `envconfig:"QUERY_PAGE_SIZE"`
	SubaccountField string `envconfig:"QUERY_ENTITY_FIELD"`
}

// PageConfig missing godoc
type PageConfig struct {
	TotalPagesField   string
	TotalResultsField string
	PageNumField      string
}

type RegionConfig struct {
	EventAPIRegionalConfig
	RegionalClient EventAPIClient
}

// TenantStorageService missing godoc
//go:generate mockery --name=TenantStorageService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantStorageService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	ListsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
}

// LabelRepo missing godoc
//go:generate mockery --name=LabelRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepo interface {
	GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error)
}

// EventAPIClient missing godoc
//go:generate mockery --name=EventAPIClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventAPIClient interface {
	FetchTenantEventsPage(eventsType EventsType, additionalQueryParams QueryParams) (TenantEventsResponse, error)
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	ListByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// TenantSyncService missing godoc
//go:generate mockery --name=TenantSyncService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantSyncService interface {
	SyncTenants() error
}

// LabelDefConverter missing godoc
//go:generate mockery --name=LabelDefConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelDefConverter interface {
	ToGraphQLInput(definition model.LabelDefinition) (graphql.LabelDefinitionInput, error)
}

const (
	// RetryAttempts Failed requests retry attempts number
	RetryAttempts          = 7
	retryDelayMilliseconds = 100
	// size of a tenant and parent tenants if not already existing
	chunkSizeForTenantOnDemand = 5
)

// SubaccountOnDemandService for an on-demand creation of a subaccount tenant
type SubaccountOnDemandService struct {
	queryConfig          QueryConfig
	fieldMapping         TenantFieldMapping
	eventAPIClient       EventAPIClient
	retryAttempts        uint
	toEventsPage         func([]byte) *eventsPage
	transact             persistence.Transactioner
	tenantStorageService TenantStorageService
	gqlClient            DirectorGraphQLClient
	providerName         string
	tenantConverter      TenantConverter
	tenantsRegions       map[string]RegionDetails
}

type TenantAggregationService struct {
	config   JobConfig
	transact persistence.Transactioner

	kubeClient      KubeClient
	eventAPIClients map[string]EventAPIClient
	gqlClient       DirectorGraphQLClient

	tenantConverter      TenantConverter
	tenantStorageService TenantStorageService
	toEventsPage         func([]byte) *eventsPage
}

// GlobalAccountService missing godoc
type GlobalAccountService struct {
	config   JobConfig
	transact persistence.Transactioner

	kubeClient      KubeClient
	universalClient EventAPIClient
	regionalConfigs []RegionConfig
	gqlClient       DirectorGraphQLClient

	tenantConverter      TenantConverter
	tenantStorageService TenantStorageService
	toEventsPage         func([]byte) *eventsPage
}

// SubaccountService missing godoc
type SubaccountService struct {
	config   JobConfig
	transact persistence.Transactioner

	kubeClient      KubeClient
	universalClient EventAPIClient
	regionalConfigs []RegionConfig
	gqlClient       DirectorGraphQLClient

	tenantConverter       TenantConverter
	tenantStorageService  TenantStorageService
	runtimeStorageService RuntimeService
	labelRepo             LabelRepo

	toEventsPage func([]byte) *eventsPage
}

// NewSubaccountOnDemandService missing godoc
func NewSubaccountOnDemandService(
	queryConfig QueryConfig,
	fieldMapping TenantFieldMapping,
	client EventAPIClient,
	transact persistence.Transactioner,
	tenantStorageService TenantStorageService,
	gqlClient DirectorGraphQLClient,
	providerName string,
	tenantConverter TenantConverter,
	tenantsRegions map[string]RegionDetails) *SubaccountOnDemandService {
	return &SubaccountOnDemandService{
		queryConfig:    queryConfig,
		fieldMapping:   fieldMapping,
		eventAPIClient: client,
		retryAttempts:  RetryAttempts,
		toEventsPage: func(bytes []byte) *eventsPage {
			return &eventsPage{
				fieldMapping: fieldMapping,
				payload:      bytes,
				providerName: providerName,
			}
		},
		transact:             transact,
		tenantStorageService: tenantStorageService,
		gqlClient:            gqlClient,
		providerName:         providerName,
		tenantConverter:      tenantConverter,
		tenantsRegions:       tenantsRegions,
	}
}

// NewGlobalAccountService missing godoc
func NewGlobalAccountService(transact persistence.Transactioner, kubeClient KubeClient, fieldMapping TenantFieldMapping, providerName string, universalClient EventAPIClient, regionalConfigs []RegionConfig, tenantStorageService TenantStorageService, gqlClient DirectorGraphQLClient, tenantConverter TenantConverter, config JobConfig) *GlobalAccountService {
	return &GlobalAccountService{
		config:               config,
		transact:             transact,
		kubeClient:           kubeClient,
		universalClient:      universalClient,
		regionalConfigs:      regionalConfigs,
		gqlClient:            gqlClient,
		tenantConverter:      tenantConverter,
		tenantStorageService: tenantStorageService,
		toEventsPage: func(bytes []byte) *eventsPage {
			return &eventsPage{
				fieldMapping: fieldMapping,
				payload:      bytes,
				providerName: providerName,
			}
		},
	}
}

// NewSubaccountService missing godoc
func NewSubaccountService(
	transact persistence.Transactioner,
	kubeClient KubeClient,
	universalClient EventAPIClient,
	regionalConfigs []RegionConfig,
	tenantStorageService TenantStorageService,
	runtimeStorageService RuntimeService,
	labelRepo LabelRepo,
	gqlClient DirectorGraphQLClient,
	tenantConverter TenantConverter,
	config JobConfig) *SubaccountService {
	return &SubaccountService{
		transact:              transact,
		kubeClient:            kubeClient,
		universalClient:       universalClient,
		regionalConfigs:       regionalConfigs,
		tenantStorageService:  tenantStorageService,
		runtimeStorageService: runtimeStorageService,
		labelRepo:             labelRepo,
		toEventsPage: func(bytes []byte) *eventsPage {
			return &eventsPage{
				fieldMapping:                 config.TenantFieldMapping,
				movedSubaccountsFieldMapping: config.MovedSubaccountsFieldMapping,
				payload:                      bytes,
				providerName:                 config.TenantProvider,
			}
		},
		gqlClient:       gqlClient,
		tenantConverter: tenantConverter,
		config:          config,
	}
}

// SyncTenants missing godoc
func (s SubaccountService) SyncTenants() error {
	ctx := context.Background()
	startTime, lastConsumedTenantTimestamp, lastResyncTimestamp, err := resyncTimestamps(ctx, s.kubeClient, s.config.FullResyncInterval)
	if err != nil {
		return err
	}

	for _, regionalConfig := range s.regionalConfigs {
		region := regionalConfig.RegionName
		tenantsToCreate, err := s.getSubaccountsToCreateForRegion(lastConsumedTenantTimestamp, regionalConfig)
		if err != nil {
			return err
		}
		log.C(ctx).Printf("Got subaccount to create for region: %s", regionName)

		tenantsToDelete, err := s.getSubaccountsToDeleteForRegion(lastConsumedTenantTimestamp, regionalConfig)
		if err != nil {
			return err
		}
		log.C(ctx).Printf("Got subaccount to delete for region: %s", regionName)

		subaccountsToMove, err := s.getSubaccountsToMove(lastConsumedTenantTimestamp, regionName)
		if err != nil {
			return err
		}
		log.C(ctx).Printf("Got subaccount to move for region: %s", regionName)

		tenantsToCreate = dedupeTenants(tenantsToCreate)
		tenantsToCreate = excludeTenants(tenantsToCreate, tenantsToDelete)

		totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete) + len(subaccountsToMove)
		log.C(ctx).Printf("Amount of new events for region %s: %d", region, totalNewEvents)
		if totalNewEvents == 0 {
			continue
		}

		tx, err := s.transact.Begin()
		if err != nil {
			return err
		}
		defer s.transact.RollbackUnlessCommitted(ctx, tx)
		ctx = persistence.SaveToContext(ctx, tx)

		currentTenants := make(map[string]string)
		if len(tenantsToCreate) > 0 || len(tenantsToDelete) > 0 {
			currentTenantsIDs := getTenantsIDs(tenantsToCreate, tenantsToDelete)
			currentTenants, err = getCurrentTenants(ctx, s.tenantStorageService, currentTenantsIDs)
			if err != nil {
				return err
			}
		}

		// Order of event processing matters
		if len(tenantsToCreate) > 0 {
			fullRegionName := regionalConfig.RegionPrefix + regionalConfig.RegionName
			if err := createTenants(ctx, s.gqlClient, currentTenants, tenantsToCreate, fullRegionName, s.config.TenantProvider, s.config.TenantInsertChunkSize, s.tenantConverter); err != nil {
				return errors.Wrap(err, "while storing subaccounts")
			}
		}
		if len(subaccountsToMove) > 0 {
			if err := s.moveSubaccounts(ctx, subaccountsToMove); err != nil {
				return errors.Wrap(err, "while moving subaccounts")
			}
		}
		if len(tenantsToDelete) > 0 {
			if err := deleteTenants(ctx, s.gqlClient, currentTenants, tenantsToDelete, s.config.TenantInsertChunkSize, s.tenantConverter); err != nil {
				return errors.Wrap(err, "while deleting subaccounts")
			}
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		log.C(ctx).Printf("Processed all new events for region: %s", region)
	}

	return s.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixMilliSecondString(*startTime), lastResyncTimestamp)
}

// SyncTenant fetches creation events for a subaccount and creates a subaccount tenant in case it doesn't exist
func (s *SubaccountOnDemandService) SyncTenant(ctx context.Context, subaccountID string, parentID string) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if _, err = s.tenantStorageService.GetTenantByExternalID(ctx, subaccountID); err == nil {
		log.C(ctx).Infof("Subbaccount %s alredy exists in the database", subaccountID)
		if err := tx.Commit(); err != nil {
			log.C(ctx).Warnf("Failed to commit empty transaction: %v", err)
		}
		return nil
	} else if err != nil && !apperrors.IsNotFoundError(err) {
		return errors.Wrapf(err, "while fetching subaccount with ID %s from Director", subaccountID)
	}

	tenantToCreate, eventFound, err := s.getSubaccountToCreate(ctx, subaccountID, parentID)
	if err != nil {
		return err
	}

	parentTenantDetails, err := s.getParentDetailsForSubaccount(ctx, tenantToCreate, eventFound)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).Warnf("Failed to commit empty transaction: %v", err)
	}

	var tenantsToCreate = []model.BusinessTenantMappingInput{*tenantToCreate}
	fullRegionName := tenantToCreate.Region
	if detail, ok := s.tenantsRegions[fullRegionName]; ok {
		fullRegionName = detail.Prefix + detail.Name
	}

	if err := createTenants(ctx, s.gqlClient, parentTenantDetails, tenantsToCreate, fullRegionName, s.providerName, chunkSizeForTenantOnDemand, s.tenantConverter); err != nil {
		return errors.Wrapf(err, "while creating missing tenants from tenant hierarchy of subaccount tenant with ID %s", subaccountID)
	}

	log.C(ctx).Infof("Provided subaccount %s stored successfully with provider %s", subaccountID, tenantToCreate.Provider)
	return nil
}

// SyncTenants missing godoc
func (s GlobalAccountService) SyncTenants() error {
	ctx := context.Background()
	startTime, lastConsumedTenantTimestamp, lastResyncTimestamp, err := resyncTimestamps(ctx, s.kubeClient, s.config.FullResyncInterval)
	if err != nil {
		return err
	}

	for _, regionalConfig := range s.regionalConfigs {
		region := regionalConfig.RegionName
		tenantsToCreate, err := s.getAccountsToCreateForRegion(lastConsumedTenantTimestamp, regionalConfig)
		if err != nil {
			return err
		}
		log.C(ctx).Printf("Got accounts to create for region: %s", region)

		tenantsToDelete, err := s.getAccountsToDeleteForRegion(lastConsumedTenantTimestamp, regionalConfig)
		if err != nil {
			return err
		}
		log.C(ctx).Printf("Got accounts to delete for region: %s", region)

		tenantsToCreate = dedupeTenants(tenantsToCreate)
		tenantsToCreate = excludeTenants(tenantsToCreate, tenantsToDelete)

		totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete)
		log.C(ctx).Printf("Amount of new events for region %s: %d", region, totalNewEvents)
		if totalNewEvents == 0 {
			continue
		}

		tx, err := s.transact.Begin()
		if err != nil {
			return err
		}
		defer s.transact.RollbackUnlessCommitted(ctx, tx)
		ctx = persistence.SaveToContext(ctx, tx)

		currentTenants := make(map[string]string)
		if len(tenantsToCreate) > 0 || len(tenantsToDelete) > 0 {
			currentTenantsIDs := getTenantsIDs(tenantsToCreate, tenantsToDelete)
			currentTenants, err = getCurrentTenants(ctx, s.tenantStorageService, currentTenantsIDs)
			if err != nil {
				return err
			}
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		// Order of event processing matters
		if len(tenantsToCreate) > 0 {
			if err := createTenants(ctx, s.gqlClient, currentTenants, tenantsToCreate, region, s.config.TenantProvider, s.config.TenantInsertChunkSize, s.tenantConverter); err != nil {
				return errors.Wrapf(err, "while storing accounts from region %s", region)
			}
		}
		if len(tenantsToDelete) > 0 {
			if err := deleteTenants(ctx, s.gqlClient, currentTenants, tenantsToDelete, s.config.TenantInsertChunkSize, s.tenantConverter); err != nil {
				return errors.Wrapf(err, "moving deleting accounts from region %s", region)
			}
		}

		log.C(ctx).Printf("Processed new events for region: %s", region)
	}

	return s.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixMilliSecondString(*startTime), lastResyncTimestamp)
}

func resyncTimestamps(ctx context.Context, client KubeClient, fullResyncInterval time.Duration) (*time.Time, string, string, error) {
	startTime := time.Now()

	lastConsumedTenantTimestamp, lastResyncTimestamp, err := client.GetTenantFetcherConfigMapData(ctx)
	if err != nil {
		return nil, "", "", err
	}

	shouldFullResync, err := shouldFullResync(lastResyncTimestamp, fullResyncInterval)
	if err != nil {
		return nil, "", "", err
	}

	if shouldFullResync {
		log.C(ctx).Infof("Last full resync was %s ago. Will perform a full resync.", fullResyncInterval)
		lastConsumedTenantTimestamp = "1"
		lastResyncTimestamp = convertTimeToUnixMilliSecondString(startTime)
	}
	return &startTime, lastConsumedTenantTimestamp, lastResyncTimestamp, nil
}

func getTenantsIDs(allTenants ...[]model.BusinessTenantMappingInput) []string {
	var currentTenantsIDs []string
	for _, tenantsList := range allTenants {
		for _, tenant := range tenantsList {
			if len(tenant.Parent) > 0 {
				currentTenantsIDs = append(currentTenantsIDs, tenant.Parent)
			}
			if len(tenant.ExternalTenant) > 0 {
				currentTenantsIDs = append(currentTenantsIDs, tenant.ExternalTenant)
			}
		}
	}
	return currentTenantsIDs
}

func createTenants(ctx context.Context, gqlClient DirectorGraphQLClient, currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput, region string, provider string, maxChunkSize int, converter TenantConverter) error {
	tenantsToCreate := parents(currTenants, eventsTenants, provider)
	for _, eventTenant := range eventsTenants {
		if parentGUID, ok := currTenants[eventTenant.Parent]; ok {
			eventTenant.Parent = parentGUID
		}
		eventTenant.Region = region
		tenantsToCreate = append(tenantsToCreate, eventTenant)
	}

	tenantsToCreateGQL := converter.MultipleInputToGraphQLInput(tenantsToCreate)
	return executeInChunks(ctx, tenantsToCreateGQL, func(ctx context.Context, chunk []graphql.BusinessTenantMappingInput) error {
		return gqlClient.WriteTenants(ctx, chunk)
	}, maxChunkSize)
}

func executeInChunks(ctx context.Context, tenants []graphql.BusinessTenantMappingInput, f func(ctx context.Context, chunk []graphql.BusinessTenantMappingInput) error, maxChunkSize int) error {
	for {
		if len(tenants) == 0 {
			return nil
		}
		chunkSize := int(math.Min(float64(len(tenants)), float64(maxChunkSize)))
		tenantsChunk := tenants[:chunkSize]
		if err := f(ctx, tenantsChunk); err != nil {
			return err
		}
		tenants = tenants[chunkSize:]
	}
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

func (s *SubaccountOnDemandService) getParentDetailsForSubaccount(ctx context.Context, subaccount *model.BusinessTenantMappingInput, eventFound bool) (map[string]string, error) {
	parentTenantDetails := make(map[string]string)
	if !eventFound { // parentID is an existing internal tenant ID of a GA and will be assigned as a parent
		parentTenantDetails[subaccount.Parent] = subaccount.Parent
		return parentTenantDetails, nil
	}

	parent, err := s.tenantStorageService.GetTenantByExternalID(ctx, subaccount.Parent)
	if err == nil {
		parentTenantDetails[parent.ExternalTenant] = parent.ID
		return parentTenantDetails, nil
	} else if err != nil && apperrors.IsNotFoundError(err) {
		return parentTenantDetails, nil
	}
	return nil, err
}

func getTenantParentType(tenantType string) string {
	if tenantType == tenant.TypeToStr(tenant.Account) {
		return tenant.TypeToStr(tenant.Customer)
	}
	return tenant.TypeToStr(tenant.Account)
}

func (s SubaccountService) moveSubaccounts(ctx context.Context, movedSubaccountMappings []model.MovedSubaccountMappingInput) error {
	for _, mapping := range movedSubaccountMappings {
		if _, err := s.moveSubaccount(ctx, mapping); err != nil {
			return errors.Wrapf(err, "while moving subaccount with ID %s", mapping.SubaccountID)
		}
	}

	return nil
}

func (s SubaccountService) checkForScenarios(ctx context.Context, subaccountInternalID, sourceGATenant string) error {
	ctxWithSubaccount := tnt.SaveToContext(ctx, subaccountInternalID, "")
	runtimes, err := s.runtimeStorageService.ListByFilters(ctxWithSubaccount, nil)
	if err != nil {
		return errors.Wrapf(err, "while getting runtimes in subaccount %s", subaccountInternalID)
	}

	if len(runtimes) == 0 {
		return nil
	}

	sourceGA, err := s.tenantStorageService.GetTenantByExternalID(ctx, sourceGATenant)
	if err != nil {
		return errors.Wrapf(err, "while getting GA with externalID %s", sourceGATenant)
	}

	runtimeIDs := make([]string, 0, len(runtimes))
	for _, rt := range runtimes {
		runtimeIDs = append(runtimeIDs, rt.ID)
	}

	scenariosLabels, err := s.labelRepo.GetScenarioLabelsForRuntimes(ctx, sourceGA.ID, runtimeIDs)
	if err != nil {
		return errors.Wrapf(err, "while getting scenario labels for runtimes with ids [%s]", strings.Join(runtimeIDs, ","))
	}
	for _, scenariosLabel := range scenariosLabels {
		scenarios, err := label.ValueToStringsSlice(scenariosLabel.Value)
		if err != nil {
			return err
		}
		for _, scenario := range scenarios {
			if scenario != DefaultScenario {
				return errors.Errorf("could not move subaccount %s: runtime %s is in scenario %s in the source GA %s", subaccountInternalID, scenariosLabel.ObjectID, scenario, sourceGA.ID)
			}
		}
	}
	return nil
}

func deleteTenants(ctx context.Context, gqlClient DirectorGraphQLClient, currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput, maxChunkSize int, converter TenantConverter) error {
	tenantsToDelete := make([]model.BusinessTenantMappingInput, 0)
	for _, toDelete := range eventsTenants {
		if _, ok := currTenants[toDelete.ExternalTenant]; ok {
			tenantsToDelete = append(tenantsToDelete, toDelete)
		}
	}

	tenantsToDeleteGQL := converter.MultipleInputToGraphQLInput(tenantsToDelete)
	return executeInChunks(ctx, tenantsToDeleteGQL, func(ctx context.Context, chunk []graphql.BusinessTenantMappingInput) error {
		return gqlClient.DeleteTenants(ctx, chunk)
	}, maxChunkSize)
}

func (s GlobalAccountService) getAccountsToCreateForRegion(fromTimestamp string, config RegionConfig) ([]model.BusinessTenantMappingInput, error) {
	configProvider := eventsQueryConfigProvider(s.config, fromTimestamp)
	return fetchCreatedTenantsWithRetries(config.RegionalClient, s.config.RetryAttempts, s.supportedEventTypes, configProvider, s.toEventsPage)
}

func fetchCreatedTenantsWithRetries(eventAPIClient EventAPIClient, retryNumber uint, supportedEventsProvider func() supportedEvents, configProvider func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.BusinessTenantMappingInput, error) {
	supportedEvents := supportedEventsProvider()

	var fetchedTenants []model.BusinessTenantMappingInput

	createdTenants, err := fetchTenantsWithRetries(eventAPIClient, retryNumber, supportedEvents.createdTenantEvent, configProvider, toEventsPage)
	if err != nil {
		return nil, fmt.Errorf("while fetching created tenants: %v", err)
	}
	fetchedTenants = append(fetchedTenants, createdTenants...)

	updatedTenants, err := fetchTenantsWithRetries(eventAPIClient, retryNumber, supportedEvents.updatedTenantEvent, configProvider, toEventsPage)
	if err != nil {
		return nil, fmt.Errorf("while fetching updated tenants: %v", err)
	}

	fetchedTenants = append(fetchedTenants, updatedTenants...)
	return fetchedTenants, nil
}

func (s SubaccountService) getSubaccountsToCreateForRegion(fromTimestamp string, regionalConfig RegionConfig) ([]model.BusinessTenantMappingInput, error) {
	var fetchedTenants []model.BusinessTenantMappingInput

	if regionalConfig.UniversalClientEnabled {
		var err error
		configProvider := eventsQueryConfigProviderWithRegion(s.config, fromTimestamp, regionalConfig.RegionName)
		fetchedTenants, err = fetchCreatedTenantsWithRetries(s.universalClient, s.config.RetryAttempts, s.supportedEventTypes, configProvider, s.toEventsPage)
		if err != nil {
			return nil, err
		}
	}

	configProvider := eventsQueryConfigProvider(s.config, fromTimestamp)
	createdRegionalTenants, err := fetchCreatedTenantsWithRetries(regionalConfig.RegionalClient, s.config.RetryAttempts, s.supportedEventTypes, configProvider, s.toEventsPage)
	if err != nil {
		return nil, fmt.Errorf("while fetching created subaccounts: %v", err)
	}
	createdRegionalTenants = excludeTenants(createdRegionalTenants, fetchedTenants)
	fetchedTenants = append(fetchedTenants, createdRegionalTenants...)

	return fetchedTenants, nil
}

func (s SubaccountService) supportedEventTypes() supportedEvents {
	return supportedEvents{
		createdTenantEvent: CreatedSubaccountType,
		updatedTenantEvent: UpdatedSubaccountType,
		deletedTenantEvent: DeletedSubaccountType,
	}
}

func (s GlobalAccountService) supportedEventTypes() supportedEvents {
	return supportedEvents{
		createdTenantEvent: CreatedAccountType,
		updatedTenantEvent: UpdatedAccountType,
		deletedTenantEvent: DeletedAccountType,
	}
}

type supportedEvents struct {
	createdTenantEvent EventsType
	updatedTenantEvent EventsType
	deletedTenantEvent EventsType
}

func (s SubaccountOnDemandService) getSubaccountToCreate(ctx context.Context, subaccountID string, parentID string) (*model.BusinessTenantMappingInput, bool, error) {
	configProvider := func() (QueryParams, PageConfig) {
		return QueryParams{
				s.queryConfig.PageNumField:    s.queryConfig.PageStartValue,
				s.queryConfig.PageSizeField:   s.queryConfig.PageSizeValue,
				s.queryConfig.SubaccountField: subaccountID,
			}, PageConfig{
				TotalPagesField:   s.fieldMapping.TotalPagesField,
				TotalResultsField: s.fieldMapping.TotalResultsField,
				PageNumField:      s.queryConfig.PageNumField,
			}
	}
	fetchedTenants, err := fetchTenantsWithRetries(s.eventAPIClient, s.retryAttempts, CreatedSubaccountType, configProvider, s.toEventsPage)
	if err != nil {
		return nil, false, fmt.Errorf("while fetching subaccount by ID: %v", err)
	}

	if len(fetchedTenants) < 1 {
		log.C(ctx).Errorf("No create events for subaccount with ID %s were found", subaccountID)
		return &model.BusinessTenantMappingInput{
			Name:           subaccountID,
			ExternalTenant: subaccountID,
			Parent:         parentID,
			Type:           string(tenant.Subaccount),
			Provider:       TenantOnDemandProvider,
		}, false, nil
	}

	if len(fetchedTenants) > 1 {
		return nil, true, fmt.Errorf("expected one create event for tenant with ID %s, found %d", subaccountID, len(fetchedTenants))
	}

	return &fetchedTenants[0], true, nil
}

func (s GlobalAccountService) getAccountsToDeleteForRegion(fromTimestamp string, config RegionConfig) ([]model.BusinessTenantMappingInput, error) {
	configProvider := eventsQueryConfigProvider(s.config, fromTimestamp)
	return fetchTenantsWithRetries(config.RegionalClient, s.config.RetryAttempts, DeletedAccountType, configProvider, s.toEventsPage)
}

func (s SubaccountService) getSubaccountsToDeleteForRegion(fromTimestamp string, config RegionConfig) ([]model.BusinessTenantMappingInput, error) {
	var fetchedTenants []model.BusinessTenantMappingInput

	if config.UniversalClientEnabled {
		var err error
		configProvider := eventsQueryConfigProviderWithRegion(s.config, fromTimestamp, config.RegionName)
		fetchedTenants, err = fetchTenantsWithRetries(s.universalClient, s.config.RetryAttempts, DeletedSubaccountType, configProvider, s.toEventsPage)
		if err != nil {
			return nil, err
		}
	}

	configProvider := eventsQueryConfigProvider(s.config, fromTimestamp)
	fetchedRegionalTenants, err := fetchTenantsWithRetries(config.RegionalClient, s.config.RetryAttempts, DeletedSubaccountType, configProvider, s.toEventsPage)
	if err != nil {
		return nil, fmt.Errorf("while fetching created subaccounts: %v", err)
	}
	fetchedRegionalTenants = excludeTenants(fetchedRegionalTenants, fetchedTenants)
	fetchedTenants = append(fetchedTenants, fetchedRegionalTenants...)

	return fetchedTenants, nil
}

func (s SubaccountService) getSubaccountsToMove(fromTimestamp string, region string) ([]model.MovedSubaccountMappingInput, error) {
	configProvider := eventsQueryConfigProviderWithRegion(s.config, fromTimestamp, region)
	return fetchMovedSubaccountsWithRetries(s.universalClient, s.config.RetryAttempts, configProvider, s.toEventsPage)
}

func fetchTenantsWithRetries(eventAPIClient EventAPIClient, retryNumber uint, eventsType EventsType, configProvider func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.BusinessTenantMappingInput, error) {
	var tenants []model.BusinessTenantMappingInput
	err := fetchWithRetries(retryNumber, func() error {
		fetchedTenants, err := fetchTenants(eventAPIClient, eventsType, configProvider, toEventsPage)
		if err != nil {
			return fmt.Errorf("while fetching tenants: %v", err)
		}
		tenants = fetchedTenants
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tenants, nil
}

func fetchMovedSubaccountsWithRetries(eventAPIClient EventAPIClient, retryNumber uint, configProvider func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.MovedSubaccountMappingInput, error) {
	var tenants []model.MovedSubaccountMappingInput
	err := fetchWithRetries(retryNumber, func() error {
		fetchedTenants, err := fetchMovedSubaccounts(eventAPIClient, configProvider, toEventsPage)
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

func fetchTenants(eventAPIClient EventAPIClient, eventsType EventsType, configProvider func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.BusinessTenantMappingInput, error) {
	tenants := make([]model.BusinessTenantMappingInput, 0)
	err := walkThroughPages(eventAPIClient, eventsType, configProvider, toEventsPage, func(page *eventsPage) error {
		mappings := page.getTenantMappings(eventsType)
		tenants = append(tenants, mappings...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("while walking through pages: %v", err)
	}

	return tenants, nil
}

func fetchMovedSubaccounts(eventAPIClient EventAPIClient, configProvider func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage) ([]model.MovedSubaccountMappingInput, error) {
	allMappings := make([]model.MovedSubaccountMappingInput, 0)

	err := walkThroughPages(eventAPIClient, MovedSubaccountType, configProvider, toEventsPage, func(page *eventsPage) error {
		mappings := page.getMovedSubaccounts()
		allMappings = append(allMappings, mappings...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allMappings, nil
}

func eventsQueryConfigProvider(config JobConfig, fromTimestamp string) func() (QueryParams, PageConfig) {
	return eventsQueryConfigProviderWithRegion(config, fromTimestamp, "")
}

func eventsQueryConfigProviderWithRegion(config JobConfig, fromTimestamp string, region string) func() (QueryParams, PageConfig) {
	return func() (QueryParams, PageConfig) {
		qp := QueryParams{
			config.QueryConfig.PageNumField:   config.QueryConfig.PageStartValue,
			config.QueryConfig.PageSizeField:  config.QueryConfig.PageSizeValue,
			config.QueryConfig.TimestampField: fromTimestamp,
		}
		if len(region) > 0 {
			qp[config.QueryConfig.RegionField] = region
		}

		pc := PageConfig{
			TotalPagesField:   config.TenantFieldMapping.TotalPagesField,
			TotalResultsField: config.TenantFieldMapping.TotalResultsField,
			PageNumField:      config.QueryConfig.PageNumField,
		}
		return qp, pc
	}
}
func walkThroughPages(eventAPIClient EventAPIClient, eventsType EventsType, configProvider func() (QueryParams, PageConfig), toEventsPage func([]byte) *eventsPage, applyFunc func(*eventsPage) error) error {
	params, pageConfig := configProvider()
	firstPage, err := eventAPIClient.FetchTenantEventsPage(eventsType, params)
	if err != nil {
		return errors.Wrap(err, "while fetching tenant events page")
	}
	if firstPage == nil {
		return nil
	}

	err = applyFunc(toEventsPage(firstPage))
	if err != nil {
		return fmt.Errorf("while applyfunc on event page: %v", err)
	}

	initialCount := gjson.GetBytes(firstPage, pageConfig.TotalResultsField).Int()
	totalPages := gjson.GetBytes(firstPage, pageConfig.TotalPagesField).Int()

	pageStart, err := strconv.ParseInt(params[pageConfig.PageNumField], 10, 64)
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
		// cleaning up parents of self referencing tenants
		if t.ExternalTenant == t.Parent {
			t.Parent = ""
		}

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

func getCurrentTenants(ctx context.Context, tenantStorage TenantStorageService, tenantsIDs []string) (map[string]string, error) {
	currentTenants, listErr := tenantStorage.ListsByExternalIDs(ctx, tenantsIDs)
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

func (s *SubaccountService) moveSubaccount(ctx context.Context, mapping model.MovedSubaccountMappingInput) (*model.BusinessTenantMapping, error) {
	targetInternalTenant, err := s.tenantStorageService.GetTenantByExternalID(ctx, mapping.TargetTenant)
	if err != nil && strings.Contains(err.Error(), apperrors.NotFoundMsg) {
		parentTenant := model.BusinessTenantMappingInput{
			Name:           mapping.TargetTenant,
			ExternalTenant: mapping.TargetTenant,
			Parent:         "", // crm ID is assumed that it can be empty
			Subdomain:      "", // it is not available when event is for moving a subaccount
			Region:         "",
			Type:           tenant.TypeToStr(tenant.Account),
			Provider:       s.config.TenantProvider,
		}
		tenantsToCreateGQL := s.tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{parentTenant})
		if err := s.gqlClient.WriteTenants(ctx, tenantsToCreateGQL); err != nil {
			return nil, err
		}

		targetInternalTenant, err = s.tenantStorageService.GetTenantByExternalID(ctx, mapping.TargetTenant)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting internal tenant for external tenant ID %s", mapping.TargetTenant)
		}
	} else if err != nil {
		return nil, errors.Wrapf(err, "while getting internal tenant for external tenant ID %s", mapping.TargetTenant)
	}

	subaccountID := mapping.SubaccountID
	subaccountTenant, err := s.tenantStorageService.GetTenantByExternalID(ctx, subaccountID)
	if err != nil && strings.Contains(err.Error(), apperrors.NotFoundMsg) {
		mapping.TenantMappingInput.Parent = targetInternalTenant.ID
		tenantsToCreateGQL := s.tenantConverter.MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput{mapping.TenantMappingInput})
		if err := s.gqlClient.WriteTenants(ctx, tenantsToCreateGQL); err != nil {
			return nil, err
		}
		return targetInternalTenant, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "while getting subaccount internal tenant ID for external tenant ID %s", subaccountID)
	}

	if subaccountTenant.Parent == targetInternalTenant.ID {
		log.C(ctx).Infof("Subaccount with external id %s is already moved in global account with external id %s", subaccountTenant.ExternalTenant, mapping.TargetTenant)
		return targetInternalTenant, nil
	}

	if err := s.checkForScenarios(ctx, subaccountTenant.ID, mapping.SourceTenant); err != nil {
		return nil, err
	}

	subaccountTenant.Parent = targetInternalTenant.ID
	subaccountTenantGQL := s.tenantConverter.ToGraphQLInput(subaccountTenant.ToInput())
	if err := s.gqlClient.UpdateTenant(ctx, subaccountTenant.ID, subaccountTenantGQL); err != nil {
		return nil, errors.Wrapf(err, "while updating tenant with id %s", subaccountTenant.ID)
	}
	return targetInternalTenant, nil
}
