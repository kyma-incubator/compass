package runtime_auth

import "database/sql"

type Entity struct {
	// ID can be null to allow retrieving outer join result from DB
	ID        sql.NullString `db:"id"`
	TenantID  string         `db:"tenant_id"`
	RuntimeID string         `db:"runtime_id"`
	APIDefID  string         `db:"api_def_id"`
	Value     sql.NullString `db:"value"`
}
