package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const applicationTable string = `public.applications`

var (
	applicationColumns = []string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id"}
	tenantColumn       = "tenant_id"
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Application) (*Entity, error)
	FromEntity(entity *Entity) *model.Application
}

type pgRepository struct {
	existQuerier    repo.ExistQuerier
	singleGetter    repo.SingleGetter
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator
	updater         repo.Updater
	conv            EntityConverter
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:    repo.NewExistQuerier(applicationTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(applicationTable, tenantColumn, applicationColumns),
		deleter:         repo.NewDeleter(applicationTable, tenantColumn),
		pageableQuerier: repo.NewPageableQuerier(applicationTable, tenantColumn, applicationColumns),
		creator:         repo.NewCreator(applicationTable, applicationColumns),
		updater:         repo.NewUpdater(applicationTable, []string{"name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id"}, tenantColumn, []string{"id"}),
		conv:            conv,
	}
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

func (r *pgRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}
	filterSubquery, err := label.FilterQuery(model.ApplicationLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	var additionalConditions []string
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, fmt.Sprintf(`"id" IN (%s)`, filterSubquery))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &appsCollection, additionalConditions...)

	if err != nil {
		return nil, err
	}

	var items []*model.Application

	for _, appEnt := range appsCollection {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}
	return &model.ApplicationPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

func (r *pgRepository) ListByScenarios(ctx context.Context, tenant uuid.UUID, scenarios []string, pageSize int, cursor string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection

	var scenariosFilers []*labelfilter.LabelFilter

	for _, scenarioValue := range scenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenarioValue)
		scenariosFilers = append(scenariosFilers, &labelfilter.LabelFilter{Key: model.ScenariosKey, Query: &query})
	}

	scenariosSubquery, err := label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenant, scenariosFilers)
	if err != nil {
		return nil, errors.Wrap(err, "while creating scenarios filter query")
	}

	var additionalConditions []string
	if scenariosSubquery != "" {
		additionalConditions = append(additionalConditions, fmt.Sprintf(`"id" IN (%s)`, scenariosSubquery))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant.String(), pageSize, cursor, "id", &appsCollection, additionalConditions...)

	if err != nil {
		return nil, err
	}

	var items []*model.Application

	for _, appEnt := range appsCollection {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}
	return &model.ApplicationPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

func (r *pgRepository) Create(ctx context.Context, model *model.Application) error {
	if model == nil {
		return errors.New("model can not be empty")
	}

	appEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Application entity")
	}

	return r.creator.Create(ctx, appEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Application) error {
	if model == nil {
		return errors.New("model can not be empty")
	}

	appEnt, err := r.conv.ToEntity(model)

	if err != nil {
		return errors.Wrap(err, "while converting to Application entity")
	}

	return r.updater.UpdateSingle(ctx, appEnt)
}
