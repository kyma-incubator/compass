package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore
type SystemAuthService interface {
	GetByIDForObject(ctx context.Context, objectType systemauth.SystemAuthReferenceObjectType, authID string) (*systemauth.SystemAuth, error)
	GetGlobal(ctx context.Context, id string) (*systemauth.SystemAuth, error)
	DeleteByIDForObject(ctx context.Context, objectType systemauth.SystemAuthReferenceObjectType, authID string) error
	Update(ctx context.Context, item *systemauth.SystemAuth) error
	UpdateValue(ctx context.Context, id string, item *model.Auth) (*systemauth.SystemAuth, error)
}

// OAuth20Service missing godoc
//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore
type OAuth20Service interface {
	DeleteClientCredentials(ctx context.Context, clientID string) error
}

// SystemAuthConverter missing godoc
//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore
type SystemAuthConverter interface {
	ToGraphQL(model *systemauth.SystemAuth) (graphql.SystemAuth, error)
}

// Resolver missing godoc
type Resolver struct {
	transact   persistence.Transactioner
	svc        SystemAuthService
	oAuth20Svc OAuth20Service
	conv       SystemAuthConverter
	authConv   AuthConverter
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, svc SystemAuthService, oAuth20Svc OAuth20Service, conv SystemAuthConverter, authConverter AuthConverter) *Resolver {
	return &Resolver{transact: transact, svc: svc, oAuth20Svc: oAuth20Svc, conv: conv, authConv: authConverter}
}

// GenericDeleteSystemAuth missing godoc
func (r *Resolver) GenericDeleteSystemAuth(objectType systemauth.SystemAuthReferenceObjectType) func(ctx context.Context, id string) (graphql.SystemAuth, error) {
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

// SystemAuth missing godoc
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.conv.ToGraphQL(systemAuth)
}

// UpdateSystemAuth missing godoc
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("System Auth with id %s successfully updated", id)

	return r.conv.ToGraphQL(systemAuth)
}
