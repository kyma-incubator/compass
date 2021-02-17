package mp_package

import (
	"database/sql"
)

type Entity struct {
	ID                string         `db:"id"`
	TenantID          string         `db:"tenant_id"`
	ApplicationID     string         `db:"app_id"`
	OrdID             string         `db:"ord_id"`
	Vendor            sql.NullString `db:"vendor"`
	Title             string         `db:"title"`
	ShortDescription  string         `db:"short_description"`
	Description       string         `db:"description"`
	Version           string         `db:"version"`
	PackageLinks      sql.NullString `db:"package_links"`
	Links             sql.NullString `db:"links"`
	LicenseType       sql.NullString `db:"licence_type"`
	Tags              sql.NullString `db:"tags"`
	Countries         sql.NullString `db:"countries"`
	Labels            sql.NullString `db:"labels"`
	PolicyLevel       string         `db:"policy_level"`
	CustomPolicyLevel sql.NullString `db:"custom_policy_level"`
	PartOfProducts    sql.NullString `db:"part_of_products"`
	LineOfBusiness    sql.NullString `db:"line_of_business"`
	Industry          sql.NullString `db:"industry"`
}
