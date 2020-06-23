package labeldef

import "database/sql"

type Entity struct {
	ID         string         `db:"id"`
	TenantID   string         `db:"tenant_id"`
	Key        string         `db:"key"`
	SchemaJSON sql.NullString `db:"schema"`
}

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}
