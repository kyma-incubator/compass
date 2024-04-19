package assignmentOperation

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"time"

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
	assignmentIDColumn    = "formation_assignment_id"
	idColumn              = "id"
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
	unionLister     repo.UnionListerGlobal
	pageableQuerier repo.PageableQuerierGlobal
	updater         repo.UpdaterGlobal
	conv            EntityConverter
}

// NewRepository creates a new FormationAssignment repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:         repo.NewCreatorGlobal(resource.AssignmentOperation, tableName, tableColumns),
		getter:          repo.NewSingleGetterGlobal(resource.AssignmentOperation, tableName, tableColumns),
		unionLister:     repo.NewUnionListerGlobal(resource.AssignmentOperation, tableName, tableColumns),
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
	// todo may have to add limit
	if err := r.getter.GetGlobal(ctx, conditions, repo.OrderByParams{repo.NewDescOrderBy(startedAtColumn)}, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// ListForFormationAssignmentIDs fetches the Assignment Operations for the provided Formation Assignment IDs
func (r *repository) ListForFormationAssignmentIDs(ctx context.Context, assignmentIDs []string, pageSize int, cursor string) ([]*model.AssignmentOperationPage, error) {
	var assignmentOperationsCollection EntityCollection

	orderByColumns := repo.OrderByParams{repo.NewAscOrderBy(assignmentIDColumn), repo.NewDescOrderBy(startedAtColumn), repo.NewDescOrderBy(idColumn)}

	counts, err := r.unionLister.ListGlobal(ctx, assignmentIDs, idColumn, pageSize, cursor, orderByColumns, &assignmentOperationsCollection)
	if err != nil {
		return nil, err
	}

	assignmentOperationsByAssignmentID := map[string][]*model.AssignmentOperation{}
	for _, aoEntity := range assignmentOperationsCollection {
		assignmentOperationsByAssignmentID[aoEntity.FormationAssignmentID] = append(assignmentOperationsByAssignmentID[aoEntity.FormationAssignmentID], r.conv.FromEntity(aoEntity))
	}

	offset, err := pagination.DecodeOffsetCursor(cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	aoPages := make([]*model.AssignmentOperationPage, 0, len(assignmentIDs))
	for _, assignmentID := range assignmentIDs {
		totalCount := counts[assignmentID]
		hasNextPage := false
		endCursor := ""
		if totalCount > offset+len(assignmentOperationsByAssignmentID[assignmentID]) {
			hasNextPage = true
			endCursor = pagination.EncodeNextOffsetCursor(offset, pageSize)
		}

		page := &pagination.Page{
			StartCursor: cursor,
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		}

		aoPages = append(aoPages, &model.AssignmentOperationPage{Data: assignmentOperationsByAssignmentID[assignmentID], TotalCount: totalCount, PageInfo: page})
	}

	return aoPages, nil
}

// Finish updates the finished at timestamp for the Assignment Operation with the provided ID
func (r *repository) Finish(ctx context.Context, m *model.AssignmentOperation) error {
	if m == nil {
		return apperrors.NewInternalError("model can not be empty")
	}
	newEntity := r.conv.ToEntity(m)

	log.C(ctx).Debugf("Updating the finished at timestamp for assignment operation with ID: %s", newEntity.ID)
	now := time.Now()
	newEntity.FinishedAtTimestamp = &now

	return r.updater.UpdateSingleGlobal(ctx, newEntity)
}

func (r *repository) multipleFromEntities(entities EntityCollection) []*model.AssignmentOperation {
	items := make([]*model.AssignmentOperation, 0, len(entities))
	for _, ent := range entities {
		items = append(items, r.conv.FromEntity(ent))
	}
	return items
}
