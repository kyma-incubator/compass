package apptemplateversion

import (
	"database/sql"
	"time"
)

// Entity missing godoc
type Entity struct {
	ID                    string         `db:"id"`
	Version               string         `db:"version"`
	Title                 sql.NullString `db:"title"`
	ReleaseDate           *time.Time     `db:"release_date"`
	CorrelationIDs        sql.NullString `db:"correlation_ids"`
	CreatedAt             time.Time      `db:"created_at"`
	ApplicationTemplateID string         `db:"app_template_id"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (a EntityCollection) Len() int {
	return len(a)
}
