package mp_package

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID                            string         `db:"id"`
	TenantID                      string         `db:"tenant_id"`
	ApplicationID                 string         `db:"app_id"`
	Name                          string         `db:"name"`
	Description                   sql.NullString `db:"description"`
	InstanceAuthRequestJSONSchema sql.NullString `db:"instance_auth_request_json_schema"`
	DefaultInstanceAuth           sql.NullString `db:"default_instance_auth"`
	EntityInstanceAuth
}

type EntityInstanceAuth struct {
	Context         sql.NullString `db:"context"`
	AuthValue       sql.NullString `db:"auth_value"`
	StatusCondition string         `db:"status_condition"`
	StatusTimestamp time.Time      `db:"status_timestamp"`
	StatusMessage   string         `db:"status_message"`
	StatusReason    string         `db:"status_reason"`
}
