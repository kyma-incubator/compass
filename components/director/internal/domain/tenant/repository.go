package tenant

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantparentmapping"
	"k8s.io/utils/strings/slices"

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

	maxParameterChunkSize     int = 50000 // max parameters size in PostgreSQL is 65535
	getTenantsByParentAndType     = `SELECT %s from %s join %s on %s = %s where %s = ? and %s = ?`
)

var (
	idColumn                  = "id"
	idColumnCasted            = "id::text"
	externalNameColumn        = "external_name"
	externalTenantColumn      = "external_tenant"
	typeColumn                = "type"
	providerNameColumn        = "provider_name"
	statusColumn              = "status"
	initializedComputedColumn = "initialized"

	insertColumns      = []string{idColumn, externalNameColumn, externalTenantColumn, typeColumn, providerNameColumn, statusColumn}
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

	appTemplateIDColumn = "app_template_id"
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
	upserter                            repo.UpserterGlobal
	unsafeCreator                       repo.UnsafeCreator
	existQuerierGlobal                  repo.ExistQuerierGlobal
	existQuerierGlobalWithConditionTree repo.ExistQuerierGlobalWithConditionTree
	singleGetterGlobal                  repo.SingleGetterGlobal
	pageableQuerierGlobal               repo.PageableQuerierGlobal
	listerGlobal                        repo.ListerGlobal
	conditionTreeLister                 repo.ConditionTreeListerGlobal
	updaterGlobal                       repo.UpdaterGlobal
	deleterGlobal                       repo.DeleterGlobal

	tenantRuntimeContextQueryBuilder repo.QueryBuilderGlobal
	labelsQueryBuilder               repo.QueryBuilderGlobal
	applicationQueryBuilder          repo.QueryBuilderGlobal
	tenantApplicationsQueryBuilder   repo.QueryBuilderGlobal
	tenantParentRepo                 tenantparentmapping.TenantParentRepository

	conv Converter
}

// NewRepository returns a new entity responsible for repo-layer tenant operations. All of its methods require persistence.PersistenceOp it the provided context.
func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		upserter:                            repo.NewUpserterGlobal(resource.Tenant, tableName, insertColumns, conflictingColumns, updateColumns),
		unsafeCreator:                       repo.NewUnsafeCreator(resource.Tenant, tableName, insertColumns, conflictingColumns),
		existQuerierGlobal:                  repo.NewExistQuerierGlobal(resource.Tenant, tableName),
		existQuerierGlobalWithConditionTree: repo.NewExistsQuerierGlobalWithConditionTree(resource.Tenant, tableName),
		singleGetterGlobal:                  repo.NewSingleGetterGlobal(resource.Tenant, tableName, insertColumns),
		pageableQuerierGlobal:               repo.NewPageableQuerierGlobal(resource.Tenant, tableName, insertColumns),
		listerGlobal:                        repo.NewListerGlobal(resource.Tenant, tableName, insertColumns),
		conditionTreeLister:                 repo.NewConditionTreeListerGlobal(tableName, insertColumns),
		updaterGlobal:                       repo.NewUpdaterGlobal(resource.Tenant, tableName, []string{externalNameColumn, externalTenantColumn, typeColumn, providerNameColumn, statusColumn}, []string{idColumn}),
		deleterGlobal:                       repo.NewDeleterGlobal(resource.Tenant, tableName),
		tenantRuntimeContextQueryBuilder:    repo.NewQueryBuilderGlobal(resource.RuntimeContext, tenantRuntimeContextTable, tenantRuntimeContextSelectedColumns),
		labelsQueryBuilder:                  repo.NewQueryBuilderGlobal(resource.Label, labelsTable, labelsSelectedColumns),
		applicationQueryBuilder:             repo.NewQueryBuilderGlobal(resource.Application, applicationTable, applicationsSelectedColumns),
		tenantApplicationsQueryBuilder:      repo.NewQueryBuilderGlobal(resource.Application, tenantApplicationsTable, tenantApplicationsSelectedColumns),
		tenantParentRepo:                    tenantparentmapping.NewRepository(),
		conv:                                conv,
	}
}

