package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.automatic_scenario_assignments`

var columns = []string{"scenario", tenantColumn, selectorKeyColumn, selectorValueColumn}

var (
	tenantColumn        = "tenant_id"
	selectorKeyColumn   = "selector_key"
	selectorValueColumn = "selector_value"
	scenarioColumn      = "scenario"
)

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:      repo.NewCreator(tableName, columns),
		lister:       repo.NewLister(tableName, tenantColumn, columns),
		singleGetter: repo.NewSingleGetter(tableName, tenantColumn, columns),
		conv:         conv,
	}
}

type repository struct {
	creator      repo.Creator
	singleGetter repo.SingleGetter
	lister       repo.Lister
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

func (r *repository) GetForSelector(ctx context.Context, in model.LabelSelector, tenantID string) ([]*model.AutomaticScenarioAssignment, error) {
	var out EntityCollection

	conditionKey := fmt.Sprintf("%s = %s", selectorKeyColumn, pq.QuoteLiteral(in.Key))
	conditionValue := fmt.Sprintf("%s = %s", selectorValueColumn, pq.QuoteLiteral(in.Value))

	if err := r.lister.List(ctx, tenantID, &out, conditionKey, conditionValue); err != nil {
		return nil, errors.Wrap(err, "while getting automatic scenario assignments from db")
	}

	var items []*model.AutomaticScenarioAssignment

	for _, v := range out {
		item := r.conv.FromEntity(v)
		items = append(items, &item)
	}

	return items, nil
}
func (r *repository) GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition(scenarioColumn, scenarioName),
	}

	if err := r.singleGetter.Get(ctx, tenantID, conditions, repo.NoOrderBy, &ent); err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	assignmentModel := r.conv.FromEntity(ent)

	return assignmentModel, nil
}
