package mp_bundle

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID                            string         `db:"id"`
	OpenDiscoveryID               string         `db:"od_id"`
	TenantID                      string         `db:"tenant_id"`
	ApplicationID                 string         `db:"app_id"`
	Title                         string         `db:"title"`
	ShortDescription              string         `db:"short_description"`
	Description                   sql.NullString `db:"description"`
	InstanceAuthRequestJSONSchema sql.NullString `db:"instance_auth_request_json_schema"`
	DefaultInstanceAuth           sql.NullString `db:"default_instance_auth"`
	Tags                          sql.NullString `db:"tags"`
	LastUpdated                   time.Time      `db:"last_updated"`
	Extensions                    sql.NullString `db:"extensions"`
}
