package bundleinstanceauth

import (
	"database/sql"
	"time"
)

// Entity missing godoc
type Entity struct {
	ID               string         `db:"id"`
	BundleID         string         `db:"bundle_id"`
	OwnerID          string         `db:"owner_id"`
	RuntimeID        sql.NullString `db:"runtime_id"`
	RuntimeContextID sql.NullString `db:"runtime_context_id"`
	Context          sql.NullString `db:"context"`
	InputParams      sql.NullString `db:"input_params"`
	AuthValue        sql.NullString `db:"auth_value"`
	StatusCondition  string         `db:"status_condition"`
	StatusTimestamp  time.Time      `db:"status_timestamp"`
	StatusMessage    string         `db:"status_message"`
	StatusReason     string         `db:"status_reason"`
}

// Collection missing godoc
type Collection []Entity

// Len missing godoc
func (c Collection) Len() int {
	return len(c)
}

// DecorateWithTenantID decorates the entity with the given tenant ID.
func (e *Entity) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*Entity
		TenantID string `db:"tenant_id"`
	}{
		Entity:   e,
		TenantID: tenant,
	}
}
