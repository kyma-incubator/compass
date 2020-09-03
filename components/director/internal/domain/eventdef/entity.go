package eventdef

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
)

type Entity struct {
	ID               string         `db:"id"`
	TenantID         string         `db:"tenant_id"`
	BundleID         string         `db:"bundle_id"`
	Title            string         `db:"title"`
	ShortDescription string         `db:"short_description"`
	Description      sql.NullString `db:"description"`
	GroupName        sql.NullString `db:"group_name"`
	EventDefinitions string         `db:"event_definitions"`
	Tags             sql.NullString `db:"tags"`
	Documentation    sql.NullString `db:"documentation"`
	ChangelogEntries sql.NullString `db:"changelog_entries"`
	Logo             sql.NullString `db:"logo"`
	Image            sql.NullString `db:"image"`
	URL              sql.NullString `db:"url"`
	ReleaseStatus    string         `db:"release_status"`
	LastUpdated      time.Time      `db:"last_updated"`
	Extensions       sql.NullString `db:"extensions"`
	version.Version
	EntitySpec
}

type EntitySpec struct {
	SpecData   sql.NullString `db:"spec_data"`
	SpecFormat sql.NullString `db:"spec_format"`
	SpecType   sql.NullString `db:"spec_type"`
}
