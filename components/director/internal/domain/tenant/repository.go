package tenant

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const (
	tableName                      string = `public.business_tenant_mappings`
	labelDefinitionsTableName      string = `public.label_definitions`
	labelDefinitionsTenantIDColumn string = `tenant_id`

	maxParameterChunkSize int = 50000 // max parameters size in PostgreSQL is 65535
)

var (
	idColumn                  = "id"
	idColumnCasted            = "id::text"
	externalNameColumn        = "external_name"
	externalTenantColumn      = "external_tenant"
	parentColumn              = "parent"
	typeColumn                = "type"
	providerNameColumn        = "provider_name"
	statusColumn              = "status"
	initializedComputedColumn = "initialized"

	insertColumns      = []string{idColumn, externalNameColumn, externalTenantColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}
	conflictingColumns = []string{externalTenantColumn}
	updateColumns      = []string{externalNameColumn}
	searchColumns      = []string{idColumnCasted, externalNameColumn, externalTenantColumn}

	tenantRuntimeContextTable           = "tenant_runtime_contexts"
	tenantRuntimeContextSelectedColumns = []string{"tenant_id"}
	labelsTable                         = "labels"
	labelsSelectedColumns               = []string{"app_template_id"}
	applicationTable                    = "applications"
	applicationsSelectedColumns         = []string{"id"}
	tenantApplicationsTable             = "tenant_applications"
	tenantApplicationsSelectedColumns   = []string{"tenant_id"}

	appTemplateIdColumn = "app_template_id"
	keyColumn           = "key"
)

// Converter converts tenants between the model.BusinessTenantMapping service-layer representation of a tenant and the repo-layer representation tenant.Entity.
//
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToEntity(in *model.BusinessTenantMapping) *tenant.Entity
	FromEntity(in *tenant.Entity) *model.BusinessTenantMapping
}

type pgRepository struct {
	upserter              repo.UpserterGlobal
	unsafeCreator         repo.UnsafeCreator
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	listerGlobal          repo.ListerGlobal
	conditionTreeLister   repo.ConditionTreeListerGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal

	tenantRuntimeContextQueryBuilder repo.QueryBuilderGlobal
	labelsQueryBuilder               repo.QueryBuilderGlobal
	applicationQueryBuilder          repo.QueryBuilderGlobal
	tenantApplicationsQueryBuilder   repo.QueryBuilderGlobal

	conv Converter
}

// NewRepository returns a new entity responsible for repo-layer tenant operations. All of its methods require persistence.PersistenceOp it the provided context.
func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		upserter:                         repo.NewUpserterGlobal(resource.Tenant, tableName, insertColumns, conflictingColumns, updateColumns),
		unsafeCreator:                    repo.NewUnsafeCreator(resource.Tenant, tableName, insertColumns, conflictingColumns),
		existQuerierGlobal:               repo.NewExistQuerierGlobal(resource.Tenant, tableName),
		singleGetterGlobal:               repo.NewSingleGetterGlobal(resource.Tenant, tableName, insertColumns),
		pageableQuerierGlobal:            repo.NewPageableQuerierGlobal(resource.Tenant, tableName, insertColumns),
		listerGlobal:                     repo.NewListerGlobal(resource.Tenant, tableName, insertColumns),
		conditionTreeLister:              repo.NewConditionTreeListerGlobal(tableName, insertColumns),
		updaterGlobal:                    repo.NewUpdaterGlobal(resource.Tenant, tableName, []string{externalNameColumn, externalTenantColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}, []string{idColumn}),
		deleterGlobal:                    repo.NewDeleterGlobal(resource.Tenant, tableName),
		tenantRuntimeContextQueryBuilder: repo.NewQueryBuilderGlobal(resource.RuntimeContext, tenantRuntimeContextTable, tenantRuntimeContextSelectedColumns),
		labelsQueryBuilder:               repo.NewQueryBuilderGlobal(resource.Label, labelsTable, labelsSelectedColumns),
		applicationQueryBuilder:          repo.NewQueryBuilderGlobal(resource.Application, applicationTable, applicationsSelectedColumns),
		tenantApplicationsQueryBuilder:   repo.NewQueryBuilderGlobal(resource.Application, tenantApplicationsTable, tenantApplicationsSelectedColumns),
		conv:                             conv,
	}
}

