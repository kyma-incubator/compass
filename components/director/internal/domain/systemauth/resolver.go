package systemauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	Get(ctx context.Context, id string) (*model.SystemAuth, error)
	DeleteByIDForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, authID string) error
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(model *model.SystemAuth) *graphql.SystemAuth
}

type Resolver struct {
	transact persistence.Transactioner
	svc      SystemAuthService
	conv     SystemAuthConverter
}

func NewResolver(transact persistence.Transactioner, svc SystemAuthService, conv SystemAuthConverter) *Resolver {
	return &Resolver{transact: transact, svc: svc, conv: conv}
}

func (r *Resolver) GenericDeleteSystemAuth(objectType model.SystemAuthReferenceObjectType) func(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return func(ctx context.Context, id string) (*graphql.SystemAuth, error) {
		tx, err := r.transact.Begin()
		if err != nil {
			return nil, err
		}
		defer r.transact.RollbackUnlessCommited(tx)

		ctx = persistence.SaveToContext(ctx, tx)

		item, err := r.svc.Get(ctx, id)
		if err != nil {
			return nil, err
		}

		deletedItem := r.conv.ToGraphQL(item)

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
