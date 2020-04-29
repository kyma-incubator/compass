package fetchrequest

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID              string         `db:"id"`
	TenantID        string         `db:"tenant_id"`
	URL             string         `db:"url"`
	APIDefID        sql.NullString `db:"api_def_id"`
	EventAPIDefID   sql.NullString `db:"event_api_def_id"`
	DocumentID      sql.NullString `db:"document_id"`
	Mode            string         `db:"mode"`
	Auth            sql.NullString `db:"auth"`
	Filter          sql.NullString `db:"filter"`
	StatusCondition string         `db:"status_condition"`
	StatusMessage   sql.NullString `db:"status_message"`
	StatusTimestamp time.Time      `db:"status_timestamp"`
}
