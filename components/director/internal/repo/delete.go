package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

type Deleter interface {
	DeleteOne(ctx context.Context, tenant string, conditions Conditions) error
	DeleteMany(ctx context.Context, tenant string, conditions Conditions) error
}

type DeleterGlobal interface {
	DeleteOneGlobal(ctx context.Context, conditions Conditions) error
	DeleteManyGlobal(ctx context.Context, conditions Conditions) error
}

type universalDeleter struct {
	tableName    string
	resourceType resource.Type
	tenantColumn *string
}

func NewDeleter(resourceType resource.Type, tableName string, tenantColumn string) Deleter {
	return &universalDeleter{resourceType: resourceType, tableName: tableName, tenantColumn: &tenantColumn}
}

func NewDeleterGlobal(resourceType resource.Type, tableName string) DeleterGlobal {
	return &universalDeleter{tableName: tableName, resourceType: resourceType}
}

func (g *universalDeleter) DeleteOne(ctx context.Context, tenant string, conditions Conditions) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}
	conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
	return g.unsafeDelete(ctx, conditions, true)
}

func (g *universalDeleter) DeleteMany(ctx context.Context, tenant string, conditions Conditions) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}
	conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
	return g.unsafeDelete(ctx, conditions, false)
}

func (g *universalDeleter) DeleteOneGlobal(ctx context.Context, conditions Conditions) error {
	return g.unsafeDelete(ctx, conditions, true)
}

func (g *universalDeleter) DeleteManyGlobal(ctx context.Context, conditions Conditions) error {
	return g.unsafeDelete(ctx, conditions, false)
}

func (g *universalDeleter) unsafeDelete(ctx context.Context, conditions Conditions, requireSingleRemoval bool) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s", g.tableName))

	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err = writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	query := getQueryFromBuilder(stmtBuilder)
	log.C(ctx).Debugf("Executing DB query: %s", query)
	res, err := persist.Exec(query, allArgs...)
	if err = persistence.MapSQLError(ctx, err, g.resourceType, resource.Delete, "while deleting object from '%s' table", g.tableName); err != nil {
		return err
	}

	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if affected != 1 {
			return apperrors.NewInternalError("delete should remove single row, but removed %d rows", affected)
		}
	}

	return nil
}
