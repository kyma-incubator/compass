package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	GetByIDForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, authID string) (*model.SystemAuth, error)
	GetGlobal(ctx context.Context, id string) (*model.SystemAuth, error)
	DeleteByIDForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, authID string) error
}

//go:generate mockery -name=OAuth20Service -output=automock -outpkg=automock -case=underscore
type OAuth20Service interface {
	DeleteClientCredentials(ctx context.Context, clientID string) error
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(model *model.SystemAuth) (*graphql.SystemAuth, error)
}

type Resolver struct {
	transact   persistence.Transactioner
	svc        SystemAuthService
	oAuth20Svc OAuth20Service
	conv       SystemAuthConverter
}

func NewResolver(transact persistence.Transactioner, svc SystemAuthService, oAuth20Svc OAuth20Service, conv SystemAuthConverter) *Resolver {
	return &Resolver{transact: transact, svc: svc, oAuth20Svc: oAuth20Svc, conv: conv}
}

func (r *Resolver) GenericDeleteSystemAuth(objectType model.SystemAuthReferenceObjectType) func(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return func(ctx context.Context, id string) (*graphql.SystemAuth, error) {
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
