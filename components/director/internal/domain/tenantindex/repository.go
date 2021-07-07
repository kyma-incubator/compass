package tenantindex

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tenantIndexTableName string = `"public"."id_tenant_id_index"` // It is a view

var (
	tenantColumn = "tenant_id"
	idColumn     = "id"
)

type repository struct {
	singleGetter repo.SingleGetter
}

// NewRepository returns new Tenant Index Repository
func NewRepository() *repository {
	return &repository{
		singleGetter: repo.NewSingleGetter(resource.API, tenantIndexTableName, tenantColumn, []string{tenantColumn}),
	}
}

// GetOwnerTenantByResourceID gets the owner tenant of a resource by ID checking if the calling tenant has access to it.
// The calling tenant can be of type customer and the resource can be part of any of his child 'account' tenants.
// This function will return the tenant owning the resource in case the callingTenant has access to it (it is his child tenant) or NotFoundErr otherwise.
func (r *repository) GetOwnerTenantByResourceID(ctx context.Context, callingTenant, resourceId string) (string, error) {
	var ownerTenant string
	err := r.singleGetter.Get(ctx, callingTenant, repo.Conditions{repo.NewEqualCondition(idColumn, resourceId)}, repo.NoOrderBy, &ownerTenant)
	return ownerTenant, err
}
