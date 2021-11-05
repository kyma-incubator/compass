package fetchrequest

import (
	"database/sql"
	"time"
)

// Entity missing godoc
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

func (e *Entity) GetID() string {
	return e.ID
}

func (e *Entity) GetParentID() string {
    if e.SpecID.Valid {
        return e.SpecID.String
    }
	return e.DocumentID.String
}