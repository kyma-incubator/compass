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

var (
	operationColumns = []string{"id", "op_type", "status", "data", "error", "priority", "created_at", "updated_at"}
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
