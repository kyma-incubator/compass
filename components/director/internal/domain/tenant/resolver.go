package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery --name=BusinessTenantMappingService --output=automock --outpkg=automock --case=underscore
type BusinessTenantMappingService interface {
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	ListLabels(ctx context.Context, tenantID string) (map[string]*model.Label, error)
	GetTenantByExternalID(ctx context.Context, externalID string) (*model.BusinessTenantMapping, error)
}

//go:generate mockery --name=BusinessTenantMappingConverter --output=automock --outpkg=automock --case=underscore
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

func (r *Resolver) Tenant(ctx context.Context, externalID string) (*graphql.Tenant, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	t, err := r.srv.GetTenantByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlTenant := r.conv.MultipleToGraphQL([]*model.BusinessTenantMapping{t})
	return gqlTenant[0], nil
}

func (r *Resolver) Labels(ctx context.Context, obj *graphql.Tenant, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Tenant cannot be empty")
	}
	log.C(ctx).Infof("getting labels for tenant with ID %s, and internal ID %s", obj.ID, obj.InternalID)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.srv.ListLabels(ctx, obj.InternalID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})
	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	return resultLabels, nil
}

func NewResolver(transact persistence.Transactioner, srv BusinessTenantMappingService, conv BusinessTenantMappingConverter) *Resolver {
	return &Resolver{
		transact: transact,
		srv:      srv,
		conv:     conv,
	}
}
