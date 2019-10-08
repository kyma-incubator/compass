package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	tenantColumn *string
}

func NewDeleter(tableName string, tenantColumn string) Deleter {
	return &universalDeleter{tableName: tableName, tenantColumn: &tenantColumn}
}

func NewDeleterGlobal(tableName string) DeleterGlobal {
	return &universalDeleter{tableName: tableName}
}

func (g *universalDeleter) DeleteOne(ctx context.Context, tenant string, conditions Conditions) error {
	return g.unsafeDelete(ctx, str.Ptr(tenant), conditions, true)
}

func (g *universalDeleter) DeleteMany(ctx context.Context, tenant string, conditions Conditions) error {
	return g.unsafeDelete(ctx, str.Ptr(tenant), conditions, false)
}

func (g *universalDeleter) DeleteOneGlobal(ctx context.Context, conditions Conditions) error {
	return g.unsafeDelete(ctx, nil, conditions, true)
}

func (g *universalDeleter) DeleteManyGlobal(ctx context.Context, conditions Conditions) error {
	return g.unsafeDelete(ctx, nil, conditions, false)
}

func (g *universalDeleter) unsafeDelete(ctx context.Context, tenant *string, conditions Conditions, requireSingleRemoval bool) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var stmtBuilder strings.Builder
	startIdx := 1

	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s", g.tableName))

	if tenant != nil || len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}
	if tenant != nil {
		stmtBuilder.WriteString(fmt.Sprintf(" %s = $1", *g.tenantColumn))
		if len(conditions) > 0 {
			stmtBuilder.WriteString(" AND")
		}
		startIdx = 2
	}
	err = writeEnumeratedConditions(&stmtBuilder, startIdx, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(tenant, conditions)
	res, err := persist.Exec(stmtBuilder.String(), allArgs...)
	if err != nil {
		return errors.Wrap(err, "while deleting from database")
	}
	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if affected != 1 {
			return fmt.Errorf("delete should remove single row, but removed %d rows", affected)
		}
	}

	return nil
}
