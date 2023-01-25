package applicationtenancy

import (
	"context"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

// BusinessTenantMappingService is responsible for the service-layer tenant operations.
//go:generate mockery --name=BusinessTenantMappingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BusinessTenantMappingService interface {
	CreateTenantAccessForResource(ctx context.Context, tenantID, resourceID string, isOwner bool, resourceType resource.Type) error
	ListByParentAndType(ctx context.Context, parentID string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error)
	GetCustomerIDParentRecursively(ctx context.Context, tenantID string) (string, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// ApplicationService is responsible for the service-layer application operations.
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	ListAll(ctx context.Context) ([]*model.Application, error)
}

type directive struct {
	transact      persistence.Transactioner
	tenantService BusinessTenantMappingService
	appService    ApplicationService
}

// NewDirective creates a directive object
func NewDirective(transact persistence.Transactioner, tenantService BusinessTenantMappingService, appService ApplicationService) *directive {
	return &directive{
		transact:      transact,
		tenantService: tenantService,
		appService:    appService,
	}
}

// SynchronizeApplicationTenancy handles graphql.EventTypeNewApplication, graphql.EventTypeNewSingleTenant, and graphql.EventTypeNewMultipleTenants events.
// In EventTypeNewApplication we extract the customer parent of the owner of the new application. We give access to accounts under that customer to access the application.
// In EventTypeNewSingleTenant we get the new tenant's customer parent. If the new tenant is account we give him access to all Atom applications under the customer.
// In EventTypeNewMultipleTenants we do the same as for EventTypeNewSingleTenant event but for multiple new tenants.
func (d *directive) SynchronizeApplicationTenancy(ctx context.Context, _ interface{}, next gqlgen.Resolver, eventType graphql.EventType) (res interface{}, err error) {
	resp, err := next(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing request: %s", err.Error())
		return resp, err
	}

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening database transaction")
		return nil, apperrors.NewInternalError("Unable to initialize database transaction: %s", err.Error())
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Debugf("Preparing tenant access creation for event: %s", eventType)

	if eventType == graphql.EventTypeNewApplication {
		if err := d.handleNewApplicationCreation(ctx, resp); err != nil {
			return nil, err
		}
	} else if eventType == graphql.EventTypeNewSingleTenant {
		if err := d.handleNewSingleTenantCreation(ctx, resp); err != nil {
			return nil, err
		}
	} else if eventType == graphql.EventTypeNewMultipleTenants {
		if err := d.handleNewMultipleTenantCreation(ctx, resp); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while closing database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to finalize request %v", err)
	}

	return resp, nil
}

func (d *directive) createTenantAccessForOrgApplications(ctx context.Context, newTenantParentID, receivingAccessTnt string) error {
	orgTenants, err := d.tenantService.ListByParentAndType(ctx, newTenantParentID, tenantpkg.Organization)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while listing tenants by parent with ID %s and %s: %v", newTenantParentID, tenantpkg.Organization, err)
		return apperrors.NewInternalError("An error occurred while listing tenants by parent with ID %s and %s: %v", newTenantParentID, tenantpkg.Organization, err)
	}

	for _, orgTenant := range orgTenants {
		ctx = tenant.SaveToContext(ctx, orgTenant.ID, orgTenant.ExternalTenant)
		tenantApps, err := d.appService.ListAll(ctx)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while listing applications for tenant %s: %v", orgTenant.ID, err)
			return apperrors.NewInternalError("An error occurred while listing applications for tenant %s: %v", orgTenant.ID, err)
		}

		for _, app := range tenantApps {
			if err := d.tenantService.CreateTenantAccessForResource(ctx, receivingAccessTnt, app.ID, false, resource.Application); err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while creating tenant access: %v", err)
				return apperrors.NewInternalError("An error occurred while creating tenant access: %v", err)
			}
		}
	}

	return nil
}

