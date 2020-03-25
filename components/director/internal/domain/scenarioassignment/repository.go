package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.automatic_scenario_assignments`

var columns = []string{"scenario", tenantColumn, "selector_key", "selector_value"}
var tenantColumn = "tenant_id"

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator: repo.NewCreator(tableName, columns),
		conv:    conv,
	}
}

type repository struct {
	creator repo.Creator
	conv    EntityConverter
}

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(assignment model.AutomaticScenarioAssignment) Entity
}

func (r *repository) Create(ctx context.Context, model model.AutomaticScenarioAssignment) error {
	entity := r.conv.ToEntity(model)
	return r.creator.Create(ctx, entity)
}
