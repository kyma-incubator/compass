package ordvendor

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity is the vendor entity
type Entity struct {
	ID                           string         `db:"id"`
	OrdID                        string         `db:"ord_id"`
	ApplicationID                sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	Title                        string         `db:"title"`
	Partners                     sql.NullString `db:"partners"`
	Tags                         sql.NullString `db:"tags"`
	Labels                       sql.NullString `db:"labels"`
	DocumentationLabels          sql.NullString `db:"documentation_labels"`
}

// GetID returns the ID
func (e *Entity) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.ApplicationID.String != "" {
		return resource.Application, e.ApplicationID.String
	} else {
		return resource.ApplicationTemplateVersion, e.ApplicationTemplateVersionID.String
	}
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