// UnsafeCreate adds a new tenant in the Compass DB in case it does not exist. If it already exists, no action is taken.
// It is not guaranteed that the provided tenant ID is the same as the tenant ID in the database.
func (r *pgRepository) UnsafeCreate(ctx context.Context, item model.BusinessTenantMapping) (string, error) {
	if err := r.unsafeCreator.UnsafeCreate(ctx, r.conv.ToEntity(&item)); err != nil {
		return "", errors.Wrapf(err, "while creating business tenant mapping for tenant with external id %s", item.ExternalTenant)
	}
	btm, err := r.GetByExternalTenant(ctx, item.ExternalTenant)
	if err != nil {
		return "", errors.Wrapf(err, "while getting business tenant mapping by external id %s", item.ExternalTenant)
	}
	return btm.ID, r.tenantParentRepo.UpsertMultiple(ctx, btm.ID, item.Parents)
}

// Upsert adds the provided tenant into the Compass storage if it does not exist, or updates it if it does.
func (r *pgRepository) Upsert(ctx context.Context, item model.BusinessTenantMapping) (string, error) {
	if err := r.upserter.UpsertGlobal(ctx, r.conv.ToEntity(&item)); err != nil {
		return "", errors.Wrapf(err, "while upserting business tenant mapping for tenant with external id %s", item.ExternalTenant)
	}
	btm, err := r.GetByExternalTenant(ctx, item.ExternalTenant)
	if err != nil {
		return "", errors.Wrapf(err, "while getting business tenant mapping by external id %s", item.ExternalTenant)
	}
	return btm.ID, r.tenantParentRepo.UpsertMultiple(ctx, btm.ID, item.Parents)
}

// Get retrieves the active tenant with matching internal ID from the Compass storage.
func (r *pgRepository) Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	var entity tenant.Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(idColumn, id),
		repo.NewNotEqualCondition(statusColumn, string(tenant.Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, errors.Wrapf(err, "while getting tenant with id %s", id)
	}

	btm := r.conv.FromEntity(&entity)
	return r.enrichWithParents(ctx, btm)
}

// GetByExternalTenant retrieves the active tenant with matching external ID from the Compass storage.
func (r *pgRepository) GetByExternalTenant(ctx context.Context, externalTenantID string) (*model.BusinessTenantMapping, error) {
	var entity tenant.Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenantID),
		repo.NewNotEqualCondition(statusColumn, string(tenant.Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, errors.Wrapf(err, "while getting tenant with external id %s", externalTenantID)
	}

	btm := r.conv.FromEntity(&entity)
	return r.enrichWithParents(ctx, btm)
}

// Exists checks if tenant with the provided internal ID exists in the Compass storage.
func (r *pgRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// ExistsByExternalTenant checks if tenant with the provided external ID exists in the Compass storage.
func (r *pgRepository) ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(externalTenantColumn, externalTenant)})
}

// ExistsSubscribed checks if tenant is subscribed
func (r *pgRepository) ExistsSubscribed(ctx context.Context, id, selfDistinguishLabel string) (bool, error) {
	subaccountConditions := repo.Conditions{repo.NewEqualCondition(typeColumn, tenant.Subaccount)}

	tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs, err := r.tenantRuntimeContextQueryBuilder.BuildQueryGlobal(false, repo.Conditions{}...)
	if err != nil {
		return false, errors.Wrap(err, "while building query that fetches tenant from tenant_runtime_context")
	}

	applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs, err := r.labelsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewEqualCondition(keyColumn, selfDistinguishLabel), repo.NewNotNullCondition(appTemplateIDColumn)}...)
	if err != nil {
		return false, errors.Wrap(err, "while building query that fetches app_template_id from labels which have subscription")
	}

	applicationSubquery, applicationArgs, err := r.applicationQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(appTemplateIDColumn, applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs)}...)
	if err != nil {
		return false, errors.Wrap(err, "while building query that fetches application id from application table")
	}

	tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs, err := r.tenantApplicationsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(idColumn, applicationSubquery, applicationArgs)}...)
	if err != nil {
		return false, errors.Wrap(err, "while building query that fetches tenant id from tenant_applications table")
	}

	subscriptionConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(idColumn, tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs),
		repo.NewInConditionForSubQuery(idColumn, tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs),
	}

	conditions := repo.And(
		append(
			append(
				repo.ConditionTreesFromConditions(subaccountConditions),
				repo.Or(repo.ConditionTreesFromConditions(subscriptionConditions)...),
			),
			&repo.ConditionTree{Operand: repo.NewEqualCondition(idColumn, id)})...,
	)
	return r.existQuerierGlobalWithConditionTree.ExistsGlobalWithConditionTree(ctx, conditions)
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

	return r.enrichManyWithParents(ctx, entityCollection)
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
	for _, btm := range items {
		if _, err = r.enrichWithParents(ctx, btm); err != nil {
			return nil, err
		}
	}

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
	if len(externalTenant) == 0 {
		return []*model.BusinessTenantMapping{}, nil
	}

	conditions := repo.Conditions{
		repo.NewInConditionForStringValues(externalTenantColumn, externalTenant)}

	var entityCollection tenant.EntityCollection
	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, err
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

