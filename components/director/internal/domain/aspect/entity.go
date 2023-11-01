package aspect

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
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
	APIResources                 sql.NullString `db:"api_resources"`
	EventResources               sql.NullString `db:"event_resources"`

	*repo.BaseEntity
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.IntegrationDependencyID != "" {
		return resource.IntegrationDependency, e.IntegrationDependencyID
	}

	return "", ""
}
