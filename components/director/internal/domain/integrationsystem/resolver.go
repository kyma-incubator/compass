package integrationsystem

import (
	"context"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// IntegrationSystemService missing godoc
//go:generate mockery --name=IntegrationSystemService --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationSystemService interface {
	Create(ctx context.Context, in model.IntegrationSystemInput) (string, error)
	Get(ctx context.Context, id string) (*model.IntegrationSystem, error)
	List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error)
	Update(ctx context.Context, id string, in model.IntegrationSystemInput) error
	Delete(ctx context.Context, id string) error
}

// IntegrationSystemConverter missing godoc
//go:generate mockery --name=IntegrationSystemConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationSystemConverter interface {
	ToGraphQL(in *model.IntegrationSystem) *graphql.IntegrationSystem
	MultipleToGraphQL(in []*model.IntegrationSystem) []*graphql.IntegrationSystem
	InputFromGraphQL(in graphql.IntegrationSystemInput) model.IntegrationSystemInput
}

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string) ([]pkgmodel.SystemAuth, error)
}

// SystemAuthConverter missing godoc
//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthConverter interface {
	ToGraphQL(in *pkgmodel.SystemAuth) (graphql.SystemAuth, error)
}

// OAuth20Service missing godoc
//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type OAuth20Service interface {
	DeleteMultipleClientCredentials(ctx context.Context, auths []pkgmodel.SystemAuth) error
}

// Resolver missing godoc
type Resolver struct {
	transact persistence.Transactioner

	intSysSvc        IntegrationSystemService
	sysAuthSvc       SystemAuthService
	oAuth20Svc       OAuth20Service
	intSysConverter  IntegrationSystemConverter
	sysAuthConverter SystemAuthConverter
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, intSysSvc IntegrationSystemService, sysAuthSvc SystemAuthService, oAuth20Svc OAuth20Service, intSysConverter IntegrationSystemConverter, sysAuthConverter SystemAuthConverter) *Resolver {
	return &Resolver{
		transact:         transact,
		intSysSvc:        intSysSvc,
		sysAuthSvc:       sysAuthSvc,
		oAuth20Svc:       oAuth20Svc,
		intSysConverter:  intSysConverter,
		sysAuthConverter: sysAuthConverter,
	}
}

// IntegrationSystem missing godoc
func (r *Resolver) IntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	is, err := r.intSysSvc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.intSysConverter.ToGraphQL(is), nil
}

// IntegrationSystems missing godoc
func (r *Resolver) IntegrationSystems(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.IntegrationSystemPage, error) {
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

	intSysPage, err := r.intSysSvc.List(ctx, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlIntSys := r.intSysConverter.MultipleToGraphQL(intSysPage.Data)

	return &graphql.IntegrationSystemPage{
		Data:       gqlIntSys,
		TotalCount: intSysPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(intSysPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(intSysPage.PageInfo.EndCursor),
			HasNextPage: intSysPage.PageInfo.HasNextPage,
		},
	}, nil
}

// RegisterIntegrationSystem missing godoc
func (r *Resolver) RegisterIntegrationSystem(ctx context.Context, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	convertedIn := r.intSysConverter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	id, err := r.intSysSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	intSys, err := r.intSysSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlIntSys := r.intSysConverter.ToGraphQL(intSys)

	return gqlIntSys, nil
}

// UpdateIntegrationSystem missing godoc
func (r *Resolver) UpdateIntegrationSystem(ctx context.Context, id string, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.intSysConverter.InputFromGraphQL(in)
	err = r.intSysSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	intSys, err := r.intSysSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlIntSys := r.intSysConverter.ToGraphQL(intSys)

	return gqlIntSys, nil
}

// UnregisterIntegrationSystem missing godoc
func (r *Resolver) UnregisterIntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	intSys, err := r.intSysSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	auths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.IntegrationSystemReference, intSys.ID)
	if err != nil {
		return nil, err
	}

	err = r.intSysSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	err = r.oAuth20Svc.DeleteMultipleClientCredentials(ctx, auths)
	if err != nil {
		return nil, err
	}

	deletedIntSys := r.intSysConverter.ToGraphQL(intSys)

	return deletedIntSys, nil
}

// Auths missing godoc
func (r *Resolver) Auths(ctx context.Context, obj *graphql.IntegrationSystem) ([]*graphql.IntSysSystemAuth, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Integration System cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.IntegrationSystemReference, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out := make([]*graphql.IntSysSystemAuth, 0, len(sysAuths))
	for _, sa := range sysAuths {
		c, err := r.sysAuthConverter.ToGraphQL(&sa)
		if err != nil {
			return nil, err
		}
		out = append(out, c.(*graphql.IntSysSystemAuth))
	}

	return out, nil
}
