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

type SingleGetter interface {
	Get(ctx context.Context, tenant string, conditions Conditions, orderByParams OrderByParams, dest interface{}) error
}

type SingleGetterGlobal interface {
	GetGlobal(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error
}

type universalSingleGetter struct {
	tableName       string
	resourceType    resource.Type
	tenantColumn    *string
	selectedColumns string
}

func NewSingleGetter(resourceType resource.Type, tableName string, tenantColumn string, selectedColumns []string) SingleGetter {
	return &universalSingleGetter{
		resourceType:    resourceType,
		tableName:       tableName,
		tenantColumn:    &tenantColumn,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func NewSingleGetterGlobal(resourceType resource.Type, tableName string, selectedColumns []string) SingleGetterGlobal {
	return &universalSingleGetter{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func (g *universalSingleGetter) Get(ctx context.Context, tenant string, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}
	conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
	return g.unsafeGet(ctx, conditions, orderByParams, dest)
}

func (g *universalSingleGetter) GetGlobal(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	return g.unsafeGet(ctx, conditions, orderByParams, dest)
}

func (g *universalSingleGetter) unsafeGet(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	if dest == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(g.tableName, g.selectedColumns, conditions, orderByParams)
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	err = persist.Get(dest, query, args...)

	return persistence.MapSQLError(ctx, err, g.resourceType, resource.Get, "while getting object from '%s' table", g.tableName)
}
