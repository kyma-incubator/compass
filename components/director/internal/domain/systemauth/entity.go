package systemauth

import "database/sql"

// Entity missing godoc
type Entity struct {
	ID                  string         `db:"id"`
	TenantID            sql.NullString `db:"tenant_id"`
	AppID               sql.NullString `db:"app_id"`
	RuntimeID           sql.NullString `db:"runtime_id"`
	IntegrationSystemID sql.NullString `db:"integration_system_id"`
	Value               sql.NullString `db:"value"`
}

// Collection missing godoc
type Collection []Entity

// Len missing godoc
func (c Collection) Len() int {
	return len(c)
}
