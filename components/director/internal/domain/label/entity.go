package label

import (
	"database/sql"
)

type Entity struct {
	ID        string         `db:"id"`
	TenantID  string         `db:"tenant_id"`
	Key       string         `db:"key"`
	AppID     sql.NullString `db:"app_id"`
	RuntimeID sql.NullString `db:"runtime_id"`
	Value     string         `db:"value"`
}
