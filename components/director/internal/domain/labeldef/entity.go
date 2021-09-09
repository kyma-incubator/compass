package labeldef

import "database/sql"

// Entity missing godoc
type Entity struct {
	ID         string         `db:"id"`
	TenantID   string         `db:"tenant_id"`
	Key        string         `db:"key"`
	SchemaJSON sql.NullString `db:"schema"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (a EntityCollection) Len() int {
	return len(a)
}
