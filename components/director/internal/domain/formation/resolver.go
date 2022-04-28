package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore
type Service interface {
	CreateFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
}

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	FromGraphQL(i graphql.FormationInput) model.Formation
	ToGraphQL(i *model.Formation) *graphql.Formation
}

// TenantFetcherService calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//go:generate mockery --name=TenantFetcherService --output=automock --outpkg=automock --case=underscore
type TenantFetcherService interface {
	FetchOnDemand(tenant string) error
}

// Resolver is the formation resolver
type Resolver struct {
	transact      persistence.Transactioner
	service       Service
	conv          Converter
	tenantSvc     tenantService
	tenantFetcher TenantFetcherService
}

// NewResolver creates formation resolver
func NewResolver(transact persistence.Transactioner, service Service, conv Converter, tenantSvc tenantService, fetcher TenantFetcherService) *Resolver {
	return &Resolver{
		transact:      transact,
		service:       service,
		conv:          conv,
		tenantSvc:     tenantSvc,
		tenantFetcher: fetcher,
	}
}

// CreateFormation creates new formation for the caller tenant
func (r *Resolver) CreateFormation(ctx context.Context, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	newFormation, err := r.service.CreateFormation(ctx, tnt, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation), nil
}

// DeleteFormation deletes the formation from the caller tenant formations
func (r *Resolver) DeleteFormation(ctx context.Context, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	deletedFormation, err := r.service.DeleteFormation(ctx, tnt, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(deletedFormation), nil
}

// AssignFormation assigns object to the provided formation
func (r *Resolver) AssignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var newFormation *model.Formation
	if objectType == graphql.FormationObjectTypeTenant {
		newFormation, err = r.assignFormationForTenant(ctx, &tx, tnt, objectID, formation)
		if err != nil {
			return nil, err
		}
	} else {
		newFormation, err = r.service.AssignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation))
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation), nil
}

// UnassignFormation unassigns the object from the provided formation
func (r *Resolver) UnassignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput) (*graphql.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	newFormation, err := r.service.UnassignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing transaction")
	}

	return r.conv.ToGraphQL(newFormation), nil
}

func (r *Resolver) assignFormationForTenant(ctx context.Context, tx *persistence.PersistenceTx, tenantFromContext string, objectID string, formationInput graphql.FormationInput) (*model.Formation, error) {
	tenantFromDB, err := r.tenantSvc.GetTenantByExternalID(ctx, objectID)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "while getting tenant %s", objectID)
	} else if err != nil {
		if err = (*tx).Commit(); err != nil { // close the current transaction before the HTTP call
			return nil, errors.Wrap(err, "while committing transaction")
		}
		if err := r.tenantFetcher.FetchOnDemand(objectID); err != nil {
			return nil, errors.Wrapf(err, "while trying to create if not exists subaccount %s", objectID)
		}
		persistenceTx, err := r.transact.Begin()
		if err != nil {
			return nil, err
		}
		tx = &persistenceTx
		ctx = persistence.SaveToContext(ctx, *tx)
		tenantFromDB, err = r.tenantSvc.GetTenantByExternalID(ctx, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting tenant %s", objectID)
		}
	}
	newFormation, err := r.service.AssignFormation(ctx, tenantFromContext, tenantFromDB.ID, graphql.FormationObjectTypeTenant, r.conv.FromGraphQL(formationInput))
	if err != nil {
		return nil, err
	}
	return newFormation, nil
}
