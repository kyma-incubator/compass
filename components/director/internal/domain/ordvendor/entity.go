package ordvendor

import (
	"database/sql"
)

// Entity is the vendor entity
type Entity struct {
	ID            string         `db:"id"`
	OrdID         string         `db:"ord_id"`
	ApplicationID string         `db:"app_id"`
	Title         string         `db:"title"`
	Partners      sql.NullString `db:"partners"`
	Labels        sql.NullString `db:"labels"`
}

// GetID returns the ID
func (e *Entity) GetID() string {
	return e.ID
}

// GetParentID returns the parent ID
func (e *Entity) GetParentID() string {
	return e.ApplicationID
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