// Update updates the values of tenant with matching internal, and external IDs.
func (r *pgRepository) Update(ctx context.Context, model *model.BusinessTenantMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	tenantFromDB, err := r.Get(ctx, model.ID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant with ID %s", model.ID)
	}

	entity := r.conv.ToEntity(model)
	if err := r.updaterGlobal.UpdateSingleGlobal(ctx, entity); err != nil {
		return errors.Wrapf(err, "while updating tenant with ID %s", entity.ID)
	}

	btms, err := r.listByExternalTenantIDs(ctx, model.Parents)
	if err != nil {
		return errors.Wrapf(err, "while listing parent tenants by external IDs %v", model.Parents)
	}

	parentsInternalIDs := make([]string, 0, len(btms))
	for _, btm := range btms {
		parentsInternalIDs = append(parentsInternalIDs, btm.ID)
	}

	parentsToAdd := slices.Filter(nil, parentsInternalIDs, func(s string) bool {
		return !slices.Contains(tenantFromDB.Parents, s)
	})

	parentsToRemove := slices.Filter(nil, tenantFromDB.Parents, func(s string) bool {
		return !slices.Contains(parentsInternalIDs, s)
	})

	for _, p := range parentsToRemove {
		if err := r.removeParent(ctx, p, model.ID); err != nil {
			return err
		}
	}

	for _, p := range parentsToAdd {
		if err := r.addParent(ctx, p, model.ID); err != nil {
			return err
		}
	}

	return nil
}

func (r *pgRepository) addParent(ctx context.Context, internalParentID, internalChildID string) error {
	if err := r.tenantParentRepo.Upsert(ctx, internalChildID, internalParentID); err != nil {
		return errors.Wrapf(err, "while adding tenant parent record for tenant with ID %s and parent teannt with ID %s", internalChildID, internalParentID)
	}

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
		if err := tenantAccessLister.ListGlobal(ctx, &tenantAccesses, repo.NewEqualCondition(repo.M2MTenantIDColumn, internalChildID)); err != nil {
			return errors.Wrapf(err, "while listing tenant access records for tenant with id %s", internalChildID)
		}

		for _, ta := range tenantAccesses {
			tenantAccess := &repo.TenantAccess{
				TenantID:   internalParentID,
				ResourceID: ta.ResourceID,
				Owner:      ta.Owner,
				Source:     internalChildID,
			}
			if err := repo.CreateTenantAccessRecursively(ctx, m2mTable, tenantAccess); err != nil {
				return errors.Wrapf(err, "while creating tenant acccess record for resource %s for parent %s of tenant %s", ta.ResourceID, internalParentID, internalChildID)
			}
		}
	}
	return nil
}

func (r *pgRepository) removeParent(ctx context.Context, internalParentID, internalChildID string) error {
	if err := r.tenantParentRepo.Delete(ctx, internalChildID, internalParentID); err != nil {
		return errors.Wrapf(err, "while deleting tenant parent record for tenant with ID %s and parent tenant with ID %s", internalChildID, internalParentID)
	}

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
		if err := tenantAccessLister.ListGlobal(ctx, &tenantAccesses, repo.NewEqualCondition(repo.M2MTenantIDColumn, internalChildID)); err != nil {
			return errors.Wrapf(err, "while listing tenant access records for tenant with id %s", internalChildID)
		}

		resourceIDs := make([]string, 0, len(tenantAccesses))
		for _, ta := range tenantAccesses {
			resourceIDs = append(resourceIDs, ta.ResourceID)
		}

		if len(resourceIDs) > 0 {
			if err := repo.DeleteTenantAccessRecursively(ctx, m2mTable, internalParentID, resourceIDs, internalChildID); err != nil {
				return errors.Wrapf(err, "while deleting tenant accesses for the old parent %s of the tenant %s", internalParentID, internalChildID)
			}
		}

		if err := repo.DeleteTenantAccessFromParent(ctx, m2mTable, internalChildID, internalParentID); err != nil {
			return errors.Wrapf(err, "while deleting tenant accesses for granted from the old parent %s to tenant %s", internalParentID, internalChildID)
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
		return errors.Wrapf(err, "while getting tenant with external ID %s", externalTenant)
	}

	if tnt.Type != tenant.CostObject {
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

		if err = r.deleteChildTenantsRecursively(ctx, tnt.ID); err != nil {
			return err
		}
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenant),
	}

	return r.deleterGlobal.DeleteManyGlobal(ctx, conditions)
}

