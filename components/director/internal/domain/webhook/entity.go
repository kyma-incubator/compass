package webhook

import "database/sql"

// Entity missing godoc
type Entity struct {
	ID                    string         `db:"id"`
	TenantID              sql.NullString `db:"tenant_id"`
	ApplicationID         sql.NullString `db:"app_id"`
	ApplicationTemplateID sql.NullString `db:"app_template_id"`
	RuntimeID             sql.NullString `db:"runtime_id"`
	IntegrationSystemID   sql.NullString `db:"integration_system_id"`
	CollectionIDKey       sql.NullString `db:"correlation_id_key"`
	Type                  string         `db:"type"`
	Mode                  sql.NullString `db:"mode"`
	URL                   sql.NullString `db:"url"`
	Auth                  sql.NullString `db:"auth"`
	RetryInterval         sql.NullInt32  `db:"retry_interval"`
	Timeout               sql.NullInt32  `db:"timeout"`
	URLTemplate           sql.NullString `db:"url_template"`
	InputTemplate         sql.NullString `db:"input_template"`
	HeaderTemplate        sql.NullString `db:"header_template"`
	OutputTemplate        sql.NullString `db:"output_template"`
	StatusTemplate        sql.NullString `db:"status_template"`
}

// Collection missing godoc
type Collection []Entity

// Len missing godoc
func (c Collection) Len() int {
	return len(c)
}
