package ordvendor

import (
	"database/sql"
)

type Entity struct {
	OrdID         string         `db:"ord_id"`
	TenantID      string         `db:"tenant_id"`
	ApplicationID string         `db:"app_id"`
	Title         string         `db:"title"`
	Partners      sql.NullString `db:"partners"`
	Labels        sql.NullString `db:"labels"`
}
