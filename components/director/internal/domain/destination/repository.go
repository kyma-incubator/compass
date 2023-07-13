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
	destinationNameColumn       = "name"
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
	conv                       EntityConverter
	getter                     repo.SingleGetter
	globalCreator              repo.CreatorGlobal
	deleter                    repo.Deleter
	globalDeleter              repo.DeleterGlobal
	upserterWithEmbeddedTenant repo.UpserterGlobal
	upserterGlobal             repo.UpserterGlobal
	lister                     repo.Lister
}

// NewRepository returns new destination repository
func NewRepository(converter EntityConverter) *repository {
	return &repository{
		conv:                       converter,
		getter:                     repo.NewSingleGetter(destinationTable, destinationColumns),
		globalCreator:              repo.NewCreatorGlobal(resource.Destination, destinationTable, destinationColumns),
		deleter:                    repo.NewDeleterWithEmbeddedTenant(destinationTable, tenantIDColumn),
		globalDeleter:              repo.NewDeleterGlobal(resource.Destination, destinationTable),
		upserterWithEmbeddedTenant: repo.NewUpserterWithEmbeddedTenant(resource.Destination, destinationTable, destinationColumns, conflictingColumns, updateColumns, tenantIDColumn),
		upserterGlobal:             repo.NewUpserterGlobal(resource.Destination, destinationTable, destinationColumns, conflictingColumns, updateColumns),
		lister:                     repo.NewListerWithEmbeddedTenant(destinationTable, tenantIDColumn, destinationColumns),
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
	return r.upserterGlobal.UpsertGlobal(ctx, destination)
}

// UpsertWithEmbeddedTenant upserts a destination entity in th DB with embedded tenant
func (r *repository) UpsertWithEmbeddedTenant(ctx context.Context, destination *model.Destination) error {
	if destination == nil {
		return apperrors.NewInternalError("destination model can not be empty")
	}

	return r.upserterWithEmbeddedTenant.UpsertGlobal(ctx, r.conv.ToEntity(destination))
}

// DeleteOld deletes all destinations in a given tenant that do not have latestRevision
func (r *repository) DeleteOld(ctx context.Context, latestRevision, tenantID string) error {
	conditions := repo.Conditions{repo.NewNotEqualCondition(revisionColumn, latestRevision), repo.NewEqualCondition(tenantIDColumn, tenantID), repo.NewNotNullCondition(revisionColumn)}
	return r.globalDeleter.DeleteManyGlobal(ctx, conditions)
}

// CreateDestination creates destination in the DB with the provided `destination` data
func (r *repository) CreateDestination(ctx context.Context, destination *model.Destination) error {
	if destination == nil {
		return apperrors.NewInternalError("destination model can not be empty")
	}

	return r.globalCreator.Create(ctx, r.conv.ToEntity(destination))
}

// GetDestinationByNameAndTenant todo::: go doc
func (r *repository) GetDestinationByNameAndTenant(ctx context.Context, destinationName, tenantID string) (*model.Destination, error) {
	log.C(ctx).Infof("Getting destinations with name: %q and tenant ID: %q", destinationName, tenantID)

	var dest Entity
	conditions := repo.Conditions{repo.NewEqualCondition(destinationNameColumn, destinationName), repo.NewEqualCondition(tenantIDColumn, tenantID)}
	if err := r.getter.Get(ctx, resource.Destination, tenantID, conditions, repo.NoOrderBy, &dest); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&dest), nil
}

// ListByTenantIDAndAssignmentID returns all destinations for a given `tenantID` and `formationAssignmentID`
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

// DeleteByDestinationNameAndAssignmentID deletes all destinations for a given `destinationName`, `formationAssignmentID` and `tenantID` from the DB
func (r *repository) DeleteByDestinationNameAndAssignmentID(ctx context.Context, destinationName, formationAssignmentID, tenantID string) error {
	log.C(ctx).Infof("Deleting destination(s) by name: %q, assignment ID: %q and tenant ID: %q from the DB", destinationName, tenantID, formationAssignmentID)
	conditions := repo.Conditions{repo.NewEqualCondition(destinationNameColumn, destinationName), repo.NewEqualCondition(formationAssignmentIDColumn, formationAssignmentID)}
	return r.deleter.DeleteMany(ctx, resource.Destination, tenantID, conditions)
}
