package persistence

import "github.com/lib/pq"

type persistenceCtxKey string

const (
	// PersistenceCtxKey is a key used in context to store the persistence object
	PersistenceCtxKey persistenceCtxKey = "PersistenceCtx"
	// NotNullViolation is an error code that happens when the required data is not provided
	NotNullViolation pq.ErrorCode = "23502"
	// UniqueViolation is an error code that happens when the Unique Key is violated
	UniqueViolation pq.ErrorCode = "23505"
	// ForeignKeyViolation is an error code that happens when try to delete a row from the referenced/parent table when referencing/child table reference that row in the parent table
	// or when create/update a row in the child table with reference to the parent that does not exists
	ForeignKeyViolation pq.ErrorCode = "23503"
	// CheckViolation is an error code that happens when the values in a column do not meet a specific requirement defined by CHECK conditions
	CheckViolation pq.ErrorCode = "23514"
	// ConstraintViolation is the class of errors that happens when any constraint is violated
	ConstraintViolation pq.ErrorClass = "23"
	// NoData missing godoc
	NoData pq.ErrorClass = "02"
)
