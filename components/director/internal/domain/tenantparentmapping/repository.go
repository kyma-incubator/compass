package tenantparentmapping

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	// TenantParentsTable the name of the table containing parency relations
	TenantParentsTable = "tenant_parents"
	// TenantIDColumn the column containing the tenant ID
	TenantIDColumn = "tenant_id"
	// ParentIDColumn the column containing the parent ID
	ParentIDColumn = "parent_id"
)

var (
	tenantParentsSelectedColumns = []string{TenantIDColumn, ParentIDColumn}
)

// TenantParent defines tenant parent record
type TenantParent struct {
	TenantID string `db:"tenant_id"`
	ParentID string `db:"parent_id"`
}

// TenantParentCollection is a wrapper type for slice of entities.
type TenantParentCollection []TenantParent

// Len returns the current number of entities in the collection.
func (tpc TenantParentCollection) Len() int {
	return len(tpc)
}

// GetParentIDs returns list of all parent ids.
func (tpc TenantParentCollection) GetParentIDs() []string {
	parentIDs := make([]string, 0, len(tpc))
	for _, tp := range tpc {
		parentIDs = append(parentIDs, tp.ParentID)
	}
	return parentIDs
}

// GetTenantIDs returns list of all tenant ids.
func (tpc TenantParentCollection) GetTenantIDs() []string {
	tenantIDs := make([]string, 0, len(tpc))
	for _, tp := range tpc {
		tenantIDs = append(tenantIDs, tp.TenantID)
	}
	return tenantIDs
}

type pgRepository struct {
	listerGlobal   repo.ListerGlobal
	creatorGlobal  repo.CreatorGlobal
	upserterGlobal repo.UpserterGlobal
	deleterGlobal  repo.DeleterGlobal
}

// TenantParentRepository is responsible for the repo-layer tenant operations.
//
//go:generate mockery --name=TenantParentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantParentRepository interface {
	ListParents(ctx context.Context, tenantID string) ([]string, error)
	ListByParent(ctx context.Context, parentID string) ([]string, error)
	UpsertMultiple(ctx context.Context, tenantID string, parentIDs []string) error
	Upsert(ctx context.Context, tenantID string, parentID string) error
	Delete(ctx context.Context, tenantID string, parentID string) error
}

// NewRepository returns a new entity responsible for repo-layer tenant operations. All of its methods require persistence.PersistenceOp it the provided context.
func NewRepository() *pgRepository {
	return &pgRepository{
		listerGlobal:   repo.NewListerGlobal(resource.TenantParent, TenantParentsTable, tenantParentsSelectedColumns),
		creatorGlobal:  repo.NewCreatorGlobal(resource.TenantParent, TenantParentsTable, tenantParentsSelectedColumns),
		upserterGlobal: repo.NewUpserterGlobal(resource.TenantParent, TenantParentsTable, tenantParentsSelectedColumns, tenantParentsSelectedColumns, []string{}),
		deleterGlobal:  repo.NewDeleterGlobal(resource.TenantParent, TenantParentsTable),
	}
}

// ListParents lists all parents of the provided tenant
func (r *pgRepository) ListParents(ctx context.Context, tenantID string) ([]string, error) {
	tenantParents := TenantParentCollection{}
	conditions := repo.Conditions{
		repo.NewEqualCondition(TenantIDColumn, tenantID),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &tenantParents, conditions...); err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.List, "while listing tenant parent records from '%s' table", TenantParentsTable))
		return nil, err
	}

	return tenantParents.GetParentIDs(), nil
}

// ListByParent lists all tenant ids by the provided parent id
func (r *pgRepository) ListByParent(ctx context.Context, parentID string) ([]string, error) {
	tenantParents := TenantParentCollection{}
	conditions := repo.Conditions{
		repo.NewEqualCondition(ParentIDColumn, parentID),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &tenantParents, conditions...); err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.List, "while listing tenants by parent with id %s", parentID))
		return nil, err
	}

	return tenantParents.GetTenantIDs(), nil
}

// Upsert upserts tenant parent mapping
func (r *pgRepository) Upsert(ctx context.Context, tenantID string, parentID string) error {
	tpEntity := &TenantParent{
		TenantID: tenantID,
		ParentID: parentID,
	}
	return r.upserterGlobal.UpsertGlobal(ctx, tpEntity)
}

// UpsertMultiple upserts new tenant parent mappings
func (r *pgRepository) UpsertMultiple(ctx context.Context, tenantID string, parentIDs []string) error {
	for _, parentID := range parentIDs {
		if err := r.Upsert(ctx, tenantID, parentID); err != nil {
			log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.Create, "while creating tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID))
			return errors.Wrapf(err, "while creating tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID)
		}
	}

	return nil
}

// Delete deletes the tenant parent mapping for tenantID and parentID
func (r *pgRepository) Delete(ctx context.Context, tenantID string, parentID string) error {
	if err := r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{
		repo.NewEqualCondition(TenantIDColumn, tenantID),
		repo.NewEqualCondition(ParentIDColumn, parentID),
	}); err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.Create, "while deleting tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID))
		return err
	}

	return nil
}