func (r *pgRepository) deleteChildTenantsRecursively(ctx context.Context, parentID string) error {
	childTenants, err := r.tenantParentRepo.ListByParent(ctx, parentID)
	if err != nil {
		return errors.Wrapf(err, "while listing child tenants for tenant with ID %s", parentID)
	}
	for _, childTenant := range childTenants {
		if err := r.deleteChildTenantsRecursively(ctx, childTenant); err != nil {
			return err
		}

		conditions := repo.Conditions{
			repo.NewEqualCondition(idColumn, childTenant),
		}
		if err = r.deleterGlobal.DeleteOneGlobal(ctx, conditions); err != nil {
			return errors.Wrapf(err, "while deleting tenant with ID %s", childTenant)
		}
	}
	return nil
}

// GetLowestOwnerForResource returns the lowest tenant in the hierarchy that is owner of a given resource.
func (r *pgRepository) GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error) {
	rawStmt := `(SELECT {{ .m2mTenantID }} FROM {{ .m2mTable }} ta WHERE ta.{{ .m2mID }} = ? AND ta.{{ .owner }} = true` +
		` AND (NOT EXISTS(SELECT 1 FROM {{ .tenantsTable }} JOIN {{ .tenantParentsTable }} ON {{ .tenantsTable }}.{{ .id }} = {{ .tenantParentsTable }}.{{ .tenantID }} WHERE {{ .parentID }} = ta.{{ .m2mTenantID }})` + // the tenant has no children
		` OR (NOT EXISTS(SELECT 1 FROM {{ .m2mTable }} ta2` +
		` WHERE ta2.{{ .m2mID }} = ? AND ta2.{{ .owner }} = true AND` +
		` ta2.{{ .m2mTenantID }} IN (SELECT {{ .id }} FROM {{ .tenantsTable }} JOIN {{ .tenantParentsTable }} ON {{ .tenantsTable }}.{{ .id }} = {{ .tenantParentsTable }}.{{ .tenantID }} WHERE {{ .parentID }} = ta.{{ .m2mTenantID }})))))` // there is no child that has owner access

	t, err := template.New("").Parse(rawStmt)
	if err != nil {
		return "", err
	}

	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return "", errors.Errorf("No tenant access table for %s", resourceType)
	}

	data := map[string]string{
		"m2mTenantID":        repo.M2MTenantIDColumn,
		"m2mTable":           m2mTable,
		"m2mID":              repo.M2MResourceIDColumn,
		"owner":              repo.M2MOwnerColumn,
		"tenantsTable":       tableName,
		"tenantParentsTable": "tenant_parents",
		"parentID":           "parent_id",
		"tenantID":           "tenant_id",
		"id":                 idColumn,
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

// GetParentsRecursivelyByExternalTenant gets the top parent for a given external tenant
func (r *pgRepository) GetParentsRecursivelyByExternalTenant(ctx context.Context, externalTenant string) ([]*model.BusinessTenantMapping, error) {
	recursiveQuery := `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.external_name, t1.external_tenant, t1.provider_name, t1.status, t1.type, tp1.parent_id, 0 AS depth
                    FROM business_tenant_mappings t1 JOIN tenant_parents tp1 on t1.id = tp1.tenant_id
                    WHERE external_tenant = $1
                    UNION ALL
                    SELECT t2.id, t2.external_name, t2.external_tenant, t2.provider_name, t2.status, t2.type, tp2.parent_id, p.depth+ 1
                    FROM business_tenant_mappings t2 LEFT JOIN tenant_parents tp2 on t2.id = tp2.tenant_id
                                                     INNER JOIN parents p on p.parent_id = t2.id)
			SELECT id, external_name, external_tenant, provider_name, status, type FROM parents WHERE parent_id is NULL AND (type != 'cost-object'
                                                                                              OR (type = 'cost-object' AND depth = (SELECT MIN(depth) FROM parents WHERE type = 'cost-object')))`

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Executing DB query: %s", recursiveQuery)

	var entityCollection tenant.EntityCollection

	if err := persist.SelectContext(ctx, &entityCollection, recursiveQuery, externalTenant); err != nil {
		return nil, persistence.MapSQLError(ctx, err, resource.Tenant, resource.List, "while listing parents for external tenant: %q", externalTenant)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

func (r *pgRepository) ListBySubscribedRuntimesAndApplicationTemplates(ctx context.Context, selfRegDistinguishLabel string) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	subaccountConditions := repo.Conditions{repo.NewEqualCondition(typeColumn, tenant.Subaccount)}

	tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs, err := r.tenantRuntimeContextQueryBuilder.BuildQueryGlobal(false, repo.Conditions{}...)
	if err != nil {
		return nil, errors.Wrap(err, "while building query that fetches tenant from tenant_runtime_context")
	}

	applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs, err := r.labelsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewEqualCondition(keyColumn, selfRegDistinguishLabel), repo.NewNotNullCondition(appTemplateIDColumn)}...)
	if err != nil {
		return nil, errors.Wrap(err, "while building query that fetches app_template_id from labels which have subscription")
	}

	applicationSubquery, applicationArgs, err := r.applicationQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(appTemplateIDColumn, applicationTemplateWithSubscriptionLabelSubquery, applicationTemplateWithSubscriptionLabelArgs)}...)
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
		return nil, errors.Wrap(err, "while listing tenants by label")
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

