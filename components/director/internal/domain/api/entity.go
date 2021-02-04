package api

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
)

type Entity struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	BndlID      string         `db:"bundle_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Group       sql.NullString `db:"group_name"`
	TargetURL   string         `db:"target_url"`
	Ready       bool           `db:"ready"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   time.Time      `db:"deleted_at"`
	Error       sql.NullString `db:"error"`
	EntitySpec
	version.Version
}

type EntitySpec struct {
	SpecData   sql.NullString `db:"spec_data"`
	SpecFormat sql.NullString `db:"spec_format"`
	SpecType   sql.NullString `db:"spec_type"`
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
