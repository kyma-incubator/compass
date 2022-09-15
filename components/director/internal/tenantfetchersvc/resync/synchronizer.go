package resync

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

const (
	retryDelayMilliseconds = 100
	// TenantOnDemandProvider is the name of the business tenant mapping provider used when the tenant is not found in the events service
	TenantOnDemandProvider = "lazily-tenant-fetcher"
)

// TenantStorageService missing godoc
//go:generate mockery --name=TenantStorageService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantStorageService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	ListsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
}

// TenantCreator takes care of retrieving tenants from external tenant registry and storing them in Director
//go:generate mockery --name=TenantCreator --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantCreator interface {
	FetchTenant(ctx context.Context, externalTenantID string) (*model.BusinessTenantMappingInput, error)
	TenantsToCreate(ctx context.Context, region, fromTimestamp string) ([]model.BusinessTenantMappingInput, error)
	CreateTenants(ctx context.Context, eventsTenants []model.BusinessTenantMappingInput) error
}

// TenantDeleter takes care of retrieving no longer used tenants from external tenant registry and delete them from Director
//go:generate mockery --name=TenantDeleter --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantDeleter interface {
	TenantsToDelete(ctx context.Context, region, fromTimestamp string) ([]model.BusinessTenantMappingInput, error)
	DeleteTenants(ctx context.Context, eventsTenants []model.BusinessTenantMappingInput) error
}

// TenantMover takes care of moving tenants from one parent tenant to another.
//go:generate mockery --name=TenantMover --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMover interface {
	TenantsToMove(ctx context.Context, region, fromTimestamp string) ([]model.MovedSubaccountMappingInput, error)
	MoveTenants(ctx context.Context, movedSubaccountMappings []model.MovedSubaccountMappingInput) error
}

// TenantsSynchronizer takes care of synchronizing tenants with external tenants registry.
// It creates, updates, deletes and moves tenants that were created, updated, deleted or moved in that external registry.
type TenantsSynchronizer struct {
	supportedRegions []string

	transact             persistence.Transactioner
	tenantStorageService TenantStorageService

	creator TenantCreator
	mover   TenantMover
	deleter TenantDeleter

	kubeClient KubeClient
	config     JobConfig

	metricsReporter MetricsPusher
}

// NewTenantSynchronizer returns a new tenants synchronizer.
func NewTenantSynchronizer(config JobConfig, transact persistence.Transactioner, tenantStorageService TenantStorageService, creator TenantCreator, mover TenantMover, deleter TenantDeleter, kubeClient KubeClient, metricsReporter MetricsPusher) *TenantsSynchronizer {
	return &TenantsSynchronizer{
		supportedRegions:     supportedRegions(config),
		transact:             transact,
		tenantStorageService: tenantStorageService,
		creator:              creator,
		mover:                mover,
		deleter:              deleter,
		kubeClient:           kubeClient,
		config:               config,
		metricsReporter:      metricsReporter,
	}
}

func supportedRegions(config JobConfig) []string {
	regionNames := make([]string, 0)
	for _, regionDetails := range config.RegionalAPIConfigs {
		regionNames = append(regionNames, regionDetails.RegionName)
	}
	if len(regionNames) == 0 {
		log.D().Infof("Job %s has only one central region: %s", config.JobName, config.APIConfig.RegionName)
		regionNames = append(regionNames, config.APIConfig.RegionName)
	}
	return regionNames
}

// Name returns the name set to the tenants synchronizer.
func (ts *TenantsSynchronizer) Name() string {
	return ts.config.JobName
}

// TenantType returns the tenant type the tenants synchronizer is responsible for.
func (ts *TenantsSynchronizer) TenantType() tenant.Type {
	return ts.config.TenantType
}

// ResyncInterval returns the interval that the synchronizer is supposed to make a regular tenants resync with the external tenants registry.
func (ts *TenantsSynchronizer) ResyncInterval() time.Duration {
	return ts.config.TenantFetcherJobIntervalMins
}

// Synchronize is responsible for synchronizing the tenants of the configured type in Compass and the configured external tenants registry.
// When a tenant is created in the external registry, it is also greated in Compass. Same applies for updated, deleted and moved tenants.
func (ts *TenantsSynchronizer) Synchronize(ctx context.Context) error {
	var err error
	if err = ts.synchronizeTenants(ctx); err != nil {
		ts.metricsReporter.ReportFailedSync(ctx, err)
	}
	return err
}

