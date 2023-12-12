package dataproduct

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a Data Product in the database.
type Entity struct {
	ApplicationID                sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	OrdID                        sql.NullString `db:"ord_id"`
	LocalTenantID                sql.NullString `db:"local_tenant_id"`
	CorrelationIDs               sql.NullString `db:"correlation_ids"`
	Title                        string         `db:"title"`
	ShortDescription             sql.NullString `db:"short_description"`
	Description                  sql.NullString `db:"description"`
	PackageID                    sql.NullString `db:"package_id"`
	LastUpdate                   sql.NullString `db:"last_update"`
	Visibility                   string         `db:"visibility"`
	ReleaseStatus                sql.NullString `db:"release_status"`
	Disabled                     sql.NullBool   `db:"disabled"`
	DeprecationDate              sql.NullString `db:"deprecation_date"`
	SunsetDate                   sql.NullString `db:"sunset_date"`
	Successors                   sql.NullString `db:"successors"`
	ChangeLogEntries             sql.NullString `db:"changelog_entries"`
	Type                         string         `db:"type"`
	Category                     string         `db:"category"`
	EntityTypes                  sql.NullString `db:"entity_types"`
	InputPorts                   sql.NullString `db:"input_ports"`
	OutputPorts                  sql.NullString `db:"output_ports"`
	Responsible                  sql.NullString `db:"responsible"`
	DataProductLinks             sql.NullString `db:"data_product_links"`
	Links                        sql.NullString `db:"links"`
	Industry                     sql.NullString `db:"industry"`
	LineOfBusiness               sql.NullString `db:"line_of_business"`
	Tags                         sql.NullString `db:"tags"`
	Labels                       sql.NullString `db:"labels"`
	DocumentationLabels          sql.NullString `db:"documentation_labels"`
	PolicyLevel                  sql.NullString `db:"policy_level"`
	CustomPolicyLevel            sql.NullString `db:"custom_policy_level"`
	SystemInstanceAware          sql.NullBool   `db:"system_instance_aware"`
	version.Version
	*repo.BaseEntity
}
