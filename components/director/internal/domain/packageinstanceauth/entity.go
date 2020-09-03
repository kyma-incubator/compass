package bundleinstanceauth

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID              string         `db:"id"`
	BundleID        string         `db:"bundle_id"`
	TenantID        string         `db:"tenant_id"`
	Context         sql.NullString `db:"context"`
	InputParams     sql.NullString `db:"input_params"`
	AuthValue       sql.NullString `db:"auth_value"`
	StatusCondition string         `db:"status_condition"`
	StatusTimestamp time.Time      `db:"status_timestamp"`
	StatusMessage   string         `db:"status_message"`
	StatusReason    string         `db:"status_reason"`
}

type Collection []Entity

func (c Collection) Len() int {
	return len(c)
}
