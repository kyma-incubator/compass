package persistence

import "github.com/lib/pq"

type persistenceCtxKey string

const (
	// PersistenceCtxKey is a key used in context to store the persistance object
	PersistenceCtxKey persistenceCtxKey = "PersistenceCtx"
	// NotNullViolation is an error code that happens when the required data is not provided
	NotNullViolation 	pq.ErrorCode  = "23502"
	// UniqueViolation is an error code that happens when the Unique Key is violated
	UniqueViolation 	pq.ErrorCode  = "23505"
	// ForeignKeyViolation is an error code that happens when the referenced/parent table try to delete/update row which still exist in the referencing/child table
	ForeignKeyViolation pq.ErrorCode  = "23503"
	// CheckViolation is an error code that happens when the values in a column do not meet a specific requirement defined by CHECK conditions
	CheckViolation 		pq.ErrorCode  = "23514"
	//ConstraintViolation is the class of errors that happens when any constraint is violated
	ConstraintViolation pq.ErrorClass = "23"
	NoData              pq.ErrorClass = "02"
)