func (d *directive) createTenantAccessForNewApplication(ctx context.Context, tntFromContext *model.BusinessTenantMapping, appID string) error {
	var err error
	parentTntID := tntFromContext.ID

	if tntFromContext.Type != tenantpkg.Customer {
		parentExternalID, err := d.tenantService.GetCustomerIDParentRecursively(ctx, tntFromContext.ID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant %s customer parent", tntFromContext.ID)
			return errors.Wrapf(err, "while getting customer parent for tenant: %s", tntFromContext.ID)
		}

		parent, err := d.tenantService.GetTenantByExternalID(ctx, parentExternalID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting parent model by external ID %s", parentExternalID)
			return errors.Wrapf(err, "while getting parent: %s", tntFromContext.ID)
		}

		parentTntID = parent.ID
	}

	log.C(ctx).Debugf("Found parent: %s for tenant with ID %s", parentTntID, tntFromContext.ID)

	tenants, err := d.tenantService.ListByParentAndType(ctx, parentTntID, tenantpkg.Account)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while listing tenants by parent ID %s", parentTntID)
		return errors.Wrapf(err, "while listing tenants by parent %s and type %s", parentTntID, tenantpkg.Account)
	}

	for _, tenant := range tenants {
		if err := d.tenantService.CreateTenantAccessForResource(ctx, tenant.ID, appID, false, resource.Application); err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while creating tenant access for tenant %s and application %s", tenant.ID, appID)
			return errors.Wrap(err, "while creating tenant access")
		}
	}

	return nil
}

func (d *directive) getTenantAndValidateTenantAccessEligibility(ctx context.Context, tenantID string) (*model.BusinessTenantMapping, error) {
	tntModel, err := d.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant with ID %s", tenantID)
		return nil, err
	}

	if tntModel.Type != tenantpkg.Account || tntModel.Parent == "" {
		log.C(ctx).Debugf("Tenant with ID %s is not Account type or does not have a parent. Skipping tenant access creation.", tntModel.ID)
		return nil, nil
	}

	return tntModel, nil
}

func (d *directive) processSingleTenant(ctx context.Context, tenantID string) error {
	tntModel, err := d.getTenantAndValidateTenantAccessEligibility(ctx, tenantID)
	if err != nil {
		return err
	}

	if tntModel == nil {
		return nil
	}

	return d.createTenantAccessForOrgApplications(ctx, tntModel.Parent, tntModel.ID)
}

func (d *directive) handleNewSingleTenantCreation(ctx context.Context, resp interface{}) error {
	tenantID, ok := resp.(string)
	if !ok {
		log.C(ctx).Errorf("An error occurred while casting the graphql response entity to single tenant string")
		return apperrors.NewInvalidDataError("An error occurred while casting the response entity: %v", resp)
	}

	return d.processSingleTenant(ctx, tenantID)
}

func (d *directive) handleNewMultipleTenantCreation(ctx context.Context, resp interface{}) error {
	tenantIDs, ok := resp.([]string)
	if !ok {
		log.C(ctx).Errorf("An error occurred while casting the response entity to string array")
		return apperrors.NewInvalidDataError("An error occurred while casting the response entity: %v", resp)
	}

	for _, tenantID := range tenantIDs {
		if err := d.processSingleTenant(ctx, tenantID); err != nil {
			return err
		}
	}

	return nil
}

func (d *directive) handleNewApplicationCreation(ctx context.Context, resp interface{}) error {
	tntIDFromContext, err := tenant.LoadFromContext(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while loading tenant from context")
		return err
	}

	tntModel, err := d.tenantService.GetTenantByID(ctx, tntIDFromContext)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant with ID %s", tntIDFromContext)
		return errors.Wrapf(err, "while fetching tenant with id: %s", tntIDFromContext)
	}

	log.C(ctx).Debugf("Found a matching tenant in the database: %s", tntModel.ID)

	if !isAtomTenant(tntModel.Type) {
		log.C(ctx).Infof("Tenant type is %s. Will not continue with tanancy synchronization", tntModel.Type)
		return nil
	}

	entity, ok := resp.(graphql.Entity)
	if !ok {
		log.C(ctx).Errorf("An error occurred while casting the graphql response entity")
		return apperrors.NewInvalidDataError("An error occurred while casting the response entity: %v", resp)
	}

	return d.createTenantAccessForNewApplication(ctx, tntModel, entity.GetID())
}

func isAtomTenant(tenantType tenantpkg.Type) bool {
	if tenantType == tenantpkg.ResourceGroup || tenantType == tenantpkg.Folder || tenantType == tenantpkg.Organization {
		return true
	}

	return false
}
