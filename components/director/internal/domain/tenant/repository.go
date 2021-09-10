package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.business_tenant_mappings`
const labelDefinitionsTableName string = `public.label_definitions`
const labelDefinitionsTenantIDColumn string = `tenant_id`

var tableColumns = []string{idColumn, externalNameColumn, externalTenantColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}
var (
	idColumn                  = "id"
	externalNameColumn        = "external_name"
	externalTenantColumn      = "external_tenant"
	parentColumn              = "parent"
	typeColumn                = "type"
	providerNameColumn        = "provider_name"
	statusColumn              = "status"
	initializedComputedColumn = "initialized"
)

// Converter converts tenants between the model.BusinessTenantMapping service-layer representation of a tenant and the repo-layer representation tenant.Entity.
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in *model.BusinessTenantMapping) *tenant.Entity
	FromEntity(in *tenant.Entity) *model.BusinessTenantMapping
}

type pgRepository struct {
	creator            repo.Creator
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
		creator:            repo.NewCreator(resource.Tenant, tableName, tableColumns),
		existQuerierGlobal: repo.NewExistQuerierGlobal(resource.Tenant, tableName),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Tenant, tableName, tableColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Tenant, tableName, tableColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Tenant, tableName, []string{externalNameColumn, externalTenantColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}, []string{idColumn}),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Tenant, tableName),
		conv:               conv,
	}
}

// Create adds the provided tenant into the Compass storage.
func (r *pgRepository) Create(ctx context.Context, item model.BusinessTenantMapping) error {
	return r.creator.Create(ctx, r.conv.ToEntity(&item))
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

	prefixedFields := strings.Join(str.PrefixStrings(tableColumns, "t."), ", ")
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
