package operation

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const operationTable = `public.operation`

var (
	operationTypeColumn = "op_type"
	statusColumn        = "status"
	finishedAtColumn    = "finished_at"
	operationColumns    = []string{"id", operationTypeColumn, statusColumn, "data", "error", "priority", "created_at", finishedAtColumn}
)

// EntityConverter missing godoc
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Operation) *Entity
	FromEntity(entity *Entity) *model.Operation
}

type pgRepository struct {
	globalCreator repo.CreatorGlobal
	globalDeleter repo.DeleterGlobal
	conv          EntityConverter
}

// NewRepository creates new operation repository
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		globalCreator: repo.NewCreatorGlobal(resource.Operation, operationTable, operationColumns),
		globalDeleter: repo.NewDeleterGlobal(resource.Operation, operationTable),
		conv:          conv,
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

// DeleteOlderThan deletes all operations of type `opType` with status `status` older than `date`
func (r *pgRepository) DeleteOlderThan(ctx context.Context, opType string, status model.OperationStatus, date time.Time) error {
	log.C(ctx).Infof("Deleting all operations of type %s with status %s older than %v", opType, status, date)
	return r.globalDeleter.DeleteManyGlobal(ctx, repo.Conditions{
		repo.NewNotNullCondition(finishedAtColumn),
		repo.NewEqualCondition(operationTypeColumn, opType),
		repo.NewEqualCondition(statusColumn, string(status)),
		repo.NewLessThanCondition(finishedAtColumn, date),
	})
}
