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
	// TableName specifies the name of the business tenant mappings table
	TableName string = `public.business_tenant_mappings`
	// IDColumn specifies the name of the id column in business tenant mappings table
	IDColumn string = "id"
	// ExternalNameColumn specifies the name of the external name column in business tenant mappings table
	ExternalNameColumn string = "external_name"
	// ExternalTenantColumn specifies the name of the external tenant column in business tenant mappings table
	ExternalTenantColumn string = "external_tenant"
	// TypeColumn specifies the name of the type column in business tenant mappings table
	TypeColumn string = "type"
	// ProviderNameColumn specifies the name of the provider column in business tenant mappings table
	ProviderNameColumn string = "provider_name"
	// StatusColumn specifies the name of the status column in business tenant mappings table
	StatusColumn string = "status"

	labelDefinitionsTableName string = `public.label_definitions`
	tenantIDColumn            string = `tenant_id`
	idColumnCasted            string = "id::text"
	initializedComputedColumn string = "initialized"
	tenantRuntimeContextTable string = "tenant_runtime_contexts"
	labelsTable               string = "labels"
	applicationTable          string = "applications"
	tenantApplicationsTable   string = "tenant_applications"
	appTemplateIDColumn       string = "app_template_id"
	keyColumn                 string = "key"

	maxParameterChunkSize     int = 50000 // max parameters size in PostgreSQL is 65535
	getTenantsByParentAndType     = `SELECT %s from %s join %s on %s = %s where %s = ? and %s = ?`
)

var (
	insertColumns      = []string{IDColumn, ExternalNameColumn, ExternalTenantColumn, TypeColumn, ProviderNameColumn, StatusColumn}
	conflictingColumns = []string{ExternalTenantColumn}
	updateColumns      = []string{ExternalNameColumn}
	searchColumns      = []string{idColumnCasted, ExternalNameColumn, ExternalTenantColumn}

	tenantRuntimeContextSelectedColumns = []string{tenantIDColumn}
	labelsSelectedColumns               = []string{"app_template_id"}
	applicationsSelectedColumns         = []string{"id"}
	tenantApplicationsSelectedColumns   = []string{tenantIDColumn}
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
		upserter:                            repo.NewUpserterGlobal(resource.Tenant, TableName, insertColumns, conflictingColumns, updateColumns),
		unsafeCreator:                       repo.NewUnsafeCreator(resource.Tenant, TableName, insertColumns, conflictingColumns),
		existQuerierGlobal:                  repo.NewExistQuerierGlobal(resource.Tenant, TableName),
		existQuerierGlobalWithConditionTree: repo.NewExistsQuerierGlobalWithConditionTree(resource.Tenant, TableName),
		singleGetterGlobal:                  repo.NewSingleGetterGlobal(resource.Tenant, TableName, insertColumns),
		pageableQuerierGlobal:               repo.NewPageableQuerierGlobal(resource.Tenant, TableName, insertColumns),
		listerGlobal:                        repo.NewListerGlobal(resource.Tenant, TableName, insertColumns),
		conditionTreeLister:                 repo.NewConditionTreeListerGlobal(TableName, insertColumns),
		updaterGlobal:                       repo.NewUpdaterGlobal(resource.Tenant, TableName, []string{ExternalNameColumn, ExternalTenantColumn, TypeColumn, ProviderNameColumn, StatusColumn}, []string{IDColumn}),
		deleterGlobal:                       repo.NewDeleterGlobal(resource.Tenant, TableName),
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
	return btm.ID, r.processParents(ctx, btm.ID, item.Parents, btm.Parents)
}

// Get retrieves the active tenant with matching internal ID from the Compass storage.
func (r *pgRepository) Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	var entity tenant.Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(IDColumn, id),
		repo.NewNotEqualCondition(StatusColumn, string(tenant.Inactive))}
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
		repo.NewEqualCondition(ExternalTenantColumn, externalTenantID),
		repo.NewNotEqualCondition(StatusColumn, string(tenant.Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, errors.Wrapf(err, "while getting tenant with external id %s", externalTenantID)
	}

	btm := r.conv.FromEntity(&entity)
	return r.enrichWithParents(ctx, btm)
}

// Exists checks if tenant with the provided internal ID exists in the Compass storage.
func (r *pgRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(IDColumn, id)})
}

// ExistsByExternalTenant checks if tenant with the provided external ID exists in the Compass storage.
func (r *pgRepository) ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(ExternalTenantColumn, externalTenant)})
}

