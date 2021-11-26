package label

import (
	"database/sql"
)

// Entity is a label entity.
type Entity struct {
	ID               string         `db:"id"`
	TenantID         sql.NullString `db:"tenant_id"`
	Key              string         `db:"key"`
	AppID            sql.NullString `db:"app_id"`
	RuntimeID        sql.NullString `db:"runtime_id"`
	RuntimeContextID sql.NullString `db:"runtime_context_id"`
	Value            string         `db:"value"`
	Version          int            `db:"version"`
}

// GetID returns the ID of the label.
func (e *Entity) GetID() string {
	return e.ID
}

// GetParentID returns the parent ID of the label.
func (e *Entity) GetParentID() string {
	if e.AppID.Valid {
		return e.AppID.String
	} else if e.RuntimeID.Valid {
		return e.RuntimeID.String
	} else if e.RuntimeContextID.Valid {
		return e.RuntimeContextID.String
	}
	return e.TenantID.String
}

// Collection is a collection of label entities.
type Collection []Entity

// Len returns the number of entities in the collection.
func (c Collection) Len() int {
	return len(c)
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
