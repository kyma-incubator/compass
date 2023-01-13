package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tableName string = `public.formations`

var (
	updatableTableColumns   = []string{"name"}
	idTableColumns          = []string{"id"}
	tableColumns            = []string{"id", "tenant_id", "formation_template_id", "name"}
	tenantColumn            = "tenant_id"
	formationTemplateColumn = "formation_template_id"
)

// EntityConverter converts between the internal model and entity
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Formation) *Entity
	FromEntity(entity *Entity) *model.Formation
}

type repository struct {
	creator               repo.CreatorGlobal
	getter                repo.SingleGetter
	pageableQuerierGlobal repo.PageableQuerier
	listerGlobal          repo.ListerGlobal
	updater               repo.UpdaterGlobal
	deleter               repo.Deleter
	existQuerier          repo.ExistQuerier
	conv                  EntityConverter
}

// NewRepository creates a new Formation repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:               repo.NewCreatorGlobal(resource.Formations, tableName, tableColumns),
		getter:                repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		listerGlobal:          repo.NewListerGlobal(resource.Formations, tableName, tableColumns),
		updater:               repo.NewUpdaterWithEmbeddedTenant(resource.Formations, tableName, updatableTableColumns, tenantColumn, idTableColumns),
		deleter:               repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		existQuerier:          repo.NewExistQuerierWithEmbeddedTenant(tableName, tenantColumn),
		conv:                  conv,
	}
}

// Create creates a Formation with a given input
func (r *repository) Create(ctx context.Context, item *model.Formation) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Formation with name: %q to entity", item.Name)
	entity := r.conv.ToEntity(item)

	log.C(ctx).Debugf("Persisting Formation entity with name: %q to the DB", item.Name)
	return r.creator.Create(ctx, entity)
}

// Get returns a Formation by a given id
func (r *repository) Get(ctx context.Context, id, tenantID string) (*model.Formation, error) {
	var entity Entity
	if err := r.getter.Get(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation with id: %q", id)
		return nil, errors.Wrapf(err, "An error occurred while getting formation with id: %q", id)
	}

	return r.conv.FromEntity(&entity), nil
}

// GetByName returns a Formations by a given name
func (r *repository) GetByName(ctx context.Context, name, tenantID string) (*model.Formation, error) {
	var entity Entity
	if err := r.getter.Get(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("name", name)}, repo.NoOrderBy, &entity); err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation with name: %q", name)
		return nil, errors.Wrapf(err, "An error occurred while getting formation with name: %q", name)
	}

	return r.conv.FromEntity(&entity), nil
}

// List returns all Formations sorted by id and paginated by the pageSize and cursor parameters
func (r *repository) List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.FormationPage, error) {
	var entityCollection EntityCollection
	page, totalCount, err := r.pageableQuerierGlobal.List(ctx, resource.Formations, tenant, pageSize, cursor, "id", &entityCollection)
	if err != nil {
		return nil, err
	}

	items := make([]*model.Formation, 0, entityCollection.Len())

	for _, entity := range entityCollection {
		formationModel := r.conv.FromEntity(entity)

		items = append(items, formationModel)
	}
	return &model.FormationPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListByFormationTemplateID returns all Formations for FormationTemplate with the provided ID
func (r *repository) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.Formation, error) {
	var entityCollection EntityCollection
	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, repo.NewEqualCondition(formationTemplateColumn, formationTemplateID)); err != nil {
		return nil, err
	}

	items := make([]*model.Formation, 0, entityCollection.Len())

	for _, entity := range entityCollection {
		formationModel := r.conv.FromEntity(entity)

		items = append(items, formationModel)
	}
	return items, nil
}

// Update updates a Formation with the given input
func (r *repository) Update(ctx context.Context, model *model.Formation) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}
	return r.updater.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// DeleteByName deletes a Formation with given name
func (r *repository) DeleteByName(ctx context.Context, tenantID, name string) error {
	log.C(ctx).Debugf("Deleting formation with name: %q...", name)
	return r.deleter.DeleteOne(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("name", name)})
}

// Exists check if a Formation with given ID exists
func (r *repository) Exists(ctx context.Context, id, tenantID string) (bool, error) {
	log.C(ctx).Debugf("Cheking if formation with ID: %q exists...", id)
	return r.existQuerier.Exists(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}
