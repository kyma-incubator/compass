package assignmentOperation

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tableName string = `public.formation_assignments`

var (
	idTableColumns        = []string{"id"}
	updatableTableColumns = []string{"finished_at_timestamp"}
	tableColumns          = []string{"id", "type", "formation_assignment_id", "formation_id", "triggered_by", "started_at_timestamp", "finished_at_timestamp"}
	startedAtColumn       = "started_at_timestamp"
)

// EntityConverter converts between the internal model and entity
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.AssignmentOperation) *Entity
	FromEntity(entity *Entity) *model.AssignmentOperation
}

type repository struct {
	creator         repo.CreatorGlobal
	getter          repo.SingleGetterGlobal
	pageableQuerier repo.PageableQuerierGlobal
	updater         repo.UpdaterGlobal
	conv            EntityConverter
}

// NewRepository creates a new FormationAssignment repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:         repo.NewCreatorGlobal(resource.AssignmentOperation, tableName, tableColumns),
		getter:          repo.NewSingleGetterGlobal(resource.AssignmentOperation, tableName, tableColumns),
		pageableQuerier: repo.NewPageableQuerierGlobal(resource.AssignmentOperation, tableName, tableColumns),
		updater:         repo.NewUpdaterGlobal(resource.AssignmentOperation, tableName, updatableTableColumns, idTableColumns),
		conv:            conv,
	}
}

// Create creates a new Assignment Operation in the database with the fields from the model
func (r *repository) Create(ctx context.Context, item *model.AssignmentOperation) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Persisting Assignment Operation entity with ID: %q", item.ID)
	return r.creator.Create(ctx, r.conv.ToEntity(item))
}

func (r *repository) GetLatestOperation(ctx context.Context, formationAssignmentID, formationID string, operationType model.AssignmentOperationType) (*model.AssignmentOperation, error) {
	var entity Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("formation_assignment_id", formationAssignmentID),
		repo.NewEqualCondition("formation_id", formationID),
		repo.NewEqualCondition("type", string(operationType)),
	}
	if err := r.getter.GetGlobal(ctx, conditions, repo.OrderByParams{repo.NewDescOrderBy(startedAtColumn)}, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}
