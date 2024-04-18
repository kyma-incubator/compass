package formationassignment

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tableName string = `public.formation_assignments`

var (
	idTableColumns        = []string{"id"}
	updatableTableColumns = []string{"finished_at_timestamp"}
	tableColumns          = []string{"id", "type", "formation_assignment_id", "formation_id", "triggered_by", "started_at_timestamp", "finished_at_timestamp"}
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
	pageableQuerier repo.PageableQuerierGlobal
	updater         repo.UpdaterGlobal
	conv            EntityConverter
}

// NewRepository creates a new FormationAssignment repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:         repo.NewCreatorGlobal(resource.AssignmentOperation, tableName, tableColumns),
		pageableQuerier: repo.NewPageableQuerierGlobal(resource.AssignmentOperation, tableName, tableColumns),
		updater:         repo.NewUpdaterGlobal(resource.AssignmentOperation, tableName, updatableTableColumns, idTableColumns),
		conv:            conv,
	}
}
