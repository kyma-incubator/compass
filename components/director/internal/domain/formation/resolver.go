package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	CreateFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
}

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	FromGraphQL(i graphql.FormationInput) model.Formation
	ToGraphQL(i *model.Formation) *graphql.Formation
}

// Resolver is the formation resolver
type Resolver struct {
	transact persistence.Transactioner
	service  Service
	conv     Converter
	fetcher  tenant.FetcherOnDemand
}

// NewResolver creates formation resolver
func NewResolver(transact persistence.Transactioner, service Service, conv Converter, fetcher tenant.FetcherOnDemand) *Resolver {
	return &Resolver{
		transact: transact,
		service:  service,
		conv:     conv,
		fetcher:  fetcher,
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

	if objectType == graphql.FormationObjectTypeTenant {
		if err := r.fetcher.FetchOnDemand(objectID, tnt); err != nil {
			return nil, errors.Wrapf(err, "while trying to create if not exists subaccount %s", objectID)
		}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	newFormation, err := r.service.AssignFormation(ctx, tnt, objectID, objectType, r.conv.FromGraphQL(formation))
	if err != nil {
		return nil, err
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
