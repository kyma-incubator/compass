package application

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

const applicationTable string = `public.applications`

var (
	applicationColumns = []string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error"}
	updatableColumns   = []string{"name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error"}
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
	lister          repo.Lister
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator
	updater         repo.Updater
	conv            EntityConverter
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:    repo.NewExistQuerier(resource.Application, applicationTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(resource.Application, applicationTable, tenantColumn, applicationColumns),
		deleter:         repo.NewDeleter(resource.Application, applicationTable, tenantColumn),
		lister:          repo.NewLister(resource.Application, applicationTable, tenantColumn, applicationColumns),
		pageableQuerier: repo.NewPageableQuerier(resource.Application, applicationTable, tenantColumn, applicationColumns),
		creator:         repo.NewCreator(resource.Application, applicationTable, applicationColumns),
		updater:         repo.NewUpdater(resource.Application, applicationTable, updatableColumns, tenantColumn, []string{"id"}),
		conv:            conv,
	}
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	opMode := operation.ModeFromCtx(ctx)
	if opMode == graphql.OperationModeAsync {
		app, err := r.GetByID(ctx, tenant, id)
		if err != nil {
			return err
		}

		app.Ready = false
		if app.GetDeletedAt().IsZero() { // Needed for the tests but might be useful for the production also
			app.SetDeletedAt(time.Now())
		}

		return r.Update(ctx, app)
	}

	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

func (r *pgRepository) ListAll(ctx context.Context, tenantID string) ([]*model.Application, error) {
	var entities EntityCollection

	err := r.lister.List(ctx, tenantID, &entities)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

func (r *pgRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}
	filterSubquery, args, err := label.FilterQuery(model.ApplicationLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &appsCollection, conditions...)

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

func (r *pgRepository) ListByScenarios(ctx context.Context, tenant uuid.UUID, scenarios []string, pageSize int, cursor string, hidingSelectors map[string][]string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection

	// Scenarios query part
	var scenariosFilters []*labelfilter.LabelFilter
	for _, scenarioValue := range scenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenarioValue)
		scenariosFilters = append(scenariosFilters, labelfilter.NewForKeyWithQuery(model.ScenariosKey, query))
	}

	scenariosSubquery, scenariosArgs, err := label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenant, scenariosFilters)
	if err != nil {
		return nil, errors.Wrap(err, "while creating scenarios filter query")
	}

	// Application Hide query part
	var appHideFilters []*labelfilter.LabelFilter
	for key, values := range hidingSelectors {
		for _, value := range values {
			appHideFilters = append(appHideFilters, labelfilter.NewForKeyWithQuery(key, fmt.Sprintf(`"%s"`, value)))
		}
	}

	appHideSubquery, appHideArgs, err := label.FilterSubquery(model.ApplicationLabelableObject, label.ExceptSet, tenant, appHideFilters)
	if err != nil {
		return nil, errors.Wrap(err, "while creating scenarios filter query")
	}

	// Combining both queries
	combinedQuery := scenariosSubquery + appHideSubquery
	combinedArgs := append(scenariosArgs, appHideArgs...)

	var conditions repo.Conditions
	if combinedQuery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", combinedQuery, combinedArgs))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant.String(), pageSize, cursor, "id", &appsCollection, conditions...)

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
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Application model with id %s to entity", model.ID)
	appEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Application entity")
	}

	log.C(ctx).Debugf("Persisting Application entity with id %s to db", model.ID)
	return r.creator.Create(ctx, appEnt)
}

func (r *pgRepository) Update(ctx context.Context, model *model.Application) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Application model with id %s to entity", model.ID)
	appEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Application entity")
	}

	log.C(ctx).Debugf("Persisting updated Application entity with id %s to db", model.ID)
	return r.updater.UpdateSingle(ctx, appEnt)
}

func (r *pgRepository) multipleFromEntities(entities EntityCollection) ([]*model.Application, error) {
	var items []*model.Application
	for _, ent := range entities {
		m := r.conv.FromEntity(&ent)
		items = append(items, m)
	}
	return items, nil
}
