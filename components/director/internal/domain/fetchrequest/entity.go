package fetchrequest

import (
	"database/sql"
	"time"
)

// Entity represents a fetch request.
type Entity struct {
	ID              string         `db:"id"`
	URL             string         `db:"url"`
	SpecID          sql.NullString `db:"spec_id"`
	DocumentID      sql.NullString `db:"document_id"`
	Mode            string         `db:"mode"`
	Auth            sql.NullString `db:"auth"`
	Filter          sql.NullString `db:"filter"`
	StatusCondition string         `db:"status_condition"`
	StatusMessage   sql.NullString `db:"status_message"`
	StatusTimestamp time.Time      `db:"status_timestamp"`
}

// GetID returns the ID of the fetch request.
func (e *Entity) GetID() string {
	return e.ID
}

// GetParentID returns the ID of the parent.
func (e *Entity) GetParentID() string {
	if e.SpecID.Valid {
		return e.SpecID.String
	}
	return e.DocumentID.String
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
