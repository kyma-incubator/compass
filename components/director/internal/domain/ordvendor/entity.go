package ordvendor

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity is the vendor entity
type Entity struct {
	ID                  string         `db:"id"`
	OrdID               string         `db:"ord_id"`
	ApplicationID       string         `db:"app_id"`
	Title               string         `db:"title"`
	Partners            sql.NullString `db:"partners"`
	Labels              sql.NullString `db:"labels"`
	DocumentationLabels sql.NullString `db:"documentation_labels"`
}

// GetID returns the ID
func (e *Entity) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	return resource.Application, e.ApplicationID
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
