package tenantindex

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"

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
func (r *repository) GetOwnerTenantByResourceID(ctx context.Context, callingTenant, resourceID string) (string, error) {
	var ownerTenant string
	err := r.singleGetter.Get(ctx, callingTenant, repo.Conditions{repo.NewEqualCondition(idColumn, resourceID)}, repo.NoOrderBy, &ownerTenant)
	if apperrors.IsNotFoundError(err) {
		if err := r.materialize(ctx); err != nil { // If the resource is not found in the view we should refresh it as it can be a new resource.
			return "", errors.Wrap(err, "while materializing view")
		}
		err = r.singleGetter.Get(ctx, callingTenant, repo.Conditions{repo.NewEqualCondition(idColumn, resourceID)}, repo.NoOrderBy, &ownerTenant)
	}
	return ownerTenant, err
}

// materialize refreshes the materialized tenant index view concurrently.
// Concurrently means that instead of locking the view when refreshing it, a separate copy will be created which then will be refreshed and
// only the diff from current and new state of the view will be inserted. This makes the view available for readers while refreshing.
// Concurrent refresh of the view is way more computational intensive, however it is crucial for our use-case because
// we don't want all the clients to be locked on reading while the view is refreshed, since most of the resource will be already in the previous state of the view.
func (r *repository) materialize(ctx context.Context) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Refreshing materialized view %s...", tenantIndexTableName)

	_, err = persist.ExecContext(ctx, fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", tenantIndexTableName))
	return err
}