// ExistsSubscribed checks if tenant is subscribed
func (r *pgRepository) ExistsSubscribed(ctx context.Context, id, selfDistinguishLabel string) (bool, error) {
	subaccountConditions := repo.Conditions{repo.NewEqualCondition(TypeColumn, tenant.Subaccount)}

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

	tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs, err := r.tenantApplicationsQueryBuilder.BuildQueryGlobal(false, repo.Conditions{repo.NewInConditionForSubQuery(IDColumn, applicationSubquery, applicationArgs)}...)
	if err != nil {
		return false, errors.Wrap(err, "while building query that fetches tenant id from tenant_applications table")
	}

	subscriptionConditions := repo.Conditions{
		repo.NewInConditionForSubQuery(IDColumn, tenantFromTenantRuntimeContextsSubquery, tenantFromTenantRuntimeContextsArgs),
		repo.NewInConditionForSubQuery(IDColumn, tenantFromTenantApplicationsSubquery, tenantFromTenantApplicationsArgs),
	}

	conditions := repo.And(
		append(
			append(
				repo.ConditionTreesFromConditions(subaccountConditions),
				repo.Or(repo.ConditionTreesFromConditions(subscriptionConditions)...),
			),
			&repo.ConditionTree{Operand: repo.NewEqualCondition(IDColumn, id)})...,
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
			ORDER BY %s DESC, t.%s ASC`, prefixedFields, tenantIDColumn, initializedComputedColumn, TableName, labelDefinitionsTableName, IDColumn, tenantIDColumn, StatusColumn, initializedComputedColumn, ExternalNameColumn)

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
		&repo.ConditionTree{Operand: repo.NewEqualCondition(StatusColumn, tenant.Active)},
		repo.Or(repo.ConditionTreesFromConditions(likeConditions)...))

	page, totalCount, err := r.pageableQuerierGlobal.ListGlobalWithAdditionalConditions(ctx, pageSize, cursor, ExternalNameColumn, &entityCollection, conditions)
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
		repo.NewInConditionForStringValues(ExternalTenantColumn, externalTenant)}

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

	return r.processParents(ctx, model.ID, parentsInternalIDs, tenantFromDB.Parents)
}

func (r *pgRepository) processParents(ctx context.Context, tenantID string, parents, parentsFromDB []string) error {
	parentsToAdd := slices.Filter(nil, parents, func(s string) bool {
		return !slices.Contains(parentsFromDB, s)
	})

	parentsToRemove := slices.Filter(nil, parentsFromDB, func(s string) bool {
		return !slices.Contains(parents, s)
	})

	for _, p := range parentsToRemove {
		if err := r.removeParent(ctx, p, tenantID); err != nil {
			return err
		}
	}

	for _, p := range parentsToAdd {
		if err := r.addParent(ctx, p, tenantID); err != nil {
			return err
		}
	}

	return nil
}

func (r *pgRepository) addParent(ctx context.Context, internalParentID, internalChildID string) error {
	if err := r.tenantParentRepo.Upsert(ctx, internalChildID, internalParentID); err != nil {
		return errors.Wrapf(err, "while adding tenant parent record for tenant with ID %s and parent tenant with ID %s", internalChildID, internalParentID)
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
		if err = r.deleteChildTenantsRecursively(ctx, tnt.ID); err != nil {
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

			resourceIDs, err := r.retrieveOwningResources(ctx, m2mTable, tnt.ID)
			if err != nil {
				return errors.Wrapf(err, "while retrieving owning resources for tenant with id %s", tnt.ID)
			}

			if len(resourceIDs) > 0 {
				deleter := repo.NewDeleterGlobal(topLevelEntity, topLevelEntityTable)
				if err := deleter.DeleteManyGlobal(ctx, repo.Conditions{repo.NewInConditionForStringValues("id", resourceIDs)}); err != nil {
					return errors.Wrapf(err, "while deleting resources owned by tenant %s", tnt.ID)
				}
			}
		}
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition(ExternalTenantColumn, externalTenant),
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
			repo.NewEqualCondition(IDColumn, childTenant),
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
		"tenantsTable":       TableName,
		"tenantParentsTable": "tenant_parents",
		"parentID":           "parent_id",
		"tenantID":           "tenant_id",
		"id":                 IDColumn,
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

// ListBySubscribedRuntimesAndApplicationTemplates lists subscribed runtimes and application templates for a given self register label with a custom SQL
func (r *pgRepository) ListBySubscribedRuntimesAndApplicationTemplates(ctx context.Context, selfRegDistinguishLabel string) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	rawStmt := `
		SELECT b.id, b.external_name, b.external_tenant, b.type, b.provider_name, b.status
		FROM {{ .tenantsTable }} b
		WHERE b.{{ .tenantTypeColumn }} = ?
		  AND (
				EXISTS (SELECT 1 FROM {{ .tenantRuntimeContextTable }} trc WHERE b.{{ .id }} = trc.{{ .tenantIDColumn }})
				OR EXISTS (
					SELECT 1
					FROM {{ .tenantApplicationsTable }} ta
							 JOIN {{ .applicationsTable }} app ON ta.{{ .id }} = app.{{ .id }}
							 JOIN {{ .labelsTable }} l ON app.{{ .appTemplateIDColumn}} = l.{{ .appTemplateIDColumn}}
					WHERE b.{{ .id }} = ta.{{ .tenantIDColumn }} AND l.{{ .labelKeyColumn }} = ? AND l.{{ .appTemplateIDColumn }} IS NOT NULL
				)
		)`

	t, err := template.New("").Parse(rawStmt)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"tenantsTable":              TableName,
		"tenantTypeColumn":          TypeColumn,
		"tenantRuntimeContextTable": tenantRuntimeContextTable,
		"tenantIDColumn":            tenantIDColumn,
		"tenantApplicationsTable":   tenantApplicationsTable,
		"applicationsTable":         applicationTable,
		"labelsTable":               labelsTable,
		"appTemplateIDColumn":       appTemplateIDColumn,
		"labelKeyColumn":            keyColumn,
		"id":                        IDColumn,
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return nil, errors.Wrap(err, "while executing template")
	}

	stmt := res.String()
	stmt = sqlx.Rebind(sqlx.DOLLAR, stmt)

	log.C(ctx).Debugf("Executing DB query: %s", stmt)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching persistence from context")
	}

	err = persist.SelectContext(ctx, &entityCollection, stmt, tenant.Subaccount, selfRegDistinguishLabel)
	if err != nil {
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

	getTenantsByParentAndTypeStmt := buildTenantsByParentAndTypeQuery()

	log.C(ctx).Debugf("Executing DB query: %s", getTenantsByParentAndTypeStmt)

	if err = persist.SelectContext(ctx, &entityCollection, getTenantsByParentAndTypeStmt, parentID, tenantType); err != nil {
		return nil, errors.Wrapf(err, "while listing tenants of type %s with parent ID %s", tenantType, parentID)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

// ListByIds list tenants by ids
func (r *pgRepository) ListByIds(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	var entityCollection tenant.EntityCollection

	conditions := repo.Conditions{
		repo.NewInConditionForStringValues(IDColumn, ids),
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
		repo.NewEqualCondition(TypeColumn, tenantType),
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
		repo.NewInConditionForStringValues(IDColumn, ids),
		repo.NewEqualCondition(TypeColumn, tenantType),
	}

	if err := r.listerGlobal.ListGlobal(ctx, &entityCollection, conditions...); err != nil {
		return nil, errors.Wrapf(err, "while listing tenants of type %s with ids %v", tenantType, ids)
	}

	return r.enrichManyWithParents(ctx, entityCollection)
}

func (r *pgRepository) retrieveOwningResources(ctx context.Context, m2mTable, tenantID string) ([]string, error) {
	rawStmt := `SELECT DISTINCT ta1.{{ .m2mID }}
				FROM {{ .m2mTable }} ta1
         			LEFT JOIN {{ .m2mTable }} ta2
                	   ON ta1.{{ .m2mID }} = ta2.{{ .m2mID }}
                    	   AND ta2.{{ .m2mTenantID }} = ta2.{{ .m2mSource }}
                    	   AND ta2.{{ .m2mTenantID }} <> ta1.{{ .m2mTenantID }}
				WHERE ta1.{{ .m2mTenantID }} = ? AND ta2.{{ .m2mID }} IS NULL`

	t, err := template.New("").Parse(rawStmt)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"m2mTable":    m2mTable,
		"m2mTenantID": repo.M2MTenantIDColumn,
		"m2mID":       repo.M2MResourceIDColumn,
		"m2mSource":   repo.M2MSourceColumn,
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

	var dest []string
	if err := persist.SelectContext(ctx, &dest, stmt, tenantID); err != nil {
		return nil, persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.List, "while listing owning resources from %s table for tenant %s", m2mTable, tenantID)
	}

	return dest, nil
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

func buildTenantsByParentAndTypeQuery() string {
	resultColumns := make([]string, 0, len(insertColumns))
	for _, column := range insertColumns {
		resultColumns = append(resultColumns, prefixWithTableName(TableName, column))
	}

	tenantTypeColumn := prefixWithTableName(TableName, TypeColumn)
	tenantsTableIDColumn := prefixWithTableName(TableName, IDColumn)
	parentsTableTenantIDColumn := prefixWithTableName(tenantparentmapping.TenantParentsTable, tenantparentmapping.TenantIDColumn)
	parentsTableParentIDColumn := prefixWithTableName(tenantparentmapping.TenantParentsTable, tenantparentmapping.ParentIDColumn)

	getTenantsByParentAndTypeStmt := fmt.Sprintf(getTenantsByParentAndType, strings.Join(resultColumns, ", "), TableName, tenantparentmapping.TenantParentsTable, tenantsTableIDColumn, parentsTableTenantIDColumn, parentsTableParentIDColumn, tenantTypeColumn)
	getTenantsByParentAndTypeStmt = sqlx.Rebind(sqlx.DOLLAR, getTenantsByParentAndTypeStmt)
	return getTenantsByParentAndTypeStmt
}

func prefixWithTableName(tableName, columnName string) string {
	return tableName + "." + columnName
}
