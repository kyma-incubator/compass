package destination

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	destinationTable = "public.destinations"
	revisionColumn   = "revision"
	tenantIDColumn   = "tenant_id"
)

var (
	destinationColumns = []string{"id", "name", "type", "url", "authentication", "tenant_id", "bundle_id", "revision"}
	conflictingColumns = []string{"name", "tenant_id"}
	updateColumns      = []string{"name", "type", "url", "authentication", "revision"}
)

type repository struct {
	deleter  repo.DeleterGlobal
	upserter repo.UpserterGlobal
}

// NewRepository returns new destination repository
func NewRepository() *repository {
	return &repository{
		deleter:  repo.NewDeleterGlobal(resource.Destination, destinationTable),
		upserter: repo.NewUpserterGlobal(resource.Destination, destinationTable, destinationColumns, conflictingColumns, updateColumns),
	}
}

// Upsert upserts a destination entity in db
func (r *repository) Upsert(ctx context.Context, in model.DestinationInput, id, tenantID, bundleID, revisionID string) error {
	destination := Entity{
		ID:             id,
		Name:           in.Name,
		Type:           in.Type,
		URL:            in.URL,
		Authentication: in.Authentication,
		BundleID:       bundleID,
		TenantID:       tenantID,
		Revision:       revisionID,
	}
	return r.upserter.UpsertGlobal(ctx, destination)
}

// DeleteOld deletes all destinations in a given tenant that do not have latestRevision
func (r *repository) DeleteOld(ctx context.Context, latestRevision, tenantID string) error {
	conditions := repo.Conditions{repo.NewNotEqualCondition(revisionColumn, latestRevision), repo.NewEqualCondition(tenantIDColumn, tenantID)}
	return r.deleter.DeleteManyGlobal(ctx, conditions)
}
