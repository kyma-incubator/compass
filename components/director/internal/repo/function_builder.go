package repo

import (
	"strings"
)

// FunctionBuilder is an interface for invoking functions.
type FunctionBuilder interface {
	BuildAdvisoryLock(identifier int64) (string, []interface{}, error)
}

type functionQueryBuilder struct{}

// NewFunctionBuilder is a constructor for FunctionBuilder .
func NewFunctionBuilder() FunctionBuilder {
	return &functionQueryBuilder{}
}

// BuildAdvisoryLock builds a SQL query for advisory lock on resource.
func (b *functionQueryBuilder) BuildAdvisoryLock(identifier int64) (string, []interface{}, error) {
	var stmtBuilder strings.Builder
	stmtBuilder.WriteString("SELECT pg_try_advisory_xact_lock($1)")
	var allArgs []interface{}
	allArgs = append(allArgs, identifier)
	return stmtBuilder.String(), allArgs, nil
}
