package bundlereferences

import "database/sql"

// Entity missing godoc
type Entity struct {
	TenantID            string         `db:"tenant_id"`
	BundleID            sql.NullString `db:"bundle_id"`
	APIDefID            sql.NullString `db:"api_def_id"`
	EventDefID          sql.NullString `db:"event_def_id"`
	APIDefaultTargetURL sql.NullString `db:"api_def_url"`
}
