package repo

import (
	"strings"
)

// FunctionBuilder is an interface for building queries about global entities.
type FunctionBuilder interface {
	BuildAdvisoryLock(isRebindingNeeded bool, identifier int64) (string, []interface{}, error)
}

type functionQueryBuilder struct{}

// NewQueryBuilderGlobal is a constructor for QueryBuilderGlobal about global entities.
func NewFunctionBuilder() FunctionBuilder {
	return &functionQueryBuilder{}
}

// BuildAdvisoryLock builds a SQL query for advisory lock on resource.
func (b *functionQueryBuilder) BuildAdvisoryLock(isRebindingNeeded bool, identifier int64) (string, []interface{}, error) {
	var stmtBuilder strings.Builder
	stmtBuilder.WriteString("SELECT pg_try_advisory_xact_lock($1)")
	var allArgs []interface{}
	allArgs = append(allArgs, identifier)
	return stmtBuilder.String(), allArgs, nil
}
