package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthService interface {
	GetByIDForObject(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, authID string) (*pkgmodel.SystemAuth, error)
	GetGlobal(ctx context.Context, id string) (*pkgmodel.SystemAuth, error)
	GetByToken(ctx context.Context, token string) (*pkgmodel.SystemAuth, error)
	DeleteByIDForObject(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, authID string) error
	Update(ctx context.Context, item *pkgmodel.SystemAuth) error
	UpdateValue(ctx context.Context, id string, item *model.Auth) (*pkgmodel.SystemAuth, error)
	InvalidateToken(ctx context.Context, id string) (*pkgmodel.SystemAuth, error)
}

// OAuth20Service missing godoc
//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type OAuth20Service interface {
	DeleteClientCredentials(ctx context.Context, clientID string) error
}

// OneTimeTokenService missing godoc
//go:generate mockery --name=OneTimeTokenService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OneTimeTokenService interface {
	IsTokenValid(systemAuth *pkgmodel.SystemAuth) (bool, error)
}

// SystemAuthConverter missing godoc
//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthConverter interface {
	ToGraphQL(model *pkgmodel.SystemAuth) (graphql.SystemAuth, error)
}

// Resolver missing godoc
type Resolver struct {
	transact        persistence.Transactioner
	svc             SystemAuthService
	oAuth20Svc      OAuth20Service
	conv            SystemAuthConverter
	authConv        AuthConverter
	onetimetokenSvc OneTimeTokenService
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, svc SystemAuthService, oAuth20Svc OAuth20Service, onetimetokenSvc OneTimeTokenService, conv SystemAuthConverter, authConverter AuthConverter) *Resolver {
	return &Resolver{transact: transact, svc: svc, oAuth20Svc: oAuth20Svc, onetimetokenSvc: onetimetokenSvc, conv: conv, authConv: authConverter}
}

// GenericDeleteSystemAuth missing godoc
func (r *Resolver) GenericDeleteSystemAuth(objectType pkgmodel.SystemAuthReferenceObjectType) func(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return func(ctx context.Context, id string) (graphql.SystemAuth, error) {
		tx, err := r.transact.Begin()
		if err != nil {
			return nil, err
		}
		defer r.transact.RollbackUnlessCommitted(ctx, tx)

		ctx = persistence.SaveToContext(ctx, tx)

		item, err := r.svc.GetByIDForObject(ctx, objectType, id)
		if err != nil {
			return nil, err
		}

		deletedItem, err := r.conv.ToGraphQL(item)
		if err != nil {
			return nil, errors.Wrap(err, "while converting SystemAuth to GraphQL")
		}

		if item.Value != nil && item.Value.Credential.Oauth != nil {
			err := r.oAuth20Svc.DeleteClientCredentials(ctx, item.Value.Credential.Oauth.ClientID)
			if err != nil {
				return nil, errors.Wrap(err, "while deleting OAuth 2.0 client")
			}
		}

		err = r.svc.DeleteByIDForObject(ctx, objectType, id)
		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		return deletedItem, nil
	}
}

// SystemAuth get a SystemAuth by ID
func (r *Resolver) SystemAuth(ctx context.Context, id string) (graphql.SystemAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	systemAuth, err := r.svc.GetGlobal(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(systemAuth)
}

// SystemAuthByToken gets a SystemAuth by a provided one time token
func (r *Resolver) SystemAuthByToken(ctx context.Context, token string) (graphql.SystemAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	systemAuth, err := r.svc.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	if _, err := r.onetimetokenSvc.IsTokenValid(systemAuth); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(systemAuth)
}

// UpdateSystemAuth updates a SystemAuth with an AuthInput
func (r *Resolver) UpdateSystemAuth(ctx context.Context, id string, in graphql.AuthInput) (graphql.SystemAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Updating System Auth with id %s", id)

	convertedIn, err := r.authConv.ModelFromGraphQLInput(in)
	if err != nil {
		return nil, err
	}

	systemAuth, err := r.svc.UpdateValue(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	log.C(ctx).Infof("System Auth with id %s successfully updated", id)

	return r.conv.ToGraphQL(systemAuth)
}

// InvalidateSystemAuthOneTimeToken checks if the the OTT for the SystemAuth is valid. If yes, it invalidates the OTT. If not, returns an error
func (r *Resolver) InvalidateSystemAuthOneTimeToken(ctx context.Context, id string) (graphql.SystemAuth, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	systemAuth, err := r.svc.GetGlobal(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err := r.onetimetokenSvc.IsTokenValid(systemAuth); err != nil {
		return nil, err
	}

	systemAuth, err = r.svc.InvalidateToken(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(systemAuth)
}
