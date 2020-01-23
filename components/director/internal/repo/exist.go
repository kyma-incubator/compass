package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type ExistQuerier interface {
	Exists(ctx context.Context, tenant string, conditions Conditions) (bool, error)
}

type ExistQuerierGlobal interface {
	ExistsGlobal(ctx context.Context, conditions Conditions) (bool, error)
}

type universalExistQuerier struct {
	tableName    string
	tenantColumn *string
}

func NewExistQuerier(tableName string, tenantColumn string) ExistQuerier {
	return &universalExistQuerier{tableName: tableName, tenantColumn: &tenantColumn}
}

func NewExistQuerierGlobal(tableName string) ExistQuerierGlobal {
	return &universalExistQuerier{tableName: tableName}
}

func (g *universalExistQuerier) Exists(ctx context.Context, tenant string, conditions Conditions) (bool, error) {
	if tenant == "" {
		return false, errors.New("tenant cannot be empty")
	}
	conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
	return g.unsafeExists(ctx, conditions)
}

func (g *universalExistQuerier) ExistsGlobal(ctx context.Context, conditions Conditions) (bool, error) {
	return g.unsafeExists(ctx, conditions)
}

func (g *universalExistQuerier) unsafeExists(ctx context.Context, conditions Conditions) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	var stmtBuilder strings.Builder
	startIdx := 1

	stmtBuilder.WriteString(fmt.Sprintf("SELECT 1 FROM %s", g.tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err = writeEnumeratedConditions(&stmtBuilder, startIdx, conditions)
	if err != nil {
		return false, errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	var count int
	err = persist.Get(&count, stmtBuilder.String(), allArgs...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting object from DB")
	}
	return true, nil
}