// UnsafeCreate adds a new tenant in the Compass DB in case it does not exist. If it already exists, no action is taken.
// It is not guaranteed that the provided tenant ID is the same as the tenant ID in the database.
func (r *pgRepository) UnsafeCreate(ctx context.Context, item model.BusinessTenantMapping) error {
	return r.unsafeCreator.UnsafeCreate(ctx, r.conv.ToEntity(&item))
}

// Upsert adds the provided tenant into the Compass storage if it does not exist, or updates it if it does.
func (r *pgRepository) Upsert(ctx context.Context, item model.BusinessTenantMapping) error {
	return r.upserter.UpsertGlobal(ctx, r.conv.ToEntity(&item))
}

// Get retrieves the active tenant with matching internal ID from the Compass storage.
func (r *pgRepository) Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	var entity tenant.Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(idColumn, id),
		repo.NewNotEqualCondition(statusColumn, string(tenant.Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// GetByExternalTenant retrieves the active tenant with matching external ID from the Compass storage.
func (r *pgRepository) GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error) {
	var entity tenant.Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenant),
		repo.NewNotEqualCondition(statusColumn, string(tenant.Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

// Exists checks if tenant with the provided internal ID exists in the Compass storage.
func (r *pgRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// ExistsByExternalTenant checks if tenant with the provided external ID exists in the Compass storage.
func (r *pgRepository) ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(externalTenantColumn, externalTenant)})
}

// List retrieves all tenants from the Compass storage.
func (r *pgRepository) List(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching persistence from context")
	}

	prefixedFields := strings.Join(str.PrefixStrings(insertColumns, "t."), ", ")
	query := fmt.Sprintf(`SELECT DISTINCT %s, ld.%s IS NOT NULL AS %s
			FROM %s t LEFT JOIN %s ld ON t.%s=ld.%s
			WHERE t.%s = $1
			ORDER BY %s DESC, t.%s ASC`, prefixedFields, labelDefinitionsTenantIDColumn, initializedComputedColumn, tableName, labelDefinitionsTableName, idColumn, labelDefinitionsTenantIDColumn, statusColumn, initializedComputedColumn, externalNameColumn)

	err = persist.SelectContext(ctx, &entityCollection, query, tenant.Active)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tenants from DB")
	}

	return r.multipleFromEntities(entityCollection), nil
}

// ListPageBySearchTerm retrieves a page of tenants from the Compass storage filtered by a search term.
func (r *pgRepository) ListPageBySearchTerm(ctx context.Context, searchTerm string, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error) {
	searchTermRegex := fmt.Sprintf("%%%s%%", searchTerm)

	var entityCollection tenant.EntityCollection
	likeConditions := make([]repo.Condition, 0, len(searchColumns))
	for _, searchColumn := range searchColumns {
		likeConditions = append(likeConditions, repo.NewLikeCondition(searchColumn, searchTermRegex))
	}

	conditions := repo.And(
		&repo.ConditionTree{Operand: repo.NewEqualCondition(statusColumn, tenant.Active)},
		repo.Or(repo.ConditionTreesFromConditions(likeConditions)...))

	page, totalCount, err := r.pageableQuerierGlobal.ListGlobalWithAdditionalConditions(ctx, pageSize, cursor, externalNameColumn, &entityCollection, conditions)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tenants from DB")
	}

	items := r.multipleFromEntities(entityCollection)

	return &model.BusinessTenantMappingPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListByExternalTenants retrieves all tenants with matching external ID from the Compass storage in chunks.
