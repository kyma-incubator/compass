package aspect

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of an Aspect in the database.
type Entity struct {
	ApplicationID                sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	IntegrationDependencyID      string         `db:"integration_dependency_id"`
	Title                        string         `db:"title"`
	Description                  sql.NullString `db:"description"`
	Mandatory                    sql.NullBool   `db:"mandatory"`
	SupportMultipleProviders     sql.NullBool   `db:"support_multiple_providers"`
	ApiResources                 sql.NullString `db:"api_resources"`
	EventResources               sql.NullString `db:"event_resources"`

	*repo.BaseEntity
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
