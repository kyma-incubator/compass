package integrationdependency

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity is a representation of an Integration Dependency in the database.
type Entity struct {
	ApplicationID                  sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID   sql.NullString `db:"app_template_version_id"`
	OrdID                          sql.NullString `db:"ord_id"`
	LocalTenantID                  sql.NullString `db:"local_tenant_id"`
	CorrelationIDs                 sql.NullString `db:"correlation_ids"`
	Title                          string         `db:"title"`
	ShortDescription               sql.NullString `db:"short_description"`
	Description                    sql.NullString `db:"description"`
	PackageID                      sql.NullString `db:"package_id"`
	LastUpdate                     sql.NullString `db:"last_update"`
	Visibility                     string         `db:"visibility"`
	ReleaseStatus                  sql.NullString `db:"release_status"`
	SunsetDate                     sql.NullString `db:"sunset_date"`
	Successors                     sql.NullString `db:"successors"`
	Mandatory                      sql.NullBool   `db:"mandatory"`
	RelatedIntegrationDependencies sql.NullString `db:"related_integration_dependencies"`
	Links                          sql.NullString `db:"links"`
	Tags                           sql.NullString `db:"tags"`
	Labels                         sql.NullString `db:"labels"`
	DocumentationLabels            sql.NullString `db:"documentation_labels"`
	ResourceHash                   sql.NullString `db:"resource_hash"`

	*repo.BaseEntity
	version.Version
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.ApplicationID.Valid {
		return resource.Application, e.ApplicationID.String
	} else if e.ApplicationTemplateVersionID.Valid {
		return resource.ApplicationTemplateVersion, e.ApplicationTemplateVersionID.String
	}

	return "", ""
}

// DecorateWithTenantID decorates the entity with the given tenant ID.
func (e *Entity) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*Entity
		TenantID string `db:"tenant_id"`
	}{
		Entity:   e,
		TenantID: tenant,
	}
}
