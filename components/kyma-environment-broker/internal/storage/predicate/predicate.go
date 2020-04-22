package predicate

import (
	"sort"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"

	"github.com/gocraft/dbr"
)

// {{{ "Functional" Option Interfaces

// InMemoryPredicate allows to apply predicates for InMemory queries.
type InMemoryPredicate interface {
	ApplyToInMemory([]internal.InstanceWithOperation)
}

// PostgresPredicate allows to apply predicates for Postgres queries.
type PostgresPredicate interface {
	ApplyToPostgres(*dbr.SelectStmt)
}

// Predicate defines interface which allows to make a generic func signature
// between different driver implementations
type Predicate interface {
	PostgresPredicate
	InMemoryPredicate
}

// }}}

// {{{ Multi-Type Options

// SortAscByCreatedAt sorts the query output based on CreatedAt field on Instance.
func SortAscByCreatedAt() SortByCreatedAt {
	return SortByCreatedAt{}
}

type SortByCreatedAt struct{}

func (w SortByCreatedAt) ApplyToPostgres(stmt *dbr.SelectStmt) {
	stmt.OrderAsc("instances.created_at")
}

// TODO: It can be more generic but right now there's no reason to complicate it.
func (w SortByCreatedAt) ApplyToInMemory(in []internal.InstanceWithOperation) {
	sort.Slice(in, func(i, j int) bool {
		return in[i].CreatedAt.Before(in[j].CreatedAt)
	})
}

var _ InMemoryPredicate = &SortByCreatedAt{}
var _ PostgresPredicate = &SortByCreatedAt{}

// }}}
