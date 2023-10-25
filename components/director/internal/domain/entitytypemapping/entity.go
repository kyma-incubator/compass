package entitytypemapping

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a EntityTypeMapping in the database.
type Entity struct {
	APIDefinitionID   sql.NullString `db:"api_definition_id"`
	EventDefinitionID sql.NullString `db:"event_definition_id"`
	APIModelSelectors sql.NullString `db:"api_model_selectors"`
	EntityTypeTargets sql.NullString `db:"entity_type_targets"`
	*repo.BaseEntity
}

// GetID returns the ID of the entity.
func (e *Entity) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.APIDefinitionID.Valid {
		return resource.API, e.APIDefinitionID.String
	} else if e.EventDefinitionID.Valid {
		return resource.EventDefinition, e.EventDefinitionID.String
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