// ListByParentAndType list tenants by parent ID and tenant.Type
func (r *pgRepository) ListByParentAndType(ctx context.Context, parentID string, tenantType tenant.Type) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, err
	}

	resultColumns := make([]string, 0, len(insertColumns))
	for _, column := range insertColumns {
		resultColumns = append(resultColumns, prefixWithTableName(tableName, column))
	}

	tenantTypeColumn := prefixWithTableName(tableName, typeColumn)
	tenantsTableIDColumn := prefixWithTableName(tableName, idColumn)
	parentsTableTenantIDColumn := prefixWithTableName(tenantparentmapping.TenantParentsTable, tenantparentmapping.TenantIDColumn)
	parentsTableParentIDColumn := prefixWithTableName(tenantparentmapping.TenantParentsTable, tenantparentmapping.ParentIDColumn)

	deleteTenantAccessStmt := fmt.Sprintf(getTenantsByParentAndType, strings.Join(resultColumns, ", "), tableName, tenantparentmapping.TenantParentsTable, tenantsTableIDColumn, parentsTableTenantIDColumn, parentsTableParentIDColumn, tenantTypeColumn)
	deleteTenantAccessStmt = sqlx.Rebind(sqlx.DOLLAR, deleteTenantAccessStmt)

	log.C(ctx).Debugf("Executing DB query: %s", deleteTenantAccessStmt)

	if err = persist.SelectContext(ctx, &entityCollection, deleteTenantAccessStmt, parentID, tenantType); err != nil {
		return nil, errors.Wrapf(err, "while listing tenants of type %s with parent ID %s", tenantType, parentID)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

// ListByIds list tenants by ids
func (r *pgRepository) ListByIds(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewInConditionForStringValues(idColumn, ids),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, errors.Wrapf(err, "while listing tenants with ids %v", ids)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

// ListByType list tenants by tenant.Type
func (r *pgRepository) ListByType(ctx context.Context, tenantType tenant.Type) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewEqualCondition(typeColumn, tenantType),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, errors.Wrapf(err, "while listing tenants of type %s", tenantType)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

// ListByIdsAndType list tenants by ids that are of the specified type
func (r *pgRepository) ListByIdsAndType(ctx context.Context, ids []string, tenantType tenant.Type) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewInConditionForStringValues(idColumn, ids),
		repo.NewEqualCondition(typeColumn, tenantType),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, errors.Wrapf(err, "while listing tenants of type %s with ids %v", tenantType, ids)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

func (r *pgRepository) multipleFromEntities(entities tenant.EntityCollection) []*model.BusinessTenantMapping {
	items := make([]*model.BusinessTenantMapping, 0, len(entities))

	for _, entity := range entities {
		tmModel := r.conv.FromEntity(&entity)
		items = append(items, tmModel)
	}

	return items
}

func (r *pgRepository) enrichWithParents(ctx context.Context, btm *model.BusinessTenantMapping) (*model.BusinessTenantMapping, error) {
	parents, err := r.tenantParentRepo.ListParents(ctx, btm.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing parent tenants for tenant with ID %s", btm.ID)
	}

	btm.Parents = parents
	return btm, nil
}

func (r *pgRepository) enrichManyWithParents(ctx context.Context, entityCollection tenant.EntityCollection) ([]*model.BusinessTenantMapping, error) {
	btms := r.multipleFromEntities(entityCollection)
	for _, btm := range btms {
		if _, err := r.enrichWithParents(ctx, btm); err != nil {
			return nil, err
		}
	}
	return btms, nil
}

func prefixWithTableName(tableName, columnName string) string {
	return tableName + "." + columnName
}
