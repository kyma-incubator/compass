package mp_bundle

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID                            string         `db:"id"`
	TenantID                      string         `db:"tenant_id"`
	ApplicationID                 string         `db:"app_id"`
	Name                          string         `db:"name"`
	Description                   sql.NullString `db:"description"`
	InstanceAuthRequestJSONSchema sql.NullString `db:"instance_auth_request_json_schema"`
	DefaultInstanceAuth           sql.NullString `db:"default_instance_auth"`
	Ready                         bool           `db:"ready"`
	CreatedAt                     time.Time      `db:"created_at"`
	UpdatedAt                     time.Time      `db:"updated_at"`
	DeletedAt                     time.Time      `db:"deleted_at"`
	Error                         sql.NullString `db:"error"`
}

func (e *Entity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *Entity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

func (e *Entity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

func (e *Entity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}
