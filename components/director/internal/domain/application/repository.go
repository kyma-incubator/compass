package application

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"

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
	applicationColumns    = []string{"id", "app_template_id", "system_number", "local_tenant_id", "name", "description", "status_condition", "status_timestamp", "system_status", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "application_namespace", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "documentation_labels"}
	updatableColumns      = []string{"name", "description", "status_condition", "status_timestamp", "system_status", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "application_namespace", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "documentation_labels", "system_number", "local_tenant_id"}
	upsertableColumns     = []string{"name", "description", "status_condition", "system_status", "provider_name", "base_url", "application_namespace", "labels"}
	matchingSystemColumns = []string{"system_number"}
)

// EntityConverter missing godoc
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Application) (*Entity, error)
	FromEntity(entity *Entity) *model.Application
}

type pgRepository struct {
	existQuerier          repo.ExistQuerier
	ownerExistQuerier     repo.ExistQuerier
	singleGetter          repo.SingleGetter
	globalGetter          repo.SingleGetterGlobal
	globalDeleter         repo.DeleterGlobal
	lister                repo.Lister
	listerGlobal          repo.ListerGlobal
	deleter               repo.Deleter
	pageableQuerier       repo.PageableQuerier
	globalPageableQuerier repo.PageableQuerierGlobal
	creator               repo.Creator
	updater               repo.Updater
	globalUpdater         repo.UpdaterGlobal
	upserter              repo.Upserter
	trustedUpserter       repo.Upserter
	conv                  EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:          repo.NewExistQuerier(applicationTable),
		ownerExistQuerier:     repo.NewExistQuerierWithOwnerCheck(applicationTable),
		singleGetter:          repo.NewSingleGetter(applicationTable, applicationColumns),
		globalGetter:          repo.NewSingleGetterGlobal(resource.Application, applicationTable, applicationColumns),
		globalDeleter:         repo.NewDeleterGlobal(resource.Application, applicationTable),
		deleter:               repo.NewDeleter(applicationTable),
		lister:                repo.NewLister(applicationTable, applicationColumns),
		listerGlobal:          repo.NewListerGlobal(resource.Application, applicationTable, applicationColumns),
		pageableQuerier:       repo.NewPageableQuerier(applicationTable, applicationColumns),
		globalPageableQuerier: repo.NewPageableQuerierGlobal(resource.Application, applicationTable, applicationColumns),
		creator:               repo.NewCreator(applicationTable, applicationColumns),
		updater:               repo.NewUpdater(applicationTable, updatableColumns, []string{"id"}),
		globalUpdater:         repo.NewUpdaterGlobal(resource.Application, applicationTable, updatableColumns, []string{"id"}),
		upserter:              repo.NewUpserter(applicationTable, applicationColumns, matchingSystemColumns, upsertableColumns),
		trustedUpserter:       repo.NewTrustedUpserter(applicationTable, applicationColumns, matchingSystemColumns, upsertableColumns),
		conv:                  conv,
	}
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// OwnerExists checks if application with given id and tenant exists and has owner access
func (r *pgRepository) OwnerExists(ctx context.Context, tenant, id string) (bool, error) {
	return r.ownerExistQuerier.Exists(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	opMode := operation.ModeFromCtx(ctx)
	if opMode == graphql.OperationModeAsync {
		app, err := r.GetByID(ctx, tenant, id)
		if err != nil {
			return err
		}

		app.SetReady(false)
		app.SetError("")
		if app.GetDeletedAt().IsZero() { // Needed for the tests but might be useful for the production also
			app.SetDeletedAt(time.Now())
		}

		return r.Update(ctx, tenant, app)
	}

	return r.deleter.DeleteOne(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal missing godoc
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	opMode := operation.ModeFromCtx(ctx)
	if opMode == graphql.OperationModeAsync {
		app, err := r.GetGlobalByID(ctx, id)
		if err != nil {
			return err
		}

		app.SetReady(false)
		app.SetError("")
		if app.DeletedAt.IsZero() { // Needed for the tests but might be useful for the production also
			app.SetDeletedAt(time.Now())
		}

		return r.globalUpdater.UpdateSingleGlobal(ctx, app)
	}

	return r.globalDeleter.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// GetByIDForUpdate returns the application with matching ID from the Compass DB and locks it exclusively until the transaction is finished.
func (r *pgRepository) GetByIDForUpdate(ctx context.Context, tenant, id string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.GetForUpdate(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// GetBySystemNumber missing godoc
func (r *pgRepository) GetBySystemNumber(ctx context.Context, tenant, systemNumber string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("system_number", systemNumber)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// GetByNameAndSystemNumber missing godoc
func (r *pgRepository) GetByNameAndSystemNumber(ctx context.Context, tenant, name, systemNumber string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("name", name), repo.NewEqualCondition("system_number", systemNumber)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// GetGlobalByID missing godoc
func (r *pgRepository) GetGlobalByID(ctx context.Context, id string) (*model.Application, error) {
	var appEnt Entity
	if err := r.globalGetter.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// GetByFilter retrieves Application matching on the given label filters
func (r *pgRepository) GetByFilter(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Application, error) {
	var appEnt Entity

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

	if err = r.singleGetter.Get(ctx, resource.Application, tenant, conditions, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// ListAll missing godoc
func (r *pgRepository) ListAll(ctx context.Context, tenantID string) ([]*model.Application, error) {
	var entities EntityCollection

	err := r.lister.List(ctx, resource.Application, tenantID, &entities)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListAllByFilter retrieves all applications matching on the given label filters
func (r *pgRepository) ListAllByFilter(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Application, error) {
	var entities EntityCollection

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

	if err = r.lister.List(ctx, resource.Application, tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListAllByApplicationTemplateID retrieves all applications which have the given app template id
func (r *pgRepository) ListAllByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Application, error) {
	var appsCollection EntityCollection

	conditions := repo.Conditions{
		repo.NewEqualCondition("app_template_id", applicationTemplateID),
	}
	if err := r.listerGlobal.ListGlobal(ctx, &appsCollection, conditions...); err != nil {
		return nil, err
	}

	items := make([]*model.Application, 0, len(appsCollection))

	for _, appEnt := range appsCollection {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}

	return items, nil
}

// List missing godoc
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

	page, totalCount, err := r.pageableQuerier.List(ctx, resource.Application, tenant, pageSize, cursor, "id", &appsCollection, conditions...)

	if err != nil {
		return nil, err
	}

	items := make([]*model.Application, 0, len(appsCollection))

	for _, appEnt := range appsCollection {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}
	return &model.ApplicationPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

// ListGlobal missing godoc
func (r *pgRepository) ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection

	page, totalCount, err := r.globalPageableQuerier.ListGlobal(ctx, pageSize, cursor, "id", &appsCollection)

	if err != nil {
		return nil, err
	}

	items := make([]*model.Application, 0, len(appsCollection))

	for _, appEnt := range appsCollection {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}
	return &model.ApplicationPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

// ListByScenarios missing godoc
func (r *pgRepository) ListByScenarios(ctx context.Context, tenant uuid.UUID, scenarios []string, pageSize int, cursor string, hidingSelectors map[string][]string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection

	// Scenarios query part
	scenariosFilters := make([]*labelfilter.LabelFilter, 0, len(scenarios))
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

	page, totalCount, err := r.pageableQuerier.List(ctx, resource.Application, tenant.String(), pageSize, cursor, "id", &appsCollection, conditions...)

	if err != nil {
		return nil, err
	}

	items := make([]*model.Application, 0, len(appsCollection))

	for _, appEnt := range appsCollection {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}
	return &model.ApplicationPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

// ListByScenariosNoPaging lists all applications that are in any of the given scenarios
func (r *pgRepository) ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error) {
	tenantUUID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, apperrors.NewInvalidDataError("tenantID is not UUID")
	}

	var entities EntityCollection

	// Scenarios query part
	scenariosFilters := make([]*labelfilter.LabelFilter, 0, len(scenarios))
	for _, scenarioValue := range scenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenarioValue)
		scenariosFilters = append(scenariosFilters, labelfilter.NewForKeyWithQuery(model.ScenariosKey, query))
	}

	scenariosSubquery, scenariosArgs, err := label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenantUUID, scenariosFilters)
	if err != nil {
		return nil, errors.Wrap(err, "while creating scenarios filter query")
	}

	var conditions repo.Conditions
	if scenariosSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", scenariosSubquery, scenariosArgs))
	}

	if err = r.lister.List(ctx, resource.Application, tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	items := make([]*model.Application, 0, len(entities))

	for _, appEnt := range entities {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}

	return items, nil
}

// ListByScenariosAndIDs lists all apps with given IDs that are in any of the given scenarios
func (r *pgRepository) ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Application, error) {
	if len(scenarios) == 0 || len(ids) == 0 {
		return nil, nil
	}
	tenantUUID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, apperrors.NewInvalidDataError("tenantID is not UUID")
	}

	var entities EntityCollection

	// Scenarios query part
	scenariosFilters := make([]*labelfilter.LabelFilter, 0, len(scenarios))
	for _, scenarioValue := range scenarios {
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, scenarioValue)
		scenariosFilters = append(scenariosFilters, labelfilter.NewForKeyWithQuery(model.ScenariosKey, query))
	}

	scenariosSubquery, scenariosArgs, err := label.FilterQuery(model.ApplicationLabelableObject, label.UnionSet, tenantUUID, scenariosFilters)
	if err != nil {
		return nil, errors.Wrap(err, "while creating scenarios filter query")
	}

	var conditions repo.Conditions
	if scenariosSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", scenariosSubquery, scenariosArgs))
	}

	conditions = append(conditions, repo.NewInConditionForStringValues("id", ids))

	if err := r.lister.List(ctx, resource.Application, tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	items := make([]*model.Application, 0, len(entities))

	for _, appEnt := range entities {
		m := r.conv.FromEntity(&appEnt)
		items = append(items, m)
	}

	return items, nil
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Application) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Application model with id %s to entity", model.ID)
	appEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Application entity")
	}

	log.C(ctx).Debugf("Persisting Application entity with id %s to db", model.ID)
	return r.creator.Create(ctx, resource.Application, tenant, appEnt)
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Application) error {
	return r.updateSingle(ctx, tenant, model, false)
}

// Upsert inserts application for given tenant or update it if it already exists
func (r *pgRepository) Upsert(ctx context.Context, tenant string, model *model.Application) (string, error) {
	return r.genericUpsert(ctx, tenant, model, r.upserter)
}

// TrustedUpsert inserts application for given tenant or update it if it already exists ignoring tenant isolation
func (r *pgRepository) TrustedUpsert(ctx context.Context, tenant string, model *model.Application) (string, error) {
	return r.genericUpsert(ctx, tenant, model, r.trustedUpserter)
}

// TechnicalUpdate missing godoc
func (r *pgRepository) TechnicalUpdate(ctx context.Context, model *model.Application) error {
	return r.updateSingle(ctx, "", model, true)
}

func (r *pgRepository) genericUpsert(ctx context.Context, tenant string, model *model.Application, upserter repo.Upserter) (string, error) {
	if model == nil {
		return "", apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Application model with id %s to entity", model.ID)
	appEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return "", errors.Wrap(err, "while converting to Application entity")
	}

	log.C(ctx).Debugf("Upserting Application entity with id %s to db", model.ID)
	return upserter.Upsert(ctx, resource.Application, tenant, appEnt)
}

func (r *pgRepository) updateSingle(ctx context.Context, tenant string, model *model.Application, isTechnical bool) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Application model with id %s to entity", model.ID)
	appEnt, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrap(err, "while converting to Application entity")
	}

	log.C(ctx).Debugf("Persisting updated Application entity with id %s to db", model.ID)
	if isTechnical {
		return r.globalUpdater.TechnicalUpdate(ctx, appEnt)
	}
	return r.updater.UpdateSingle(ctx, resource.Application, tenant, appEnt)
}

func (r *pgRepository) multipleFromEntities(entities EntityCollection) ([]*model.Application, error) {
	items := make([]*model.Application, 0, len(entities))
	for _, ent := range entities {
		m := r.conv.FromEntity(&ent)
		items = append(items, m)
	}
	return items, nil
}
