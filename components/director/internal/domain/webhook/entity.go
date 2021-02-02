package webhook

import "database/sql"

type Entity struct {
	ID                  string         `db:"id"`
	TenantID            string         `db:"tenant_id"`
	ApplicationID       sql.NullString `db:"app_id"`
	RuntimeID           sql.NullString `db:"runtime_id"`
	IntegrationSystemID sql.NullString `db:"integration_system_id"`
	CollectionIDKey     sql.NullString `db:"correlation_id_key"`
	Type                string         `db:"type"`
	Mode                string         `db:"mode"`
	URL                 string         `db:"url"`
	Auth                sql.NullString `db:"auth"`
	RetryInterval       int            `db:"retry_interval"`
	Timeout             int            `db:"timeout"`
	URLTemplate         string         `db:"url_template"`
	InputTemplate       string         `db:"input_template"`
	HeaderTemplate      string         `db:"header_template"`
	OutputTemplate      string         `db:"output_template"`
	StatusTemplate      sql.NullString `db:"status_template"`
}

type Collection []Entity

func (c Collection) Len() int {
	return len(c)
}
