package application

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID                  string         `db:"id"`
	TenantID            string         `db:"tenant_id"`
	Name                string         `db:"name"`
	ProviderName        sql.NullString `db:"provider_name"`
	Description         sql.NullString `db:"description"`
	StatusCondition     string         `db:"status_condition"`
	StatusTimestamp     time.Time      `db:"status_timestamp"`
	HealthCheckURL      sql.NullString `db:"healthcheck_url"`
	IntegrationSystemID sql.NullString `db:"integration_system_id"`
	Ready               bool           `db:"ready"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
	DeletedAt           time.Time      `db:"deleted_at"`
	Error               sql.NullString `db:"error"`
}

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}

func (e *Entity) SetReady(ready bool) {
	e.Ready = ready
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

func (e *Entity) GetDeletedAt() time.Time {
	return e.DeletedAt
}

func (e *Entity) SetDeletedAt(t time.Time) {
	e.DeletedAt = t
}
