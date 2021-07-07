package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
)

const tableName string = `public.business_tenant_mappings`
const (
	idColumn                  = "id"
	externalNameColumn        = "external_name"
	externalTenantColumn      = "external_tenant"
	parentColumn              = "parent"
	typeColumn                = "type"
	providerNameColumn        = "provider_name"
	statusColumn              = "status"
	initializedComputedColumn = "initialized"
)

var tableColumns = []string{idColumn, externalNameColumn, externalTenantColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}
var updatableTableColumns = []string{externalNameColumn, parentColumn, typeColumn, providerNameColumn, statusColumn}

//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore
type TenantRepository interface {
	Create(ctx context.Context, item model.TenantModel) error
	GetByExternalID(ctx context.Context, tenantId string) (model.TenantModel, error)
	Update(ctx context.Context, item model.TenantModel) error
}

//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in model.TenantModel) tenant.Entity
	FromEntity(in tenant.Entity) model.TenantModel
}

type repository struct {
	converter        Converter
	tableName        string
	columns          []string
	updatableColumns []string
}

func NewRepository(conv Converter) *repository {
	return &repository{
		converter:        conv,
		tableName:        tableName,
		columns:          tableColumns,
		updatableColumns: updatableTableColumns,
	}
}

func (r *repository) Create(ctx context.Context, item model.TenantModel) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	dbEntity := r.converter.ToEntity(item)

	var values []string
	for _, c := range r.columns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", r.tableName, strings.Join(r.columns, ", "), strings.Join(values, ", "))

	log.C(ctx).Infof("Executing DB query: %s", stmt)
	_, err = persist.NamedExecContext(ctx, stmt, dbEntity)

	return persistence.MapSQLError(ctx, err, resource.Tenant, resource.Create, "while inserting row to '%s' table", r.tableName)
}

func (r *repository) GetByExternalID(ctx context.Context, tenantId string) (model.TenantModel, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return model.TenantModel{}, err
	}

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE %s='%s'", strings.Join(r.columns, ", "), r.tableName, externalTenantColumn, tenantId)

	log.C(ctx).Infof("Executing DB query: %s", stmt)
	var tenantEntity tenant.Entity
	if err := persist.GetContext(ctx, &tenantEntity, stmt); err != nil {
		return model.TenantModel{}, err
	}

	return r.converter.FromEntity(tenantEntity), persistence.MapSQLError(ctx, err, resource.Tenant, resource.Get, "while fetching row to '%s' table", r.tableName)
}

func (r *repository) Update(ctx context.Context, item model.TenantModel) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	dbEntity := r.converter.ToEntity(item)

	var fieldsToSet []string
	for _, c := range r.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s='%s'", r.tableName, strings.Join(fieldsToSet, ", "), externalTenantColumn, dbEntity.ExternalTenant)

	log.C(ctx).Infof("Executing DB query: %s", stmt)
	_, err = persist.NamedExecContext(ctx, stmt, dbEntity)

	return persistence.MapSQLError(ctx, err, resource.Tenant, resource.Update, "while updating row to '%s' table", r.tableName)
}
