package formation

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.formations`

var (
	updatableTableColumns = []string{"name", "state", "error", "last_state_change_timestamp", "last_notification_sent_timestamp"}
	idTableColumns        = []string{"id"}
	tableColumns          = []string{"id", "tenant_id", "formation_template_id", "name", "state", "error", "last_state_change_timestamp", "last_notification_sent_timestamp"}
	tenantColumn          = "tenant_id"
	formationNameColumn   = "name"
	idTableColumn         = "id"
)

// EntityConverter converts between the internal model and entity
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Formation) *Entity
	FromEntity(entity *Entity) *model.Formation
}

type repository struct {
	creator         repo.CreatorGlobal
	getter          repo.SingleGetter
	globalGetter    repo.SingleGetterGlobal
	pageableQuerier repo.PageableQuerier
	lister          repo.Lister
	listerGlobal    repo.ListerGlobal
	updater         repo.UpdaterGlobal
	deleter         repo.Deleter
	existQuerier    repo.ExistQuerier
	conv            EntityConverter
}

// NewRepository creates a new Formation repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:         repo.NewCreatorGlobal(resource.Formations, tableName, tableColumns),
		getter:          repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		globalGetter:    repo.NewSingleGetterGlobal(resource.Formations, tableName, tableColumns),
		pageableQuerier: repo.NewPageableQuerierWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		lister:          repo.NewListerWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		listerGlobal:    repo.NewListerGlobal(resource.Formations, tableName, tableColumns),
		updater:         repo.NewUpdaterWithEmbeddedTenant(resource.Formations, tableName, updatableTableColumns, tenantColumn, idTableColumns),
		deleter:         repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		existQuerier:    repo.NewExistQuerierWithEmbeddedTenant(tableName, tenantColumn),
		conv:            conv,
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

// GetGlobalByID retrieves formation matching ID `id` globally without tenant parameter
func (r *repository) GetGlobalByID(ctx context.Context, id string) (*model.Formation, error) {
	log.C(ctx).Debugf("Getting formation with ID: %q globally", id)
	var entity Entity
	if err := r.globalGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
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
	page, totalCount, err := r.pageableQuerier.List(ctx, resource.Formations, tenant, pageSize, cursor, "id", &entityCollection)
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

// ListByIDsGlobal returns all Formations with id in `formationIDs` globally
func (r *repository) ListByIDsGlobal(ctx context.Context, formationIDs []string) ([]*model.Formation, error) {
	if len(formationIDs) == 0 {
		return nil, nil
	}

	var entityCollection EntityCollection
	err := r.listerGlobal.ListGlobal(ctx, &entityCollection, repo.NewInConditionForStringValues(idTableColumn, formationIDs))
	if err != nil {
		return nil, err
	}

	items := make([]*model.Formation, 0, entityCollection.Len())
	for _, entity := range entityCollection {
		formationModel := r.conv.FromEntity(entity)

		items = append(items, formationModel)
	}

	return items, nil
}

// ListByFormationNames returns all Formations with name in formationNames
func (r *repository) ListByFormationNames(ctx context.Context, formationNames []string, tenantID string) ([]*model.Formation, error) {
	var entityCollection EntityCollection
	if err := r.lister.List(ctx, resource.Formations, tenantID, &entityCollection, repo.NewInConditionForStringValues(formationNameColumn, formationNames)); err != nil {
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
	newEntity := r.conv.ToEntity(model)

	var retrievedEntity Entity
	if err := r.globalGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", model.ID)}, repo.NoOrderBy, &retrievedEntity); err != nil {
		return err
	}

	if retrievedEntity.State != newEntity.State {
		log.C(ctx).Debugf("State of formation with ID: %s was changed from: %s to: %s, updating the last state change timestamp", newEntity.ID, retrievedEntity.State, newEntity.State)
		now := time.Now()
		newEntity.LastStateChangeTimestamp = &now
	}

	log.C(ctx).Debugf("Updating formation with ID: %q and name: %q...", newEntity.ID, newEntity.Name)
	return r.updater.UpdateSingleGlobal(ctx, newEntity)
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
