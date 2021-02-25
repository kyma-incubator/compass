package ordvendor

import (
	"database/sql"
)

type Entity struct {
	OrdID         string         `db:"ord_id"`
	TenantID      string         `db:"tenant_id"`
	ApplicationID string         `db:"app_id"`
	Title         string         `db:"title"`
	Type          string         `db:"type"`
	Labels        sql.NullString `db:"labels"`
}
