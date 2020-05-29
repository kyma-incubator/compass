package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/pkg/errors"

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
		creator:         repo.NewCreator(resource.AutomaticScenarioAssigment, tableName, columns),
		lister:          repo.NewLister(resource.AutomaticScenarioAssigment, tableName, tenantColumn, columns),
		singleGetter:    repo.NewSingleGetter(resource.AutomaticScenarioAssigment, tableName, tenantColumn, columns),
		pageableQuerier: repo.NewPageableQuerier(resource.AutomaticScenarioAssigment, tableName, tenantColumn, columns),
		deleter:         repo.NewDeleter(resource.AutomaticScenarioAssigment, tableName, tenantColumn),
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

	conditions := repo.Conditions{
		repo.NewEqualCondition(selectorKeyColumn, in.Key),
		repo.NewEqualCondition(selectorValueColumn, in.Value),
	}

	if err := r.lister.List(ctx, tenantID, &out, conditions...); err != nil {
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
