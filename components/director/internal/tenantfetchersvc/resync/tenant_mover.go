package resync

import (
	"context"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	// DefaultScenario is the name of the default scenario
	DefaultScenario = "DEFAULT"
)

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	ListByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// LabelRepo missing godoc
//go:generate mockery --name=LabelRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepo interface {
	GetScenarioLabelsForRuntimes(ctx context.Context, tenantID string, runtimesIDs []string) ([]model.Label, error)
}

type tenantMover struct {
	externalTenantsManager

	transact              persistence.Transactioner
	tenantStorageService  TenantStorageService
	runtimeStorageService RuntimeService
	labelRepo             LabelRepo
}

func NewSubaccountsMover(jobConfig JobConfig, transact persistence.Transactioner, directorClient DirectorGraphQLClient, eventAPIClient EventAPIClient, tenantConverter TenantConverter, storageSvc TenantStorageService, runtimeSvc RuntimeService, labelRepo LabelRepo) TenantMover {
	return &tenantMover{
		externalTenantsManager: externalTenantsManager{
			gqlClient:       directorClient,
			eventAPIClient:  eventAPIClient,
			config:          jobConfig.EventsConfig,
			tenantConverter: tenantConverter,
			tenantProvider:  jobConfig.TenantProvider,
		},
		transact:              transact,
		tenantStorageService:  storageSvc,
		runtimeStorageService: runtimeSvc,
		labelRepo:             labelRepo,
	}
}

func (tmv *tenantMover) TenantsToMove(ctx context.Context, region, fromTimestamp string) ([]model.MovedSubaccountMappingInput, error) {
	configProvider := eventsQueryConfigProviderWithRegion(tmv.config, fromTimestamp, region)
	return fetchMovedSubaccountsWithRetries(tmv.eventAPIClient, tmv.config.RetryAttempts, configProvider)
}

func (tmv *tenantMover) MoveTenants(ctx context.Context, movedSubaccountMappings []model.MovedSubaccountMappingInput) error {
	tx, err := tmv.transact.Begin()
	if err != nil {
		return err
	}
	defer tmv.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	tenantsToUpdate, tenantsToCreate, err := tmv.tenantsToUpsert(ctx, movedSubaccountMappings)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Moving tenants from one parent tenant to another")
	for _, t := range tenantsToUpdate {
		if err := tmv.moveSubaccount(ctx, t); err != nil {
			return errors.Wrapf(err, "while moving subaccount with ID %s", t.ExternalTenant)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "while committing transaction")
	}

	if len(tenantsToCreate) > 0 {
		log.C(ctx).Infof("Creating non-existing tenants in the correct parent tenant")
		tenantsToCreateGQL := tmv.tenantConverter.MultipleInputToGraphQLInput(tenantsToCreate)
		if err := tmv.gqlClient.WriteTenants(ctx, tenantsToCreateGQL); err != nil {
			return errors.Wrap(err, "while creating missing tenants")
		}
	}

	return nil
}

func (tmv *tenantMover) moveSubaccount(ctx context.Context, subaccountTenant *model.BusinessTenantMapping) error {
	subaccountTenantGQL := tmv.tenantConverter.ToGraphQLInput(subaccountTenant.ToInput())
	if err := tmv.gqlClient.UpdateTenant(ctx, subaccountTenant.ID, subaccountTenantGQL); err != nil {
		return errors.Wrapf(err, "while updating tenant with id %s", subaccountTenant.ID)
	}
	log.C(ctx).Infof("Successfully moved subaccount tenant %s to new parent with ID %s", subaccountTenant.ID, subaccountTenant.Parent)
	return nil
}

