package repo

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// AdvisoryLockResult is structure used for result when advisory lock is acquired.
type AdvisoryLockResult struct {
	IsLocked bool `db:"pg_try_advisory_xact_lock"`
}

// FunctionerGlobal is an interface for invoking functions.
type FunctionerGlobal interface {
	AdvisoryLock(ctx context.Context, identifier int64) (bool, error)
}

type functionerGlobal struct{}

// NewFunctionerGlobal is a constructor for DBFunction .
func NewFunctionerGlobal() FunctionerGlobal {
	return &functionerGlobal{}
}

// AdvisoryLock executes SQL query for advisory lock on resource.
func (b *functionerGlobal) AdvisoryLock(ctx context.Context, identifier int64) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}
	var stmtBuilder strings.Builder
	stmtBuilder.WriteString("SELECT pg_try_advisory_xact_lock($1)")

	var allArgs []interface{}
	allArgs = append(allArgs, identifier)

	dest := []AdvisoryLockResult{}
	err = persist.SelectContext(ctx, &dest, stmtBuilder.String(), allArgs...)
	if len(dest) == 1 {
		return dest[0].IsLocked, persistence.MapSQLError(ctx, err, resource.Operation, resource.Get, "while executing advisory lock for identifier '%d'", identifier)
	}
	return false, apperrors.NewInternalError("while executing advisory lock")
}