func (r *pgRepository) ListByExternalTenants(ctx context.Context, externalTenantIDs []string) ([]*model.BusinessTenantMapping, error) {
	tenants := make([]*model.BusinessTenantMapping, 0)

	for len(externalTenantIDs) > 0 {
		chunkSize := int(math.Min(float64(len(externalTenantIDs)), float64(maxParameterChunkSize)))
		tenantsChunk := externalTenantIDs[:chunkSize]
		tenantsFromDB, err := r.listByExternalTenantIDs(ctx, tenantsChunk)
		if err != nil {
			return nil, err
		}
		externalTenantIDs = externalTenantIDs[chunkSize:]
		tenants = append(tenants, tenantsFromDB...)
	}

	return tenants, nil
}

func (r *pgRepository) listByExternalTenantIDs(ctx context.Context, externalTenant []string) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewInConditionForStringValues(externalTenantColumn, externalTenant)}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection), nil
}

// Update updates the values of tenant with matching internal, and external IDs.
func (r *pgRepository) Update(ctx context.Context, model *model.BusinessTenantMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	tntFromDB, err := r.Get(ctx, model.ID)
	if err != nil {
		return err
	}

	entity := r.conv.ToEntity(model)

	if err := r.updaterGlobal.UpdateSingleGlobal(ctx, entity); err != nil {
		return err
	}

	if tntFromDB.Parent != model.Parent {
		for topLevelEntity := range resource.TopLevelEntities {
			if _, ok := topLevelEntity.IgnoredTenantAccessTable(); ok {
				log.C(ctx).Debugf("top level entity %s does not need a tenant access table", topLevelEntity)
				continue
			}

			m2mTable, ok := topLevelEntity.TenantAccessTable()
			if !ok {
				return errors.Errorf("top level entity %s does not have tenant access table", topLevelEntity)
			}

			tenantAccesses := repo.TenantAccessCollection{}

			tenantAccessLister := repo.NewListerGlobal(resource.TenantAccess, m2mTable, repo.M2MColumns)
			if err := tenantAccessLister.ListGlobal(ctx, &tenantAccesses, repo.NewEqualCondition(repo.M2MTenantIDColumn, model.ID), repo.NewEqualCondition(repo.M2MOwnerColumn, true)); err != nil {
				return errors.Wrapf(err, "while listing tenant access records for tenant with id %s", model.ID)
			}

			for _, ta := range tenantAccesses {
				tenantAccess := &repo.TenantAccess{
					TenantID:   model.Parent,
					ResourceID: ta.ResourceID,
					Owner:      true,
				}
				if err := repo.CreateTenantAccessRecursively(ctx, m2mTable, tenantAccess); err != nil {
					return errors.Wrapf(err, "while creating tenant acccess record for resource %s for parent %s of tenant %s", ta.ResourceID, model.Parent, model.ID)
				}
			}

			if len(tntFromDB.Parent) > 0 && len(tenantAccesses) > 0 {
				resourceIDs := make([]string, 0, len(tenantAccesses))
				for _, ta := range tenantAccesses {
					resourceIDs = append(resourceIDs, ta.ResourceID)
				}

				if err := repo.DeleteTenantAccessRecursively(ctx, m2mTable, tntFromDB.Parent, resourceIDs); err != nil {
					return errors.Wrapf(err, "while deleting tenant accesses for the old parent %s of the tenant %s", tntFromDB.Parent, model.ID)
				}
			}
		}
	}

	return nil
}

