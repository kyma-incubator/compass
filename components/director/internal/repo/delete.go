package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

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
	if tenant == "" {
		return errors.New("tenant cannot be empty")
	}
	conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
	return g.unsafeDelete(ctx, conditions, true)
}

func (g *universalDeleter) DeleteMany(ctx context.Context, tenant string, conditions Conditions) error {
	if tenant == "" {
		return errors.New("tenant cannot be empty")
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
	startIdx := 1

	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s", g.tableName))

	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err = writeEnumeratedConditions(&stmtBuilder, startIdx, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	res, err := persist.Exec(stmtBuilder.String(), allArgs...)
	if persistence.IsConstraintViolation(err) {
		return apperrors.NewConstraintViolationError(err.(*pq.Error).Table)
	}

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
