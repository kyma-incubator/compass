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
	tenantParentsTable = "tenant_parents"
	tenantIDColumn     = "tenant_id"
	parentIDColumn     = "parent_id"
)

var (
	tenantParentsSelectedColumns = []string{tenantIDColumn, parentIDColumn}
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
	var parentIDs []string
	for _, tp := range tpc {
		parentIDs = append(parentIDs, tp.ParentID)
	}
	return parentIDs
}

// GetTenantIDs returns list of all tenant ids.
func (tpc TenantParentCollection) GetTenantIDs() []string {
	var tenantIDs []string
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
	CreateMultiple(ctx context.Context, tenantID string, parentIDs []string) error
	Create(ctx context.Context, tenantID string, parentID string) error
	Delete(ctx context.Context, tenantID string, parentID string) error
}

// NewRepository returns a new entity responsible for repo-layer tenant operations. All of its methods require persistence.PersistenceOp it the provided context.
func NewRepository() *pgRepository {
	return &pgRepository{
		listerGlobal:   repo.NewListerGlobal(resource.TenantParent, tenantParentsTable, tenantParentsSelectedColumns),
		creatorGlobal:  repo.NewCreatorGlobal(resource.TenantParent, tenantParentsTable, tenantParentsSelectedColumns),
		upserterGlobal: repo.NewUpserterGlobal(resource.TenantParent, tenantParentsTable, tenantParentsSelectedColumns, tenantParentsSelectedColumns, []string{}),
		deleterGlobal:  repo.NewDeleterGlobal(resource.TenantParent, tenantParentsTable),
	}
}

// ListParents lists all parents of the provided tenant
func (r *pgRepository) ListParents(ctx context.Context, tenantID string) ([]string, error) {
	tenantParents := TenantParentCollection{}
	conditions := repo.Conditions{
		repo.NewEqualCondition(tenantIDColumn, tenantID),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &tenantParents, conditions...); err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.List, "while listing tenant parent records from '%s' table", tenantParentsTable))
		return nil, err
	}

	return tenantParents.GetParentIDs(), nil
}

// ListByParent lists all tenant ids by the provided parent id
func (r *pgRepository) ListByParent(ctx context.Context, parentID string) ([]string, error) {
	tenantParents := TenantParentCollection{}
	conditions := repo.Conditions{
		repo.NewEqualCondition(parentIDColumn, parentID),
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

// TODO change the method name
// CreateMultiple creates new tenant parent mappings
func (r *pgRepository) CreateMultiple(ctx context.Context, tenantID string, parentIDs []string) error {
	for _, parentID := range parentIDs {
		if err := r.Upsert(ctx, tenantID, parentID); err != nil {
			log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.Create, "while creating tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID))
			return errors.Wrapf(err, "while creating tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID)
		}
	}

	return nil
}

func (r *pgRepository) Create(ctx context.Context, tenantID string, parentID string) error {
	if err := r.Upsert(ctx, tenantID, parentID); err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.Create, "while creating tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID))
		return err
	}

	return nil
}

func (r *pgRepository) Delete(ctx context.Context, tenantID string, parentID string) error {
	if err := r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{
		repo.NewEqualCondition(tenantIDColumn, tenantID),
		repo.NewEqualCondition(parentIDColumn, parentID),
	}); err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantParent, resource.Create, "while deleting tenant parent mapping for tenant with id %s and parent %s", tenantID, parentID))
		return err
	}

	return nil
}