func (tmv tenantMover) checkForScenarios(ctx context.Context, subaccountInternalID, sourceGATenant string) error {
	ctxWithSubaccount := tnt.SaveToContext(ctx, subaccountInternalID, "")
	runtimes, err := tmv.runtimeStorageService.ListByFilters(ctxWithSubaccount, nil)
	if err != nil {
		return errors.Wrapf(err, "while getting runtimes in subaccount %s", subaccountInternalID)
	}

	if len(runtimes) == 0 {
		return nil
	}

	sourceGA, err := tmv.tenantStorageService.GetTenantByExternalID(ctx, sourceGATenant)
	if err != nil {
		return errors.Wrapf(err, "while getting GA with externalID %s", sourceGATenant)
	}

	runtimeIDs := make([]string, 0, len(runtimes))
	for _, rt := range runtimes {
		runtimeIDs = append(runtimeIDs, rt.ID)
	}

	scenariosLabels, err := tmv.labelRepo.GetScenarioLabelsForRuntimes(ctx, sourceGA.ID, runtimeIDs)
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

func (tmv *tenantMover) tenantsToUpsert(ctx context.Context, mappings []model.MovedSubaccountMappingInput) ([]*model.BusinessTenantMapping, []model.BusinessTenantMappingInput, error) {
	tenantsToMove, subaccountIDs, parentTenants, err := tmv.tenantsWithExistingTargetParentTenants(ctx, mappings)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while filtering out tenants without pre-existing target parent tenant")
	}
	if len(tenantsToMove) == 0 {
		log.C(ctx).Infof("No tenants are available for moving")
		return nil, nil, nil
	}

	existingTenants, err := tmv.tenantStorageService.ListsByExternalIDs(ctx, subaccountIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while getting tenants from DB")
	}

	existingTenantsMap := tenantMappings(existingTenants)

	tenantsToCreate := make([]model.BusinessTenantMappingInput, 0)
	tenantsToUpdate := make([]*model.BusinessTenantMapping, 0)

	for _, mapping := range tenantsToMove {
		tenantFromDB, ok := existingTenantsMap[mapping.SubaccountID]
		if !ok {
			log.C(ctx).Infof("Subaccount with external id %s does not exist, will be created in the correct parent tenant", mapping.SubaccountID)
			mapping.TenantMappingInput.Parent = mapping.TargetTenant
			tenantsToCreate = append(tenantsToCreate, mapping.TenantMappingInput)
			continue
		}

		if tenantFromDB.Parent == parentTenants[mapping.TargetTenant].ID {
			log.C(ctx).Infof("Subaccount with external id %s is already moved in global account with external id %s", tenantFromDB.ExternalTenant, mapping.TargetTenant)
			continue
		}

		if err := tmv.checkForScenarios(ctx, tenantFromDB.ID, mapping.SourceTenant); err != nil {
			return nil, nil, errors.Wrapf(err, "subaccount with external id %s is part of a scenario and cannot be moved", mapping.SubaccountID)
		}

		tenantFromDB.Parent = mapping.TargetTenant
		tenantsToUpdate = append(tenantsToUpdate, tenantFromDB)
	}

	return tenantsToUpdate, tenantsToCreate, nil
}

func (tmv *tenantMover) tenantsWithExistingTargetParentTenants(ctx context.Context, mappings []model.MovedSubaccountMappingInput) ([]model.MovedSubaccountMappingInput, []string, map[string]*model.BusinessTenantMapping, error) {
	existingParentTenants, err := tmv.targetParentTenants(ctx, mappings)
	if err != nil {
		return nil, nil, nil, err
	}

	subaccountIDs := make([]string, 0)
	mappingsWithParentTenants := make([]model.MovedSubaccountMappingInput, 0)
	for _, mapping := range mappings {
		if _, ok := existingParentTenants[mapping.TargetTenant]; !ok {
			log.C(ctx).Errorf("Target parent tenant %s of moved subaccount tenant %s does not exist, skipping...", mapping.TargetTenant, mapping.SubaccountID)
			continue
		}

		subaccountIDs = append(subaccountIDs, mapping.SubaccountID)
		mappingsWithParentTenants = append(mappingsWithParentTenants, mapping)
	}

	return mappingsWithParentTenants, subaccountIDs, existingParentTenants, nil
}

func (tmv *tenantMover) targetParentTenants(ctx context.Context, mappings []model.MovedSubaccountMappingInput) (map[string]*model.BusinessTenantMapping, error) {
	parentTenantIDs := make([]string, 0)
	for _, mapping := range mappings {
		parentTenantIDs = append(parentTenantIDs, mapping.TargetTenant)
	}

	existingParents, err := tmv.tenantStorageService.ListsByExternalIDs(ctx, parentTenantIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while getting target parent tenant ID of moved subaccounts")
	}

	return tenantMappings(existingParents), nil
}
func tenantMappings(tenants []*model.BusinessTenantMapping) map[string]*model.BusinessTenantMapping {
	tenantsMap := make(map[string]*model.BusinessTenantMapping)
	for _, t := range tenants {
		tenantsMap[t.ExternalTenant] = t
	}
	return tenantsMap
}

func fetchMovedSubaccountsWithRetries(eventAPIClient EventAPIClient, retryAttempts uint, configProvider func() (QueryParams, PageConfig)) ([]model.MovedSubaccountMappingInput, error) {
	var tenants []model.MovedSubaccountMappingInput
	err := retry.Do(func() error {
		fetchedTenants, err := fetchMovedSubaccounts(eventAPIClient, configProvider)
		if err != nil {
			return err
		}
		tenants = fetchedTenants
		return nil
	}, retry.Attempts(retryAttempts), retry.Delay(retryDelayMilliseconds*time.Millisecond))
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching moved tenants after %d retries", retryAttempts)
	}

	return tenants, nil
}

func fetchMovedSubaccounts(eventAPIClient EventAPIClient, configProvider func() (QueryParams, PageConfig)) ([]model.MovedSubaccountMappingInput, error) {
	allMappings := make([]model.MovedSubaccountMappingInput, 0)

	err := walkThroughPages(eventAPIClient, MovedSubaccountType, configProvider, func(page *EventsPage) error {
		mappings := page.getMovedSubaccounts()
		allMappings = append(allMappings, mappings...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allMappings, nil
}

type noOpMover struct{}

func newNoOpsMover() *noOpMover {
	return &noOpMover{}
}

func (*noOpMover) TenantsToMove(context.Context, string, string) ([]model.MovedSubaccountMappingInput, error) {
	return []model.MovedSubaccountMappingInput{}, nil
}

func (*noOpMover) MoveTenants(context.Context, []model.MovedSubaccountMappingInput) error {
	return nil
}
