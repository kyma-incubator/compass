package product

import (
	"database/sql"
)

type Entity struct {
	ID               string         `db:"id"`
	OrdID            string         `db:"ord_id"`
	TenantID         string         `db:"tenant_id"`
	ApplicationID    string         `db:"app_id"`
	Title            string         `db:"title"`
	ShortDescription string         `db:"short_description"`
	Vendor           string         `db:"vendor"`
	Parent           sql.NullString `db:"parent"`
	CorrelationIds   sql.NullString `db:"correlation_ids"`
	Labels           sql.NullString `db:"labels"`
}
