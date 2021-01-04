package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=BusinessTenantMappingService -output=automock -outpkg=automock -case=underscore
type BusinessTenantMappingService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
}

//go:generate mockery -name=BusinessTenantMappingConverter -output=automock -outpkg=automock -case=underscore
type BusinessTenantMappingConverter interface {
	MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant
}

type Resolver struct {
	transact persistence.Transactioner

	srv  BusinessTenantMappingService
	conv BusinessTenantMappingConverter
}

func (r *Resolver) Tenants(ctx context.Context) ([]*graphql.Tenant, error) {

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tenants, err := r.srv.List(ctx)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenants := r.conv.MultipleToGraphQL(tenants)
	return gqlTenants, nil

}

func NewResolver(transact persistence.Transactioner, srv BusinessTenantMappingService, conv BusinessTenantMappingConverter) *Resolver {
	return &Resolver{
		transact: transact,
		srv:      srv,
		conv:     conv,
	}
}
