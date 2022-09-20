package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	Get(ctx context.Context, id string) (*model.Formation, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationPage, error)
	CreateFormation(ctx context.Context, tnt string, formation model.Formation, templateName string) (*model.Formation, error)
	DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error)
	AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
	UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error)
}

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	FromGraphQL(i graphql.FormationInput) model.Formation
	ToGraphQL(i *model.Formation) *graphql.Formation
	MultipleToGraphQL(in []*model.Formation) []*graphql.Formation
}

// TenantFetcher calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//go:generate mockery --name=TenantFetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantFetcher interface {
	FetchOnDemand(tenant, parentTenant string) error
}

// Resolver is the formation resolver
type Resolver struct {
	transact persistence.Transactioner
	service  Service
	conv     Converter
	fetcher  TenantFetcher
}

// NewResolver creates formation resolver
func NewResolver(transact persistence.Transactioner, service Service, conv Converter, fetcher TenantFetcher) *Resolver {
	return &Resolver{
		transact: transact,
		service:  service,
		conv:     conv,
		fetcher:  fetcher,
	}
}

// Formation returns a Formation by its id
func (r *Resolver) Formation(ctx context.Context, id string) (*graphql.Formation, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formation, err := r.service.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(formation), nil
}

// Formations returns paginated Formations based on first and after
func (r *Resolver) Formations(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.FormationPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	formationPage, err := r.service.List(ctx, *first, cursor)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	formations := r.conv.MultipleToGraphQL(formationPage.Data)

	return &graphql.FormationPage{
		Data:       formations,
		TotalCount: formationPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(formationPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(formationPage.PageInfo.EndCursor),
			HasNextPage: formationPage.PageInfo.HasNextPage,
		},
	}, nil
}

// CreateFormation creates new formation for the caller tenant
func (r *Resolver) CreateFormation(ctx context.Context, formationInput graphql.FormationInput) (*graphql.Formation, error) {
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

	templateName := model.DefaultTemplateName
	if formationInput.TemplateName != nil && *formationInput.TemplateName != "" {
		templateName = *formationInput.TemplateName
	}

	newFormation, err := r.service.CreateFormation(ctx, tnt, r.conv.FromGraphQL(formationInput), templateName)
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
