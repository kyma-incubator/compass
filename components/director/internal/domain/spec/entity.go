package spec

import "database/sql"

type Entity struct {
	ID                string         `db:"id"`
	TenantID          string         `db:"tenant_id"`
	APIDefinitionID   sql.NullString `db:"api_def_id"`
	EventDefinitionID sql.NullString `db:"event_def_id"`
	SpecData          sql.NullString `db:"spec_data"`
	SpecFormat        sql.NullString `db:"spec_format"`
	SpecType          sql.NullString `db:"spec_type"`
}
