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
//
//go:generate mockery --name=BusinessTenantMappingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BusinessTenantMappingService interface {
	CreateTenantAccessForResource(ctx context.Context, tenantAccess *model.TenantAccess) error
	ListByParentAndType(ctx context.Context, parentID string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error)
	ListByIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
	GetCustomerIDParentRecursively(ctx context.Context, tenantID string) (string, error)
	GetParentsRecursivelyByExternalTenant(ctx context.Context, externalTenant string) ([]*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

// ApplicationService is responsible for the service-layer application operations.
//
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

func (d *directive) createTenantAccessForOrgApplications(ctx context.Context, newTenantParentIDs []string, receivingAccessTnt string) error {
	parentTenants, err := d.tenantService.ListByIDs(ctx, newTenantParentIDs)
	if err != nil {
		return err
	}

	var childType tenantpkg.Type
	var parent *model.BusinessTenantMapping
	for _, parentTenant := range parentTenants {
		if parentTenant.Type == tenantpkg.Customer {
			childType = tenantpkg.Organization
			parent = parentTenant
			break
		}
		if parentTenant.Type == tenantpkg.CostObject {
			childType = tenantpkg.Folder
			parent = parentTenant
			break
		}
	}
	if parent == nil {
		return apperrors.NewInternalError("Unexpected error. The parent tenant must be Customer or CostObject.")
	}

	childTenants, err := d.tenantService.ListByParentAndType(ctx, parent.ID, childType)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while listing tenants by parent with ID %s and %s: %v", parent.ID, childType, err)
		return apperrors.NewInternalError("An error occurred while listing tenants by parent with ID %s and %s: %v", parent.ID, childType, err)
	}

	for _, childTenant := range childTenants {
		ctx = tenant.SaveToContext(ctx, childTenant.ID, childTenant.ExternalTenant)
		tenantApps, err := d.appService.ListAll(ctx)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while listing applications for tenant %s: %v", childTenant.ID, err)
			return apperrors.NewInternalError("An error occurred while listing applications for tenant %s: %v", childTenant.ID, err)
		}

		for _, app := range tenantApps {
			if err := d.tenantService.CreateTenantAccessForResource(ctx, &model.TenantAccess{InternalTenantID: receivingAccessTnt, ResourceType: resource.Application, ResourceID: app.ID, Owner: false, Source: parent.ID}); err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while creating tenant access: %v", err)
				return apperrors.NewInternalError("An error occurred while creating tenant access: %v", err)
			}
		}
	}

	return nil
}

func (d *directive) createTenantAccessForNewApplication(ctx context.Context, tntFromContext *model.BusinessTenantMapping, appID string) error {
	var err error
	var parentType tenantpkg.Type
	var parentTntID string

	if tntFromContext.Type != tenantpkg.Customer && tntFromContext.Type != tenantpkg.CostObject {
		parents, err := d.tenantService.GetParentsRecursivelyByExternalTenant(ctx, tntFromContext.ExternalTenant)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting parents for tenant %s", tntFromContext.ID)
			return errors.Wrapf(err, "while getting parents for tenant: %s", tntFromContext.ID)
		}

		for _, parent := range parents {
			if parent.Type == tenantpkg.Customer || parent.Type == tenantpkg.CostObject {
				parentTntID = parent.ID
				parentType = parent.Type
				break
			}
		}
	}

	if parentTntID == "" {
		return nil
	}

	log.C(ctx).Debugf("Found parent: %s with type %s for tenant with ID %s", parentTntID, parentType, tntFromContext.ID)

	tenants, err := d.tenantService.ListByParentAndType(ctx, parentTntID, tenantpkg.Account)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while listing tenants by parent ID %s", parentTntID)
		return errors.Wrapf(err, "while listing tenants by parent %s and type %s", parentTntID, tenantpkg.Account)
	}

	for _, tnt := range tenants {
		if err := d.tenantService.CreateTenantAccessForResource(ctx, &model.TenantAccess{InternalTenantID: tnt.ID, ResourceType: resource.Application, ResourceID: appID, Owner: true, Source: parentTntID}); err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while creating tenant access for tenant %s and application %s", tnt.ID, appID)
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

	if tntModel.Type != tenantpkg.Account || len(tntModel.Parents) == 0 {
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

	return d.createTenantAccessForOrgApplications(ctx, tntModel.Parents, tntModel.ID)
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
