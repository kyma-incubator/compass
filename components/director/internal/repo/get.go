package repo

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// SingleGetter missing godoc
type SingleGetter interface {
	Get(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions, orderByParams OrderByParams, dest interface{}) error
}

// SingleGetterGlobal missing godoc
type SingleGetterGlobal interface {
	GetGlobal(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error
}

type universalSingleGetter struct {
	tableName       string
	resourceType    resource.Type
	tenantColumn    *string
	selectedColumns string
}

// NewSingleGetterWithEmbeddedTenant missing godoc
func NewSingleGetterWithEmbeddedTenant(tableName string, tenantColumn string, selectedColumns []string) SingleGetter {
	return &universalSingleGetter{
		tableName:       tableName,
		tenantColumn:    &tenantColumn,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func NewSingleGetter(tableName string, selectedColumns []string) SingleGetter {
	return &universalSingleGetter{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

// NewSingleGetterGlobal missing godoc
func NewSingleGetterGlobal(resourceType resource.Type, tableName string, selectedColumns []string) SingleGetterGlobal {
	return &universalSingleGetter{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

// Get missing godoc
func (g *universalSingleGetter) Get(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
		return g.unsafeGet(ctx, tenant, resourceType, conditions, orderByParams, dest)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, false)
	if err != nil {
		return err
	}

	conditions = append(conditions, tenantIsolation)

	return g.unsafeGet(ctx, tenant, resourceType, conditions, orderByParams, dest)
}

// GetGlobal missing godoc
func (g *universalSingleGetter) GetGlobal(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	return g.unsafeGet(ctx, "", g.resourceType, conditions, orderByParams, dest)
}

func (g *universalSingleGetter) unsafeGet(ctx context.Context, tenant string, resourceType resource.Type, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	if dest == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(resourceType, g.tableName, g.selectedColumns, tenant, conditions, orderByParams, true)
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	err = persist.GetContext(ctx, dest, query, args...)

	return persistence.MapSQLError(ctx, err, resourceType, resource.Get, "while getting object from '%s' table", g.tableName)
}
