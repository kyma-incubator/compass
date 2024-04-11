package application

import (
	"context"
	"fmt"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"strings"
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

const (
	applicationTable string = `public.applications`
	// listeningApplicationsView provides a structured view of applications that have a Webhook or their ApplicationTemplate has a Webhook
	listeningApplicationsView = `listening_applications`
)

var (
	applicationColumns    = []string{"id", "app_template_id", "system_number", "local_tenant_id", "name", "description", "status_condition", "status_timestamp", "system_status", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "application_namespace", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "tags", "documentation_labels"}
	updatableColumns      = []string{"name", "description", "status_condition", "status_timestamp", "system_status", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "application_namespace", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "tags", "documentation_labels", "system_number", "local_tenant_id"}
	upsertableColumns     = []string{"name", "description", "status_condition", "system_status", "provider_name", "base_url", "local_tenant_id", "application_namespace", "labels"}
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
	listeningAppsLister   repo.Lister
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
		listeningAppsLister:   repo.NewLister(listeningApplicationsView, applicationColumns),
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

// GetBySystemNumber returns an application retrieved by systemNumber from the Compass DB
func (r *pgRepository) GetBySystemNumber(ctx context.Context, tenant, systemNumber string) (*model.Application, error) {
	var appEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Application, tenant, repo.Conditions{repo.NewEqualCondition("system_number", systemNumber)}, repo.NoOrderBy, &appEnt); err != nil {
		return nil, err
	}

	appModel := r.conv.FromEntity(&appEnt)

	return appModel, nil
}

// ListByLocalTenantID returns applications with matching local tenant id and optionally - a filter
func (r *pgRepository) ListByLocalTenantID(ctx context.Context, tenant, localTenantID string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error) {
	var appsCollection EntityCollection
	conditions := repo.Conditions{repo.NewEqualCondition("local_tenant_id", localTenantID)}

	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	filterSubquery, args, err := label.FilterQuery(model.ApplicationLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

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

// GetByLocalTenantIDAndAppTemplateID returns the application with matching local tenant id and app template id from the Compass DB
func (r *pgRepository) GetByLocalTenantIDAndAppTemplateID(ctx context.Context, tenant, localTenantID, appTemplateID string) (*model.Application, error) {
	var appEnt Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition("local_tenant_id", localTenantID),
		repo.NewEqualCondition("app_template_id", appTemplateID),
	}
	if err := r.singleGetter.Get(ctx, resource.Application, tenant, conditions, repo.NoOrderBy, &appEnt); err != nil {
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

type EntityWithAppID struct {
	*tenant.Entity
	AppID string `db:"app_id"`
}

type EntityWithAppIDCollection []EntityWithAppID

// Len returns the current number of entities in the collection.
func (a EntityWithAppIDCollection) Len() int {
	return len(a)
}

func (r *pgRepository) ListAllGlobalByFilter(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationWithTenantsPage, error) {
	var entities EntityCollection

	filterSubquery, args, err := label.FilterQueryGlobal(model.ApplicationLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	log.C(ctx).Infof("Kalo filter subquery: %s", filterSubquery)

	var conditionTree *repo.ConditionTree
	if filterSubquery != "" {
		conditionTree = &repo.ConditionTree{Operand: repo.NewInConditionForSubQuery("id", filterSubquery, args)}
	}

	log.C(ctx).Infof("Kalo ListGlobalWithAdditionalConditions")
	page, totalCount, err := r.globalPageableQuerier.ListGlobalWithAdditionalConditions(ctx, pageSize, cursor, "id", &entities, conditionTree)
	if err != nil {
		return nil, err
	}

	// These are all filtered applications
	applications, err := r.multipleFromEntities(entities)
	if err != nil {
		return nil, err
	}

	filteredApplicationIDs := make([]string, 0, len(applications))
	for _, app := range applications {
		filteredApplicationIDs = append(filteredApplicationIDs, app.ID)
	}

	//--------------
	// Retrieve all customer/co tenants for the filtered apps

	if len(filteredApplicationIDs) == 0 {
		appWithTenantsData := make([]*model.ApplicationWithTenants, 0)
		return &model.ApplicationWithTenantsPage{
			Data:       appWithTenantsData,
			TotalCount: totalCount,
			PageInfo:   page}, nil
	}

	tntQuery := `SELECT DISTINCT ta.id as app_id,btm.id,btm.external_name,btm.external_tenant,btm.type,btm.provider_name,btm.status
FROM business_tenant_mappings btm
    JOIN tenant_applications ta ON btm.id = ta.tenant_id
WHERE ta.id in (%s) AND (btm.type = 'customer' or btm.type='cost-object')`

	//btm.type IN ('customer', 'cost-object')
	//SELECT DISTINCT ta_filtered.id AS app_id,
	//	btm.id,
	//	btm.external_name,
	//	btm.external_tenant,
	//	btm.type,
	//btm.provider_name,
	//	btm.status
	//FROM (
	//	SELECT id, tenant_id
	//FROM tenant_applications
	//WHERE id IN ('0bcaa9b6-797d-46f4-a444-4bd8abef4d51') -- Replace with actual IDs
	//) ta_filtered
	//JOIN business_tenant_mappings btm ON ta_filtered.tenant_id = btm.id
	//WHERE btm.type IN ('customer', 'cost-object');
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Executing DB query: %s", tntQuery)

	var entityCollection EntityWithAppIDCollection

	if err := persist.SelectContext(ctx, &entityCollection, fmt.Sprintf(tntQuery, arrayToString(filteredApplicationIDs))); err != nil {
		return nil, persistence.MapSQLError(ctx, err, resource.Tenant, resource.List, "while listing tenants")
	}

	for _, tntE := range entityCollection {
		log.C(ctx).Infof("Kalo found tenant: %s", tntE.ID)
	}

	tenantConverter := tnt.NewConverter()
	appWithTenantsData := make([]*model.ApplicationWithTenants, 0)
	for _, app := range applications {

		tntForApp := make([]*model.BusinessTenantMapping, 0)
		for _, e := range entityCollection {
			if e.AppID == app.ID {
				tntForApp = append(tntForApp, tenantConverter.FromEntity(e.Entity))
			}
		}

		appWithTnt := &model.ApplicationWithTenants{
			Application: *app,
			Tenants:     tntForApp,
		}

		appWithTenantsData = append(appWithTenantsData, appWithTnt)
	}

	return &model.ApplicationWithTenantsPage{
		Data:       appWithTenantsData,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

func arrayToString(arr []string) string {
	// Join array elements with single quotes and commas
	return "'" + strings.Join(arr, "','") + "'"
}

//// Without pages
//func (r *pgRepository) ListAllGlobalByFilter(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.Application, error) {
//	var entities EntityCollection
//
//	filterSubquery, args, err := label.FilterQueryGlobal(model.ApplicationLabelableObject, label.IntersectSet, filter)
//	if err != nil {
//		return nil, errors.Wrap(err, "while building filter query")
//	}
//
//	log.C(ctx).Infof("Kalo filter subquery: %s", filterSubquery)
//
//	var conditions repo.Conditions
//	//var conditionTree *repo.ConditionTree
//	if filterSubquery != "" {
//		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
//		//conditionTree = &repo.ConditionTree{Operand: repo.NewInConditionForSubQuery("id", filterSubquery, args)}
//	}
//
//	log.C(ctx).Infof("Kalo listing global")
//	if err = r.listerGlobal.ListGlobal(ctx, &entities, conditions...); err != nil {
//		return nil, err
//	}
//
//	r.globalPageableQuerier.ListGlobalWithAdditionalConditions(ctx, 200, "", "id", &entities, conditionTree)
//
//	return r.multipleFromEntities(entities)
//}

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

// ListAllByIDs lists all apps with given IDs
func (r *pgRepository) ListAllByIDs(ctx context.Context, tenantID string, ids []string) ([]*model.Application, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var entities EntityCollection
	err := r.lister.List(ctx, resource.Application, tenantID, &entities, repo.NewInConditionForStringValues("id", ids))

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListListeningApplications lists all application that either have webhook of type whType, or their application template has a webhook of type whType
func (r *pgRepository) ListListeningApplications(ctx context.Context, tenant string, whType model.WebhookType) ([]*model.Application, error) {
	var entities EntityCollection

	conditions := repo.Conditions{
		repo.NewEqualCondition("webhook_type", whType),
	}

	if err := r.listeningAppsLister.List(ctx, resource.Application, tenant, &entities, conditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
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
