package scenarioassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.automatic_scenario_assignments`

var columns = []string{scenarioColumn, tenantColumn, targetTenantColumn}

var (
	scenarioColumn     = "scenario"
	tenantColumn       = "tenant_id"
	targetTenantColumn = "target_tenant_id"
)

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:         repo.NewCreatorGlobal(resource.AutomaticScenarioAssigment, tableName, columns),
		lister:          repo.NewListerWithEmbeddedTenant(tableName, tenantColumn, columns),
		singleGetter:    repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, columns),
		pageableQuerier: repo.NewPageableQuerierWithEmbeddedTenant(tableName, tenantColumn, columns),
		deleter:         repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		conv:            conv,
	}
}

type repository struct {
	creator         repo.CreatorGlobal
	singleGetter    repo.SingleGetter
	lister          repo.Lister
	pageableQuerier repo.PageableQuerier
	deleter         repo.Deleter
	conv            EntityConverter
}

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(assignment model.AutomaticScenarioAssignment) Entity
	FromEntity(assignment Entity) model.AutomaticScenarioAssignment
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, model model.AutomaticScenarioAssignment) error {
	entity := r.conv.ToEntity(model)
	return r.creator.Create(ctx, entity)
}

func (r *repository) ListAll(ctx context.Context, tenantID string) ([]*model.AutomaticScenarioAssignment, error) {
	var out EntityCollection

	if err := r.lister.List(ctx, resource.AutomaticScenarioAssigment, tenantID, &out); err != nil {
		return nil, errors.Wrap(err, "while getting automatic scenario assignments from db")
	}

	items := make([]*model.AutomaticScenarioAssignment, 0, len(out))

	for _, v := range out {
		item := r.conv.FromEntity(v)
		items = append(items, &item)
	}

	return items, nil
}

func (r *repository) ListForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) ([]*model.AutomaticScenarioAssignment, error) {
	var out EntityCollection

	conditions := repo.Conditions{
		repo.NewEqualCondition(targetTenantColumn, targetTenantID),
	}

	if err := r.lister.List(ctx, resource.AutomaticScenarioAssigment, tenantID, &out, conditions...); err != nil {
		return nil, errors.Wrap(err, "while getting automatic scenario assignments from db")
	}

	items := make([]*model.AutomaticScenarioAssignment, 0, len(out))

	for _, v := range out {
		item := r.conv.FromEntity(v)
		items = append(items, &item)
	}

	return items, nil
}

// GetForScenarioName missing godoc
func (r *repository) GetForScenarioName(ctx context.Context, tenantID, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition(scenarioColumn, scenarioName),
	}

	if err := r.singleGetter.Get(ctx, resource.AutomaticScenarioAssigment, tenantID, conditions, repo.NoOrderBy, &ent); err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	assignmentModel := r.conv.FromEntity(ent)

	return assignmentModel, nil
}

// List missing godoc
func (r *repository) List(ctx context.Context, tenantID string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error) {
	var collection EntityCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, resource.AutomaticScenarioAssigment, tenantID, pageSize, cursor, scenarioColumn, &collection)
	if err != nil {
		return nil, err
	}

	items := make([]*model.AutomaticScenarioAssignment, 0, len(collection))

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

func (r *repository) DeleteForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) error {
	conditions := repo.Conditions{
		repo.NewEqualCondition(targetTenantColumn, targetTenantID),
	}

	return r.deleter.DeleteMany(ctx, resource.AutomaticScenarioAssigment, tenantID, conditions)
}

// DeleteForScenarioName missing godoc
func (r *repository) DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error {
	conditions := repo.Conditions{
		repo.NewEqualCondition(scenarioColumn, scenarioName),
	}

	return r.deleter.DeleteOne(ctx, resource.AutomaticScenarioAssigment, tenantID, conditions)
}
