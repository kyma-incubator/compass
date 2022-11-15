package spec

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity represents a specification entity.
type Entity struct {
	ID            string         `db:"id"`
	APIDefID      sql.NullString `db:"api_def_id"`
	EventAPIDefID sql.NullString `db:"event_def_id"`
	SpecData      sql.NullString `db:"spec_data"`
	SpecDataHash  sql.NullString `db:"spec_data_hash"`

	APISpecFormat sql.NullString `db:"api_spec_format"`
	APISpecType   sql.NullString `db:"api_spec_type"`

	EventSpecFormat sql.NullString `db:"event_spec_format"`
	EventSpecType   sql.NullString `db:"event_spec_type"`

	CustomType sql.NullString `db:"custom_type"`
}

// GetID returns the ID of the entity.
func (e *Entity) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.APIDefID.Valid {
		return resource.API, e.APIDefID.String
	}
	return resource.EventDefinition, e.EventAPIDefID.String
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
