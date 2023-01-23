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

func (d *directive) SynchronizeApplicationTenancy(ctx context.Context, _ interface{}, next gqlgen.Resolver, eventType graphql.EventType) (res interface{}, err error) {
	resp, err := next(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while processing request: %s", err.Error())
		return resp, err
	}

	tx, err := d.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening database transaction: %s", err.Error())
		return nil, apperrors.NewInternalError("Unable to initialize database transaction: %s", err.Error())
	}
	defer d.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if eventType == graphql.EventTypeNewApplication {
		tntIDFromContext, err := tenant.LoadFromContext(ctx)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while loading tenant from context: %v", err)
			return nil, err
		}

		tntModel, err := d.tenantService.GetTenantByID(ctx, tntIDFromContext)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant: %v", err)
			return nil, errors.Wrapf(err, "while fetching tenant with id: %s", tntIDFromContext)
		}

		if !(tntModel.Type == tenantpkg.ResourceGroup || tntModel.Type == tenantpkg.Folder || tntModel.Type == tenantpkg.Organization) {
			log.C(ctx).Infof("Tenant type is %s. Will not continue with tanancy synchronization", tntModel.Type)
			return resp, nil
		}

		entity, ok := resp.(graphql.Entity)
		if !ok {
			log.C(ctx).WithError(err).Errorf("An error occurred while casting the response entity: %v", resp)
			return nil, apperrors.NewInvalidDataError("An error occurred while casting the response entity: %v", resp)
		}

		if err := d.createTenantAccessForNewApplication(ctx, tntModel, entity.GetID()); err != nil {
			return nil, err
		}
	} else if eventType == graphql.EventTypeNewSingleTenant {
		tenantID, ok := resp.(string)
		if !ok {
			log.C(ctx).WithError(err).Errorf("An error occurred while casting the response entity: %v", err)
			return nil, apperrors.NewInvalidDataError("An error occurred while casting the response entity: %v", resp)
		}

		tntModel, err := d.tenantService.GetTenantByID(ctx, tenantID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant: %v", err)
			return nil, err
		}

		if tntModel.Type != tenantpkg.Account || tntModel.Parent == "" {
			return resp, nil
		}

		if err := d.createTenantAccessForOrgApplications(ctx, tntModel.Parent, tntModel.ID); err != nil {
			return nil, err
		}
	} else if eventType == graphql.EventTypeNewMultipleTenants {
		tenantIDs, ok := resp.([]string)
		if !ok {
			log.C(ctx).WithError(err).Errorf("An error occurred while casting the response entity: %v", err)
			return nil, apperrors.NewInvalidDataError("An error occurred while casting the response entity: %v", resp)
		}

		filteredTenants := make([]*model.BusinessTenantMapping, 0)

		for _, tenantID := range tenantIDs {
			t, err := d.tenantService.GetTenantByID(ctx, tenantID)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant: %v", err)
				return nil, err
			}

			if t.Type != tenantpkg.Account || t.Parent == "" {
				continue
			}

			filteredTenants = append(filteredTenants, t)
		}

		for _, t := range filteredTenants {
			if err := d.createTenantAccessForOrgApplications(ctx, t.Parent, t.ID); err != nil {
				return nil, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
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
			log.C(ctx).WithError(err).Errorf("An error occurred while getting tenant customer parent: %v", err)
			return errors.Wrapf(err, "while getting customer parent for tenant: %s", tntFromContext.ID)
		}

		parent, err := d.tenantService.GetTenantByExternalID(ctx, parentExternalID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while getting parent: %v", err)
			return errors.Wrapf(err, "while getting parent: %s", tntFromContext.ID)
		}

		parentTntID = parent.ID
	}

	log.C(ctx).Infof("Found parent: %s for tenant with ID %s", parentTntID, tntFromContext.ID)
	tenants, err := d.tenantService.ListByParentAndType(ctx, parentTntID, tenantpkg.Account)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while listing tenants by parent: %v", err)
		return errors.Wrapf(err, "while listing tenants by parent %s and type %s", parentTntID, tenantpkg.Account)
	}

	for _, tenant := range tenants {
		if err := d.tenantService.CreateTenantAccessForResource(ctx, tenant.ID, appID, false, resource.Application); err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while creating tenant access: %v", err)
			return errors.Wrap(err, "while creating tenant access")
		}
	}

	return nil
}
