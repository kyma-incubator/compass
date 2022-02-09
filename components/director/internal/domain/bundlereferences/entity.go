package bundlereferences

import "database/sql"

// Entity represents a Compass BundleReference
type Entity struct {
	ID                  string         `db:"id"`
	BundleID            sql.NullString `db:"bundle_id"`
	APIDefID            sql.NullString `db:"api_def_id"`
	EventDefID          sql.NullString `db:"event_def_id"`
	APIDefaultTargetURL sql.NullString `db:"api_def_url"`
	Visibility          string         `db:"visibility"`
	IsDefaultBundle     sql.NullBool   `db:"is_default_bundle"`
}
