package document

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a document entity in the database.
type Entity struct {
	BndlID                       string         `db:"bundle_id"`
	AppID                        sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	Title                        string         `db:"title"`
	DisplayName                  string         `db:"display_name"`
	Description                  string         `db:"description"`
	Format                       string         `db:"format"`
	Kind                         sql.NullString `db:"kind"`
	Data                         sql.NullString `db:"data"`
	*repo.BaseEntity
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	return resource.Bundle, e.BndlID
}

// Collection is a collection of entities.
type Collection []Entity

// Len returns the length of the collection.
func (r Collection) Len() int {
	return len(r)
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
