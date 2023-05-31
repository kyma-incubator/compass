package destination

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	destinationTable            = "public.destinations"
	revisionColumn              = "revision"
	tenantIDColumn              = "tenant_id"
	formationAssignmentIDColumn = "formation_assignment_id"
)

var (
	destinationColumns = []string{"id", "name", "type", "url", "authentication", "tenant_id", "bundle_id", "revision", "formation_assignment_id"}
	conflictingColumns = []string{"name", "tenant_id"}
	updateColumns      = []string{"name", "type", "url", "authentication", "revision"}
)

// EntityConverter missing godoc
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Destination) *Entity
	FromEntity(entity *Entity) *model.Destination
}

type repository struct {
	conv          EntityConverter
	globalCreator repo.CreatorGlobal
	deleter       repo.Deleter
	globalDeleter repo.DeleterGlobal
	upserter      repo.UpserterGlobal
	lister        repo.Lister
}

// NewRepository returns new destination repository
func NewRepository(converter EntityConverter) *repository {
	return &repository{
		conv:          converter,
		globalCreator: repo.NewCreatorGlobal(resource.Destination, destinationTable, destinationColumns),
		deleter:       repo.NewDeleterWithEmbeddedTenant(destinationTable, tenantIDColumn),
		globalDeleter: repo.NewDeleterGlobal(resource.Destination, destinationTable),
		upserter:      repo.NewUpserterGlobal(resource.Destination, destinationTable, destinationColumns, conflictingColumns, updateColumns),
		lister:        repo.NewListerWithEmbeddedTenant(destinationTable, tenantIDColumn, destinationColumns),
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
		BundleID:       repo.NewNullableString(&bundleID),
		TenantID:       tenantID,
		Revision:       repo.NewNullableString(&revisionID),
	}
	return r.upserter.UpsertGlobal(ctx, destination)
}

// DeleteOld deletes all destinations in a given tenant that do not have latestRevision
func (r *repository) DeleteOld(ctx context.Context, latestRevision, tenantID string) error {
	conditions := repo.Conditions{repo.NewNotEqualCondition(revisionColumn, latestRevision), repo.NewEqualCondition(tenantIDColumn, tenantID), repo.NewNotNullCondition(revisionColumn)}
	return r.globalDeleter.DeleteManyGlobal(ctx, conditions)
}

func (r *repository) CreateDestination(ctx context.Context, destination *model.Destination) error {
	if destination == nil {
		return apperrors.NewInternalError("destination model can not be empty")
	}

	return r.globalCreator.Create(ctx, r.conv.ToEntity(destination))
}

func (r *repository) ListByTenantIDAndAssignmentID(ctx context.Context, tenantID, formationAssignmentID string) ([]*model.Destination, error) {
	log.C(ctx).Infof("Listing destinations by tenant ID: %q and assignment ID: %q from the DB", tenantID, formationAssignmentID)
	var destCollection EntityCollection
	conditions := repo.Conditions{repo.NewEqualCondition(formationAssignmentIDColumn, formationAssignmentID)}
	if err := r.lister.List(ctx, resource.Destination, tenantID, &destCollection, conditions...); err != nil {
		return nil, err
	}

	items := make([]*model.Destination, 0, destCollection.Len())
	for _, destEntity := range destCollection {
		items = append(items, r.conv.FromEntity(&destEntity))
	}

	return items, nil
}

func (r *repository) DeleteByTenantIDAndAssignmentID(ctx context.Context, tenantID, formationAssignmentID string) error {
	log.C(ctx).Infof("Deleting destination(s) by tenant ID: %q and assignment ID: %q from the DB", tenantID, formationAssignmentID)
	conditions := repo.Conditions{repo.NewEqualCondition(formationAssignmentIDColumn, formationAssignmentID)}
	return r.deleter.DeleteMany(ctx, resource.Destination, tenantID, conditions)
}
