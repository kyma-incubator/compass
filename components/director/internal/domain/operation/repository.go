package operation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/util"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const operationTable = `public.operation`
const scheduledOperationsView = `public.scheduled_operations`

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
	globalLister                  repo.ListerGlobal
	globalCreator                 repo.CreatorGlobal
	globalDeleter                 repo.DeleterGlobal
	globalUpdater                 repo.UpdaterGlobal
	globalSingleGetter            repo.SingleGetterGlobal
	globalFunctioner              repo.FunctionerGlobal
	scheduledOperationsViewLister repo.ListerGlobal
	conv                          EntityConverter
}

// NewRepository creates new operation repository
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		globalLister:                  repo.NewListerGlobal(resource.Operation, operationTable, operationColumns),
		globalCreator:                 repo.NewCreatorGlobal(resource.Operation, operationTable, operationColumns),
		globalDeleter:                 repo.NewDeleterGlobal(resource.Operation, operationTable),
		globalUpdater:                 repo.NewUpdaterGlobal(resource.Operation, operationTable, updatableTableColumns, idTableColumns),
		globalSingleGetter:            repo.NewSingleGetterGlobal(resource.Operation, operationTable, operationColumns),
		globalFunctioner:              repo.NewFunctionerGlobal(),
		scheduledOperationsViewLister: repo.NewListerGlobal(resource.Operation, scheduledOperationsView, operationColumns),
		conv:                          conv,
	}
}

// ListAll missing godoc
func (r *pgRepository) ListAllByType(ctx context.Context, opType model.OperationType) ([]*model.Operation, error) {
	var entities EntityCollection

	err := r.globalLister.ListGlobal(ctx, &entities, repo.Conditions{repo.NewEqualCondition("op_type", opType)}...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
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

// GetByDataAndType retrieves an operation by data and type
func (r *pgRepository) GetByDataAndType(ctx context.Context, data interface{}, opType model.OperationType) (*model.Operation, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var entity Entity
	if err := r.globalSingleGetter.GetGlobal(ctx, repo.Conditions{repo.NewJSONCondition("data", string(dataBytes)), repo.NewEqualCondition("op_type", opType)}, repo.NoOrderBy, &entity); err != nil {
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
func (r *pgRepository) PriorityQueueListByType(ctx context.Context, queueLimit int, opType model.OperationType) ([]*model.Operation, error) {
	var entities EntityCollection
	if err := r.scheduledOperationsViewLister.ListGlobalWithLimit(ctx, &entities, queueLimit, repo.Conditions{repo.NewEqualCondition("op_type", opType)}...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities), nil
}

// LockOperation locks operations by given operation id
func (r *pgRepository) LockOperation(ctx context.Context, operationID string) (bool, error) {
	identifier, err := util.StringToInt64(operationID)
	if err != nil {
		return false, err
	}
	return r.globalFunctioner.AdvisoryLock(ctx, identifier)
}

// RescheduleOperations reschedules the operations.
func (r *pgRepository) RescheduleOperations(ctx context.Context, operationType model.OperationType, reschedulePeriod time.Duration) error {
	log.C(ctx).Debug("Rescheduling Operations")
	inCondition := repo.NewInConditionForStringValues("status", []string{"COMPLETED", "FAILED"})
	equalTypeCondition := repo.NewEqualCondition("op_type", string(operationType))
	dateCondition := repo.NewLessThanCondition("updated_at", time.Now().Add(-1*reschedulePeriod))
	return r.globalUpdater.UpdateFieldsGlobal(ctx, repo.Conditions{inCondition, equalTypeCondition, dateCondition}, map[string]interface{}{"status": "SCHEDULED", "updated_at": time.Now()})
}

// RescheduleHangedOperations reschedules operations that are hanged.
func (r *pgRepository) RescheduleHangedOperations(ctx context.Context, operationType model.OperationType, hangPeriod time.Duration) error {
	log.C(ctx).Debug("Rescheduling Operations")
	equalCondition := repo.NewEqualCondition("status", "IN_PROGRESS")
	equalTypeCondition := repo.NewEqualCondition("op_type", string(operationType))
	dateCondition := repo.NewLessThanCondition("updated_at", time.Now().Add(-1*hangPeriod))
	return r.globalUpdater.UpdateFieldsGlobal(ctx, repo.Conditions{equalCondition, equalTypeCondition, dateCondition}, map[string]interface{}{"status": "SCHEDULED", "updated_at": time.Now()})
}

func (r *pgRepository) multipleFromEntities(entities EntityCollection) []*model.Operation {
	items := make([]*model.Operation, 0, len(entities))
	for _, ent := range entities {
		model := r.conv.FromEntity(&ent)

		items = append(items, model)
	}
	return items
}
