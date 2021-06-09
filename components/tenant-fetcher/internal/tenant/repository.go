package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const tableName string = `public.business_tenant_mappings`
const (
	idColumn                  = "id"
	externalNameColumn        = "external_name"
	externalTenantColumn      = "external_tenant"
	customerId                = "customer_id"
	subdomain                 = "subdomain"
	providerNameColumn        = "provider_name"
	statusColumn              = "status"
	initializedComputedColumn = "initialized"
)

var tableColumns = []string{idColumn, externalNameColumn, externalTenantColumn, customerId, subdomain, providerNameColumn, statusColumn}

//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore
type TenantRepository interface {
	Create(ctx context.Context, item model.TenantModel) error
	DeleteByExternalID(ctx context.Context, tenantId string) error
}

//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore
type Converter interface {
	ToEntity(in model.TenantModel) tenant.Entity
	FromEntity(in *tenant.Entity) *model.TenantModel
}

type repository struct {
	converter Converter
	tableName string
	columns   []string
}

func NewRepository(conv Converter) *repository {
	return &repository{
		converter: conv,
		tableName: tableName,
		columns:   tableColumns,
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

func (r *repository) DeleteByExternalID(ctx context.Context, tenantId string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", r.tableName, externalTenantColumn)

	log.C(ctx).Infof("Executing DB query: %s", stmt)
	_, err = persist.ExecContext(ctx, stmt, tenantId)

	return persistence.MapSQLError(ctx, err, resource.Tenant, resource.Delete, "while deleting row to '%s' table", r.tableName)
}
