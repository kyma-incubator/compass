package product

import (
	"database/sql"
)

type Entity struct {
	OrdID            string         `db:"ord_id"`
	TenantID         string         `db:"tenant_id"`
	ApplicationID    string         `db:"app_id"`
	Title            string         `db:"title"`
	ShortDescription string         `db:"short_description"`
	Vendor           string         `db:"vendor"`
	Parent           sql.NullString `db:"parent"`
	PPMSObjectID     sql.NullString `db:"sap_ppms_object_id"`
	Labels           sql.NullString `db:"labels"`
}
