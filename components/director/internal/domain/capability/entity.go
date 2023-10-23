package capability

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
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
	CustomType                   sql.NullString `db:"custom_type"`
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
	LastUpdate                   sql.NullString `db:"last_update"`

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
