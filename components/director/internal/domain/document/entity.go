package document

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	BndlID      string         `db:"bundle_id"`
	Title       string         `db:"title"`
	DisplayName string         `db:"display_name"`
	Description string         `db:"description"`
	Format      string         `db:"format"`
	Kind        sql.NullString `db:"kind"`
	Data        sql.NullString `db:"data"`
	Ready       bool           `db:"ready"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   time.Time      `db:"deleted_at"`
	Error       sql.NullString `db:"error"`
}

type Collection []Entity

func (r Collection) Len() int {
	return len(r)
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
