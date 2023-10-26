package runtime

import (
	"database/sql"
	"time"
)

// Runtime struct represents database entity for Runtime
type Runtime struct {
	ID                   string         `db:"id"`
	Name                 string         `db:"name"`
	Description          sql.NullString `db:"description"`
	StatusCondition      string         `db:"status_condition"`
	StatusTimestamp      time.Time      `db:"status_timestamp"`
	CreationTimestamp    time.Time      `db:"creation_timestamp"`
	ApplicationNamespace sql.NullString `db:"application_namespace"`
}

// GetID returns ID of the runtime
func (e *Runtime) GetID() string {
	return e.ID
}

// DecorateWithTenantID decorates the entity with the given tenant ID.
func (e *Runtime) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*Runtime
		TenantID string `db:"tenant_id"`
	}{
		Runtime:  e,
		TenantID: tenant,
	}
}
