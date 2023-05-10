package tenantbusinesstype

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tableName string = `public.tenant_business_types`

var (
	tableColumns = []string{"id", "code", "name"}
)

// EntityConverter converts between the internal model and entity
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.TenantBusinessType) *Entity
	FromEntity(entity *Entity) *model.TenantBusinessType
}

type repository struct {
	creatorGlobal repo.CreatorGlobal
	globalGetter  repo.SingleGetterGlobal
	listerGlobal  repo.ListerGlobal
	conv          EntityConverter
}

// NewRepository creates a new FormationAssignment repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creatorGlobal: repo.NewCreatorGlobal(resource.TenantBusinessType, tableName, tableColumns),
		globalGetter:  repo.NewSingleGetterGlobal(resource.TenantBusinessType, tableName, tableColumns),
		listerGlobal:  repo.NewListerGlobal(resource.TenantBusinessType, tableName, tableColumns),
		conv:          conv,
	}
}

// Create creates a new Tenant Business Type in the database with the fields from the model
func (r *repository) Create(ctx context.Context, item *model.TenantBusinessType) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Persisting Tenant Business type entity with ID: %q", item.ID)
	return r.creatorGlobal.Create(ctx, r.conv.ToEntity(item))
}

// GetByID retrieves tenant business type with given id
func (r *repository) GetByID(ctx context.Context, id string) (*model.TenantBusinessType, error) {
	var entity Entity
	if err := r.globalGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// ListAll retrieves all tenant business types
func (r *repository) ListAll(ctx context.Context) ([]*model.TenantBusinessType, error) {
	var entities EntityCollection

	err := r.listerGlobal.ListGlobal(ctx, &entities)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
}

func (r *repository) multipleFromEntities(entities EntityCollection) []*model.TenantBusinessType {
	items := make([]*model.TenantBusinessType, 0, len(entities))
	for _, ent := range entities {
		items = append(items, r.conv.FromEntity(ent))
	}
	return items
}
