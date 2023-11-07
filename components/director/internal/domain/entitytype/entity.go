package entitytype

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a API in the database.
type Entity struct {
	ApplicationID                sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	OrdID                        string         `db:"ord_id"`
	LocalID                      string         `db:"local_id"`
	CorrelationIDs               sql.NullString `db:"correlation_ids"`
	Level                        string         `db:"level"`
	Title                        string         `db:"title"`
	ShortDescription             sql.NullString `db:"short_description"`
	Description                  sql.NullString `db:"description"`
	SystemInstanceAware          sql.NullBool   `db:"system_instance_aware"`
	ChangeLogEntries             sql.NullString `db:"changelog_entries"`
	PackageID                    string         `db:"package_id"`
	Visibility                   string         `db:"visibility"`
	Links                        sql.NullString `db:"links"`
	PartOfProducts               sql.NullString `db:"part_of_products"`
	LastUpdate                   sql.NullString `db:"last_update"`
	PolicyLevel                  sql.NullString `db:"policy_level"`
	CustomPolicyLevel            sql.NullString `db:"custom_policy_level"`
	ReleaseStatus                string         `db:"release_status"`
	SunsetDate                   sql.NullString `db:"sunset_date"`
	DeprecationDate              sql.NullString `db:"deprecation_date"`
	Successors                   sql.NullString `db:"successors"`
	Extensible                   sql.NullString `db:"extensible"`
	Tags                         sql.NullString `db:"tags"`
	Labels                       sql.NullString `db:"labels"`
	DocumentationLabels          sql.NullString `db:"documentation_labels"`
	ResourceHash                 sql.NullString `db:"resource_hash"`
	version.Version
	*repo.BaseEntity
}

// GetID returns the ID of the entity.
func (e *Entity) GetID() string {
	return e.ID
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
