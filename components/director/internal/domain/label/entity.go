package label

import (
	"database/sql"
)

// Entity missing godoc
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

func (e *Entity) GetID() string {
	return e.ID
}

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

// Collection missing godoc
type Collection []Entity

// Len missing godoc
func (c Collection) Len() int {
	return len(c)
}
