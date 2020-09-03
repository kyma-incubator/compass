package mp_bundle

import (
	"database/sql"
)

type Entity struct {
	ID                            string         `db:"id"`
	TenantID                      string         `db:"tenant_id"`
	ApplicationID                 string         `db:"app_id"`
	Name                          string         `db:"name"`
	Description                   sql.NullString `db:"description"`
	InstanceAuthRequestJSONSchema sql.NullString `db:"instance_auth_request_json_schema"`
	DefaultInstanceAuth           sql.NullString `db:"default_instance_auth"`
}
