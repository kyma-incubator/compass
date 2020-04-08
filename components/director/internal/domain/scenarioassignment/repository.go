package scenarioassignment

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.automatic_scenario_assignments`

var columns = []string{scenarioColumn, tenantColumn, selectorKeyColumn, selectorValueColumn}

var (
	tenantColumn        = "tenant_id"
	selectorKeyColumn   = "selector_key"
	selectorValueColumn = "selector_value"
	scenarioColumn      = "scenario"
)

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:         repo.NewCreator(tableName, columns),
		lister:          repo.NewLister(tableName, tenantColumn, columns),
		singleGetter:    repo.NewSingleGetter(tableName, tenantColumn, columns),
		pageableQuerier: repo.NewPageableQuerier(tableName, tenantColumn, columns),
		deleter:         repo.NewDeleter(tableName, tenantColumn),
		conv:            conv,
	}
}

type repository struct {
	creator         repo.Creator
	singleGetter    repo.SingleGetter
	lister          repo.Lister
	pageableQuerier repo.PageableQuerier
	deleter         repo.Deleter
	conv            EntityConverter
}

const updateQuery = `UPDATE labels AS l SET value=SCENARIOS.SCENARIOS 
		FROM (SELECT array_to_json(array_agg(scenario)) AS SCENARIOS FROM automatic_scenario_assignments 
					WHERE selector_key=$1 AND selector_value=$2 AND tenant_id=$3) AS SCENARIOS
		WHERE l.runtime_id IN (SELECT runtime_id FROM labels  
									WHERE key =$1 AND value ?| array[$2] AND runtime_id IS NOT NULL AND tenant_ID=$3) 
			AND l.key ='scenarios'
			AND l.tenant_id=$3`

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(assignment model.AutomaticScenarioAssignment) Entity
	FromEntity(assignment Entity) model.AutomaticScenarioAssignment
}

func (r *repository) Create(ctx context.Context, model model.AutomaticScenarioAssignment) error {
	entity := r.conv.ToEntity(model)
	return r.creator.Create(ctx, entity)
}

func (r *repository) ListForSelector(ctx context.Context, in model.LabelSelector, tenantID string) ([]*model.AutomaticScenarioAssignment, error) {
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

func (r *repository) List(ctx context.Context, tenantID string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error) {
	var collection EntityCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, tenantID, pageSize, cursor, scenarioColumn, &collection)
	if err != nil {
		return nil, err
	}

	var items []*model.AutomaticScenarioAssignment

	for _, ent := range collection {
		m := r.conv.FromEntity(ent)
		items = append(items, &m)
	}

	return &model.AutomaticScenarioAssignmentPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *repository) DeleteForSelector(ctx context.Context, tenantID string, selector model.LabelSelector) error {
	conditions := repo.Conditions{
		repo.NewEqualCondition(selectorKeyColumn, selector.Key),
		repo.NewEqualCondition(selectorValueColumn, selector.Value),
	}

	return r.deleter.DeleteMany(ctx, tenantID, conditions)
}

func (r *repository) DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error {
	conditions := repo.Conditions{
		repo.NewEqualCondition(scenarioColumn, scenarioName),
	}

	return r.deleter.DeleteOne(ctx, tenantID, conditions)
}

func (r *repository) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while getting persitance from context")
	}

	_, err = persist.Exec(updateQuery, in.Selector.Key, in.Selector.Value, in.Tenant)
	if err != nil {
		return errors.Wrap(err, "while updating scenarios")
	}
	return nil
}
