package spec

import (
	"database/sql"
)

type Entity struct {
	ID            string         `db:"id"`
	TenantID      string         `db:"tenant_id"`
	APIDefID      sql.NullString `db:"api_def_id"`
	EventAPIDefID sql.NullString `db:"event_def_id"`
	SpecData      sql.NullString `db:"spec_data"`

	APISpecFormat sql.NullString `db:"api_spec_format"`
	APISpecType   sql.NullString `db:"api_spec_type"`

	EventSpecFormat sql.NullString `db:"event_spec_format"`
	EventSpecType   sql.NullString `db:"event_spec_type"`

	CustomType sql.NullString `db:"custom_type"`
}
