package tenant

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"text/template"

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
)

var (
	idColumn                  = "id"
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
)

// Converter converts tenants between the model.BusinessTenantMapping service-layer representation of a tenant and the repo-layer representation tenant.Entity.
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in *model.BusinessTenantMapping) *tenant.Entity
	FromEntity(in *tenant.Entity) *model.BusinessTenantMapping
}

type pgRepository struct {
	upserter           repo.Upserter
	unsafeCreator      repo.UnsafeCreator
	existQuerierGlobal repo.ExistQuerierGlobal
	singleGetterGlobal repo.SingleGetterGlobal
	listerGlobal       repo.ListerGlobal
	updaterGlobal      repo.UpdaterGlobal
	deleterGlobal      repo.DeleterGlobal

	conv Converter
}

// NewRepository returns a new entity responsible for repo-layer tenant operations. All of its methods require persistence.PersistenceOp it the provided context.
func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		upserter:           repo.NewUpserter(resource.Tenant, tableName, insertColumns, conflictingColumns, updateColumns),
		unsafeCreator:      repo.NewUnsafeCreator(resource.Tenant, tableName, insertColumns, conflictingColumns),
		existQuerierGlobal: repo.NewExistQuerierGlobal(resource.Tenant, tableName),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Tenant, tableName, insertColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Tenant, tableName, insertColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Tenant, tableName, []string{externalNameColumn, externalTenantColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}, []string{idColumn}),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Tenant, tableName),
		conv:               conv,
	}
}

// UnsafeCreate adds a new tenant in the Compass DB in case it does not exist. If it already exists, no action is taken.
// It is not guaranteed that the provided tenant ID is the same as the tenant ID in the database.
func (r *pgRepository) UnsafeCreate(ctx context.Context, item model.BusinessTenantMapping) error {
	return r.unsafeCreator.UnsafeCreate(ctx, r.conv.ToEntity(&item))
}

// Upsert adds the provided tenant into the Compass storage if it does not exist, or updates it if it does.
func (r *pgRepository) Upsert(ctx context.Context, item model.BusinessTenantMapping) error {
	return r.upserter.Upsert(ctx, r.conv.ToEntity(&item))
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

	items := make([]*model.BusinessTenantMapping, 0, len(entityCollection))

	for _, entity := range entityCollection {
		tmModel := r.conv.FromEntity(&entity)
		items = append(items, tmModel)
	}
	return items, nil
}

// Update updates the values of tenant with matching internal, and external IDs.
func (r *pgRepository) Update(ctx context.Context, model *model.BusinessTenantMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	entity := r.conv.ToEntity(model)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// DeleteByExternalTenant removes a tenant with matching external ID from the Compass storage.
func (r *pgRepository) DeleteByExternalTenant(ctx context.Context, externalTenant string) error {
	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenant),
	}

	return r.deleterGlobal.DeleteManyGlobal(ctx, conditions)
}

// GetLowestOwnerForResource returns the lowest tenant in the hierarchy that is owner of a given resource.
func (r *pgRepository) GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error) {
	rawStmt := `(SELECT {{ .m2mTenantID }} FROM {{ .m2mTable }} ta WHERE ta.{{ .m2mID }} = ? AND ta.{{ .owner }} = true` +
				 ` AND (NOT EXISTS(SELECT 1 FROM {{ .tenantsTable }} WHERE {{ .parent }} = ta.{{ .m2mTenantID }})` +  // the tenant has no children
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
		"m2mTenantID": repo.M2MTenantIDColumn,
		"m2mTable": m2mTable,
		"m2mID": repo.M2MResourceIDColumn,
		"owner": repo.M2MOwnerColumn,
		"tenantsTable": tableName,
		"parent": parentColumn,
		"id": idColumn,
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
