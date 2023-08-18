package operation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const operationTable = `public.operation`
const priorityView = `public.scheduled_operations`

var (
	idTableColumns        = []string{"id"}
	operationColumns      = []string{"id", "op_type", "status", "data", "error", "priority", "created_at", "updated_at"}
	updatableTableColumns = []string{"status", "error", "priority", "updated_at"}
)

// EntityConverter missing godoc
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Operation) *Entity
	FromEntity(entity *Entity) *model.Operation
}

type pgRepository struct {
	globalCreator      repo.CreatorGlobal
	globalDeleter      repo.DeleterGlobal
	globalUpdater      repo.UpdaterGlobal
	globalSingleGetter repo.SingleGetterGlobal
	functionBuilder    repo.FunctionBuilder
	prioritiViewLister repo.ListerGlobal
	conv               EntityConverter
}

// NewRepository creates new operation repository
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		globalCreator:      repo.NewCreatorGlobal(resource.Operation, operationTable, operationColumns),
		globalDeleter:      repo.NewDeleterGlobal(resource.Operation, operationTable),
		globalUpdater:      repo.NewUpdaterGlobal(resource.Operation, operationTable, updatableTableColumns, idTableColumns),
		globalSingleGetter: repo.NewSingleGetterGlobal(resource.Operation, operationTable, operationColumns),
		functionBuilder:    repo.NewFunctionBuilder(),
		prioritiViewLister: repo.NewListerGlobal(resource.Operation, priorityView, operationColumns),
		conv:               conv,
	}
}

// Create creates operation entity
func (r *pgRepository) Create(ctx context.Context, model *model.Operation) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Operation model with id %s to entity", model.ID)
	operationEnt := r.conv.ToEntity(model)

	log.C(ctx).Debugf("Persisting Operation entity with id %s to db", model.ID)
	return r.globalCreator.Create(ctx, operationEnt)
}

// Update updates the operation
func (r *pgRepository) Update(ctx context.Context, model *model.Operation) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Operation model with id %s to entity", model.ID)
	operationEnt := r.conv.ToEntity(model)

	log.C(ctx).Debugf("Updating Operation entity with id %s", model.ID)
	return r.globalUpdater.UpdateSingleGlobal(ctx, operationEnt)
}

// Get retrieves an operation by id
func (r *pgRepository) Get(ctx context.Context, id string) (*model.Operation, error) {
	var entity Entity
	if err := r.globalSingleGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// Delete deletes an operation by id
func (r *pgRepository) Delete(ctx context.Context, id string) error {
	return r.globalDeleter.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteMultiple deletes all operations by given list of ids
func (r *pgRepository) DeleteMultiple(ctx context.Context, ids []string) error {
	return r.globalDeleter.DeleteManyGlobal(ctx, repo.Conditions{repo.NewInConditionForStringValues("id", ids)})
}

// PriorityQueueListByType returns top priority operations from priority view for specified type
func (r *pgRepository) PriorityQueueListByType(ctx context.Context, operationType string) ([]*model.Operation, error) {
	var entities EntityCollection
	if err := r.prioritiViewLister.ListGlobalWithLimit(ctx, &entities, 10, repo.Conditions{repo.NewEqualCondition("op_type", operationType)}...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
}

func (r *pgRepository) multipleFromEntities(entities EntityCollection) []*model.Operation {
	items := make([]*model.Operation, 0, len(entities))
	for _, ent := range entities {
		model := r.conv.FromEntity(&ent)

		items = append(items, model)
	}
	return items
}
