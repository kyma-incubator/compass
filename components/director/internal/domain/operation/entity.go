package operation

import (
	"database/sql"
	"time"
)

// Entity is a representation of an Operation in the database.
type Entity struct {
	ID        string         `db:"id"`
	Type      string         `db:"op_type"`
	Status    string         `db:"status"`
	Data      sql.NullString `db:"data"`
	Error     sql.NullString `db:"error"`
	Priority  int            `db:"priority"`
	CreatedAt *time.Time     `db:"created_at"`
	UpdatedAt *time.Time     `db:"updated_at"`
}

// EntityCollection is a collection of operations.
type EntityCollection []Entity

// Len returns the number of entities in the collection.
func (s EntityCollection) Len() int {
	return len(s)
}