// DeleteByExternalTenant removes a tenant with matching external ID from the Compass storage.
// It also deletes all the accesses for resources that the tenant is owning for its parents.
func (r *pgRepository) DeleteByExternalTenant(ctx context.Context, externalTenant string) error {
	tnt, err := r.GetByExternalTenant(ctx, externalTenant)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil
		}
		return err
	}

	for topLevelEntity, topLevelEntityTable := range resource.TopLevelEntities {
		if _, ok := topLevelEntity.IgnoredTenantAccessTable(); ok {
			log.C(ctx).Debugf("top level entity %s does not need a tenant access table", topLevelEntity)
			continue
		}

		m2mTable, ok := topLevelEntity.TenantAccessTable()
		if !ok {
			return errors.Errorf("top level entity %s does not have tenant access table", topLevelEntity)
		}

		tenantAccesses := repo.TenantAccessCollection{}

		tenantAccessLister := repo.NewListerGlobal(resource.TenantAccess, m2mTable, repo.M2MColumns)
		if err := tenantAccessLister.ListGlobal(ctx, &tenantAccesses, repo.NewEqualCondition(repo.M2MTenantIDColumn, tnt.ID), repo.NewEqualCondition(repo.M2MOwnerColumn, true)); err != nil {
			return errors.Wrapf(err, "while listing tenant access records for tenant with id %s", tnt.ID)
		}

		if len(tenantAccesses) > 0 {
			resourceIDs := make([]string, 0, len(tenantAccesses))
			for _, ta := range tenantAccesses {
				resourceIDs = append(resourceIDs, ta.ResourceID)
			}

			deleter := repo.NewDeleterGlobal(topLevelEntity, topLevelEntityTable)
			if err := deleter.DeleteManyGlobal(ctx, repo.Conditions{repo.NewInConditionForStringValues("id", resourceIDs)}); err != nil {
				return errors.Wrapf(err, "while deleting resources owned by tenant %s", tnt.ID)
			}
		}
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenant),
	}

	return r.deleterGlobal.DeleteManyGlobal(ctx, conditions)
}

// GetLowestOwnerForResource returns the lowest tenant in the hierarchy that is owner of a given resource.
func (r *pgRepository) GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error) {
	rawStmt := `(SELECT {{ .m2mTenantID }} FROM {{ .m2mTable }} ta WHERE ta.{{ .m2mID }} = ? AND ta.{{ .owner }} = true` +
		` AND (NOT EXISTS(SELECT 1 FROM {{ .tenantsTable }} WHERE {{ .parent }} = ta.{{ .m2mTenantID }})` + // the tenant has no children
		` OR (NOT EXISTS(SELECT 1 FROM {{ .m2mTable }} ta2` +
		` WHERE ta2.{{ .m2mID }} = ? AND ta2.{{ .owner }} = true AND` +
		` ta2.{{ .m2mTenantID }} IN (SELECT {{ .id }} FROM {{ .tenantsTable }} WHERE {{ .parent }} = ta.{{ .m2mTenantID }})))))` // there is no child that has owner access

	t, err := template.New("").Parse(rawStmt)
	if err != nil {
		return "", err
	}

	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return "", errors.Errorf("No tenant access table for %s", resourceType)
	}

	data := map[string]string{
		"m2mTenantID":  repo.M2MTenantIDColumn,
		"m2mTable":     m2mTable,
		"m2mID":        repo.M2MResourceIDColumn,
		"owner":        repo.M2MOwnerColumn,
		"tenantsTable": tableName,
		"parent":       parentColumn,
		"id":           idColumn,
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return "", errors.Wrapf(err, "while executing template")
	}

	stmt := res.String()
	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return "", err
	}

	log.C(ctx).Debugf("Executing DB query: %s", stmt)

	dest := struct {
		TenantID string `db:"tenant_id"`
	}{}

	if err := persist.GetContext(ctx, &dest, stmt, objectID, objectID); err != nil {
		return "", persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Get, "while getting lowest tenant from %s table for resource %s with id %s", m2mTable, resourceType, objectID)
	}

	return dest.TenantID, nil
}

