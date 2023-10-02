package capability

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a Capability in the database.
type Entity struct {
	ApplicationID                sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	PackageID                    sql.NullString `db:"package_id"`
	Name                         string         `db:"name"`
	Description                  sql.NullString `db:"description"`
	OrdID                        sql.NullString `db:"ord_id"`
	Type                         string         `db:"type"`
	CustomType                   sql.NullString `db:"customType"`
	LocalTenantID                sql.NullString `db:"local_tenant_id"`
	ShortDescription             sql.NullString `db:"short_description"`
	SystemInstanceAware          sql.NullBool   `db:"system_instance_aware"`
	Tags                         sql.NullString `db:"tags"`
	Links                        sql.NullString `db:"links"`
	ReleaseStatus                sql.NullString `db:"release_status"`
	Labels                       sql.NullString `db:"labels"`
	Visibility                   string         `db:"visibility"`
	ResourceHash                 sql.NullString `db:"resource_hash"`
	DocumentationLabels          sql.NullString `db:"documentation_labels"`
	CorrelationIDs               sql.NullString `db:"correlation_ids"`

	*repo.BaseEntity
	version.Version
}
