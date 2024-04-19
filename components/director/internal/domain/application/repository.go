package application

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"text/template"
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
	tenantConv            TenantConverter
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
		tenantConv:            tnt.NewConverter(),
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

// ListAllGlobalByFilter lists a page of applications with their associated tenants filtered by the provided filters.
// Associated tenants are all tenants of type 'customer' or 'cost-object' that have access to the application.
func (r *pgRepository) ListAllGlobalByFilter(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationWithTenantsPage, error) {
	var entities EntityCollection

	filterSubquery, args, err := label.FilterQueryGlobal(model.ApplicationLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditionTree *repo.ConditionTree
	if filterSubquery != "" {
		conditionTree = &repo.ConditionTree{Operand: repo.NewInConditionForSubQuery("id", filterSubquery, args)}
	}

	page, totalCount, err := r.globalPageableQuerier.ListGlobalWithAdditionalConditions(ctx, pageSize, cursor, "id", &entities, conditionTree)
	if err != nil {
		return nil, err
	}

	filteredApplications, err := r.multipleFromEntities(entities)
	if err != nil {
		return nil, err
	}

	filteredApplicationIDs := make([]string, 0, len(filteredApplications))
	for _, app := range filteredApplications {
		filteredApplicationIDs = append(filteredApplicationIDs, app.ID)
	}

	if len(filteredApplicationIDs) == 0 {
		appWithTenantsData := make([]*model.ApplicationWithTenants, 0)
		return &model.ApplicationWithTenantsPage{
			Data:       appWithTenantsData,
			TotalCount: totalCount,
			PageInfo:   page,
		}, nil
	}

	entityWithAppIDCollection, err := r.listAssociatedTenants(ctx, filteredApplicationIDs)
	if err != nil {
		return nil, err
	}

	appToTenantsMap := make(map[string][]*model.BusinessTenantMapping)
	for _, e := range entityWithAppIDCollection {
		if _, ok := appToTenantsMap[e.AppID]; ok {
			appToTenantsMap[e.AppID] = append(appToTenantsMap[e.AppID], r.tenantConv.FromEntity(e.Entity))
		} else {
			appToTenantsMap[e.AppID] = []*model.BusinessTenantMapping{r.tenantConv.FromEntity(e.Entity)}
		}
	}

	applicationWithTenantsData := make([]*model.ApplicationWithTenants, 0, len(filteredApplications))
	for _, app := range filteredApplications {
		tenantsForApp := make([]*model.BusinessTenantMapping, 0)
		if _, ok := appToTenantsMap[app.ID]; ok {
			tenantsForApp = appToTenantsMap[app.ID]
		}

		applicationWithTenants := &model.ApplicationWithTenants{
			Application: *app,
			Tenants:     tenantsForApp,
		}

		applicationWithTenantsData = append(applicationWithTenantsData, applicationWithTenants)
	}

	return &model.ApplicationWithTenantsPage{
		Data:       applicationWithTenantsData,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// listAssociatedTenants retrieves all tenants of type 'cost-object' or 'customer' that have access to the provided applications
func (r *pgRepository) listAssociatedTenants(ctx context.Context, applicationIDs []string) (tenant.EntityWithAppIDCollection, error) {
	rawStmt := `SELECT DISTINCT ta_filtered.{{ .m2mID }} AS app_id,
				btm.{{ .btmID }},
				btm.{{ .btmExternalName }},
				btm.{{ .btmExternalTenant }},
				btm.{{ .btmType }},
				btm.{{ .btmProviderName }},
				btm.{{ .btmStatus }}
	FROM (	SELECT {{ .m2mID }}, {{ .m2mTenantID }}
			FROM {{ .m2mTable }}
			WHERE %s
	) ta_filtered
	JOIN {{ .btmTable }} btm ON ta_filtered.{{ .m2mTenantID }} = btm.{{ .btmID }}
	WHERE btm.{{ .btmType }} IN ('{{ .customerType }}', '{{ .costObjectType }}')`

	inConditionSubQuery := repo.NewInConditionForStringValues(repo.M2MResourceIDColumn, applicationIDs)
	inConditionArgs, _ := inConditionSubQuery.GetQueryArgs()
	rawStmt = fmt.Sprintf(rawStmt, inConditionSubQuery.GetQueryPart())

	t, err := template.New("").Parse(rawStmt)
	if err != nil {
		return nil, err
	}

	appM2MTable, _ := resource.Application.TenantAccessTable()
	data := map[string]string{
		"m2mTable":          appM2MTable,
		"m2mTenantID":       repo.M2MTenantIDColumn,
		"m2mID":             repo.M2MResourceIDColumn,
		"customerType":      tenant.TypeToStr(tenant.Customer),
		"costObjectType":    tenant.TypeToStr(tenant.CostObject),
		"btmTable":          tnt.TableName,
		"btmID":             tnt.IDColumn,
		"btmType":           tnt.TypeColumn,
		"btmStatus":         tnt.StatusColumn,
		"btmExternalName":   tnt.ExternalNameColumn,
		"btmExternalTenant": tnt.ExternalTenantColumn,
		"btmProviderName":   tnt.ProviderNameColumn,
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return nil, errors.Wrapf(err, "while executing template")
	}

	stmt := res.String()
	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Executing DB query: %s", stmt)

	var entityWithAppIDCollection tenant.EntityWithAppIDCollection
	if err := persist.SelectContext(ctx, &entityWithAppIDCollection, stmt, inConditionArgs...); err != nil {
		return nil, persistence.MapSQLError(ctx, err, resource.Tenant, resource.List, "while listing tenants")
	}

	return entityWithAppIDCollection, nil
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
