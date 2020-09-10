package mp_package

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID               string         `db:"id"`
	OpenDiscoveryID  string         `db:"od_id"`
	TenantID         string         `db:"tenant_id"`
	ApplicationID    string         `db:"app_id"`
	Title            string         `db:"title"`
	ShortDescription string         `db:"short_description"`
	Description      string         `db:"description"`
	Version          string         `db:"version"`
	Licence          sql.NullString `db:"licence"`
	LicenceType      sql.NullString `db:"licence_type"`
	TermsOfService   sql.NullString `db:"terms_of_service"`
	Logo             sql.NullString `db:"logo"`
	Image            sql.NullString `db:"image"`
	Provider         sql.NullString `db:"provider"`
	Actions          sql.NullString `db:"actions"`
	Tags             sql.NullString `db:"tags"`
	LastUpdated      time.Time      `db:"last_updated"`
	Extensions       sql.NullString `db:"extensions"`
}