func (ts *TenantsSynchronizer) synchronizeTenants(ctx context.Context) error {
	startTime, lastConsumedTenantTimestamp, lastResyncTimestamp, err := resyncTimestamps(ctx, ts.kubeClient, ts.config.FullResyncInterval)
	if err != nil {
		return err
	}

	for _, region := range ts.supportedRegions {
		log.C(ctx).Printf("Processing new events for region: %s...", region)
		tenantsToCreate, err := ts.creator.TenantsToCreate(ctx, region, lastConsumedTenantTimestamp)
		if err != nil {
			return err
		}

		tenantsToMove, err := ts.mover.TenantsToMove(ctx, region, lastConsumedTenantTimestamp)
		if err != nil {
			return err
		}

		tenantsToDelete, err := ts.deleter.TenantsToDelete(ctx, region, lastConsumedTenantTimestamp)
		if err != nil {
			return err
		}

		tenantsToCreate = dedupeTenants(tenantsToCreate)
		tenantsToCreate = excludeTenants(tenantsToCreate, tenantsToDelete)

		totalNewEvents := len(tenantsToCreate) + len(tenantsToDelete) + len(tenantsToMove)
		log.C(ctx).Printf("Amount of new events for region %s: %d", region, totalNewEvents)
		if totalNewEvents == 0 {
			log.C(ctx).Printf("No new events for processing, resync completed for region %s", region)
			continue
		}

		currentTenants := make(map[string]string)
		if len(tenantsToCreate) > 0 || len(tenantsToDelete) > 0 {
			currentTenantsIDs := getTenantsIDs(tenantsToCreate, tenantsToDelete)
			currentTenants, err = ts.currentTenants(ctx, currentTenantsIDs)
			if err != nil {
				return err
			}
		}

		// Order of event processing matters - we want the most destructive operation to be last
		if len(tenantsToCreate) > 0 {
			if err := ts.createTenants(ctx, currentTenants, tenantsToCreate, region); err != nil {
				return errors.Wrap(err, "while creating tenants")
			}
		}

		if len(tenantsToMove) > 0 {
			if err := ts.mover.MoveTenants(ctx, tenantsToMove); err != nil {
				return errors.Wrap(err, "while moving tenants")
			}
		}

		if len(tenantsToDelete) > 0 {
			if err := ts.deleteTenants(ctx, currentTenants, tenantsToDelete); err != nil {
				return errors.Wrap(err, "while deleting tenants")
			}
		}

		log.C(ctx).Printf("Processed all new events for region: %s", region)
	}

	return ts.kubeClient.UpdateTenantFetcherConfigMapData(ctx, convertTimeToUnixMilliSecondString(*startTime), lastResyncTimestamp)
}

// SynchronizeTenant is responsible for updating the given tenant with the values available in the external registry,
// or creating it if it does not exist in Compass.
// All available regions are checked for the existence of the tenant.
func (ts *TenantsSynchronizer) SynchronizeTenant(ctx context.Context, parentTenantID, tenantID string) error {
	tnt, err := ts.fetchFromDirector(ctx, tenantID)
	if err != nil {
		return errors.Wrapf(err, "while checking if tenant eith ID %s already exists", tenantID)
	}

	if tnt != nil {
		log.C(ctx).Infof("Tenant with external ID %s already exists", tenantID)
		return nil
	}

	fetchedTenant, err := ts.creator.FetchTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	if fetchedTenant == nil {
		log.C(ctx).Infof("Tenant with ID %s was not found, it will be stored lazily", tenantID)
		fetchedTenant := model.BusinessTenantMappingInput{
			Name:           tenantID,
			ExternalTenant: tenantID,
			Parent:         parentTenantID,
			Subdomain:      "",
			Region:         "",
			Type:           string(tenant.Subaccount),
			Provider:       TenantOnDemandProvider,
		}
		return ts.creator.CreateTenants(ctx, []model.BusinessTenantMappingInput{fetchedTenant})
	}

	parentTenantID = fetchedTenant.Parent
	if len(parentTenantID) == 0 {
		return fmt.Errorf("parent tenant not found of tenant with ID %s", tenantID)
	}

	parent, err := ts.fetchFromDirector(ctx, fetchedTenant.Parent)
	if err != nil {
		return errors.Wrapf(err, "while checking if parent tenant with ID %s exists", fetchedTenant.Parent)
	}

	fetchedTenant.Parent = parent.ID
	return ts.creator.CreateTenants(ctx, []model.BusinessTenantMappingInput{*fetchedTenant})
}

