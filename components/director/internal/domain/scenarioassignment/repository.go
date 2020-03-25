package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.automatic_scenario_assignments`

var columns = []string{scenarioColumn, tenantColumn, "selector_key", "selector_value"}
var tenantColumn = "tenant_id"
var scenarioColumn = "scenario"

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:      repo.NewCreator(tableName, columns),
		singleGetter: repo.NewSingleGetter(tableName, tenantColumn, columns),
		conv:         conv,
	}
}

type repository struct {
	creator      repo.Creator
	singleGetter repo.SingleGetter
	conv         EntityConverter
}

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(assignment model.AutomaticScenarioAssignment) Entity
	FromEntity(assignment Entity) model.AutomaticScenarioAssignment
}

func (r *repository) Create(ctx context.Context, model model.AutomaticScenarioAssignment) error {
	entity := r.conv.ToEntity(model)
	return r.creator.Create(ctx, entity)
}

func (r *repository) GetByScenarioName(ctx context.Context, tnt, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition(scenarioColumn, scenarioName),
	}

	if err := r.singleGetter.Get(ctx, tnt, conditions, repo.NoOrderBy, &ent); err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	model := r.conv.FromEntity(ent)

	return model, nil
}