// GetCustomerIDParentRecursively gets the top parent external ID (customer_id) for a given tenant
func (r *pgRepository) GetCustomerIDParentRecursively(ctx context.Context, tenantID string) (string, error) {
	recursiveQuery := `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.parent, t1.external_tenant, t1.type
                    FROM business_tenant_mappings t1
                    WHERE id = $1
                    UNION ALL
                    SELECT t2.id, t2.parent, t2.external_tenant, t2.type
                    FROM business_tenant_mappings t2
                             INNER JOIN parents p on p.parent = t2.id)
			SELECT external_tenant, type FROM parents WHERE parent is null`

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return "", err
	}

	log.C(ctx).Debugf("Executing DB query: %s", recursiveQuery)

	dest := struct {
		ExternalCustomerTenant string `db:"external_tenant"`
		Type                   string `db:"type"`
	}{}

	if err := persist.GetContext(ctx, &dest, recursiveQuery, tenantID); err != nil {
		return "", persistence.MapSQLError(ctx, err, resource.Tenant, resource.Get, "while getting parent external customer ID for internal tenant: %q", tenantID)
	}

	if dest.Type != tenant.TypeToStr(tenant.Customer) {
		return "", nil
	}

	if dest.ExternalCustomerTenant == "" {
		return "", errors.Errorf("external parent customer ID for internal tenant ID: %s can not be empty", tenantID)
	}

	return dest.ExternalCustomerTenant, nil
}

func (r *pgRepository) ListBySubscribedRuntimesAndApplicationTemplates(ctx context.Context, selfRegDistinguishLabel string) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	subaccountConditions := repo.Conditions{repo.NewEqualCondition(typeColumn, tenant.Subaccount)}

	tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs, err := r.tenantRuntimeContextQueryBuilder.BuildQueryGlobal(false, repo.Conditions{}...)
	if err != nil {
		return nil, errors.Wrap(err, "while building query that fetches tenant from tenant_runtime_context")
	}

	applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs, err := r.labelsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewEqualCondition(keyColumn, selfRegDistinguishLabel), repo.NewNotNullCondition(appTemplateIdColumn)}...)
	if err != nil {
		return nil, errors.Wrap(err, "while building query that fetches app_template_id from labels which have subscription")
	}

	applicationSubquery, applicationArgs, err := r.applicationQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(appTemplateIdColumn, applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs)}...)
	if err != nil {
		return nil, errors.Wrap(err, "while building query that fetches application id from application table")
	}

	tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs, err := r.tenantApplicationsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(idColumn, applicationSubquery, applicationArgs)}...)
	if err != nil {
		return nil, errors.Wrap(err, "while building query that fetches tenant id from tenant_applications table")
	}

	subscriptionConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs),
		repo.NewInConditionForSubQuery(idColumn, tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs),
	}

	conditions := repo.And(
		append(
			repo.ConditionTreesFromConditions(subaccountConditions),
			repo.Or(repo.ConditionTreesFromConditions(subscriptionConditions)...),
		)...,
	)

	if err := r.conditionTreeLister.ListConditionTreeGlobal(ctx, resource.Tenant, &entityCollection, conditions); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection), nil
}

// ListByParentAndType list tenants by parent ID and tenant.Type
func (r *pgRepository) ListByParentAndType(ctx context.Context, parentID string, tenantType tenant.Type) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewEqualCondition(parentColumn, parentID),
		repo.NewEqualCondition(typeColumn, tenantType),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection), nil
}

// ListByType list tenants by tenant.Type
func (r *pgRepository) ListByType(ctx context.Context, tenantType tenant.Type) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewEqualCondition(typeColumn, tenantType),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entityCollection), nil
}

func (r *pgRepository) multipleFromEntities(entities tenant.EntityCollection) []*model.BusinessTenantMapping {
	items := make([]*model.BusinessTenantMapping, 0, len(entities))

	for _, entity := range entities {
		tmModel := r.conv.FromEntity(&entity)
		items = append(items, tmModel)
	}

	return items
}