func (ts *TenantsSynchronizer) fetchFromDirector(ctx context.Context, tenantID string) (*model.BusinessTenantMapping, error) {
	tx, err := ts.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ts.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Checking if tenant with external tenant ID %s already exists...", tenantID)
	tnt, err := ts.tenantStorageService.GetTenantByExternalID(ctx, tenantID)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "while checking if tenant with external ID %s already exists", tenantID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tnt, nil
}

func (ts *TenantsSynchronizer) currentTenants(ctx context.Context, tenantsIDs []string) (map[string]string, error) {
	tx, err := ts.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ts.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	currentTenants, listErr := ts.tenantStorageService.ListsByExternalIDs(ctx, tenantsIDs)
	if listErr != nil {
		return nil, errors.Wrap(listErr, "while listing tenants")
	}

	currentTenantsMap := make(map[string]string)
	for _, ct := range currentTenants {
		currentTenantsMap[ct.ExternalTenant] = ct.ID
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return currentTenantsMap, nil
}

func (ts *TenantsSynchronizer) createTenants(ctx context.Context, currentTenants map[string]string, newTenants []model.BusinessTenantMappingInput, region string) error {
	fullRegionName := ts.config.RegionPrefix + region
	// create missing parent tenants
	tenantsToCreate := missingParentTenants(currentTenants, newTenants, ts.config.TenantProvider, fullRegionName)
	for _, eventTenant := range newTenants {
		// use internal ID of parent for pre-existing targetParentTenants
		if parentGUID, ok := currentTenants[eventTenant.Parent]; ok {
			eventTenant.Parent = parentGUID
		}

		eventTenant.Region = fullRegionName
		tenantsToCreate = append(tenantsToCreate, eventTenant)
	}

	return ts.creator.CreateTenants(ctx, tenantsToCreate)
}

func (ts *TenantsSynchronizer) deleteTenants(ctx context.Context, currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput) error {
	tenantsToDelete := make([]model.BusinessTenantMappingInput, 0)
	for _, toDelete := range eventsTenants {
		if _, ok := currTenants[toDelete.ExternalTenant]; ok {
			tenantsToDelete = append(tenantsToDelete, toDelete)
		}
	}

	if len(tenantsToDelete) > 0 {
		return ts.deleter.DeleteTenants(ctx, tenantsToDelete)
	}

	return nil
}

func missingParentTenants(currTenants map[string]string, eventsTenants []model.BusinessTenantMappingInput, providerName, region string) []model.BusinessTenantMappingInput {
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
					Region:         region,
				}
				parentsToCreate = append(parentsToCreate, parentTenant)
			}
		}
	}

	return dedupeTenants(parentsToCreate)
}

func getTenantParentType(tenantType string) string {
	if tenantType == tenant.TypeToStr(tenant.Account) {
		return tenant.TypeToStr(tenant.Customer)
	}
	return tenant.TypeToStr(tenant.Account)
}

func dedupeTenants(tenants []model.BusinessTenantMappingInput) []model.BusinessTenantMappingInput {
	tenantsByExtID := make(map[string]model.BusinessTenantMappingInput)
	for _, t := range tenants {
		tenantsByExtID[t.ExternalTenant] = t
	}

	tenants = make([]model.BusinessTenantMappingInput, 0, len(tenantsByExtID))
	for _, t := range tenantsByExtID {
		// cleaning up missingParentTenants of self referencing tenants
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

func getTenantsIDs(tenants ...[]model.BusinessTenantMappingInput) []string {
	var currentTenantsIDs []string
	for _, tenantsList := range tenants {
		for _, t := range tenantsList {
			if len(t.Parent) > 0 {
				currentTenantsIDs = append(currentTenantsIDs, t.Parent)
			}
			if len(t.ExternalTenant) > 0 {
				currentTenantsIDs = append(currentTenantsIDs, t.ExternalTenant)
			}
		}
	}
	return currentTenantsIDs
}
