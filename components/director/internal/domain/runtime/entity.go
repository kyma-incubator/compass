package runtime

import (
	"database/sql"
	"time"
)

// Runtime struct represents database entity for Runtime
type Runtime struct {
	ID                string         `db:"id"`
	Name              string         `db:"name"`
	Description       sql.NullString `db:"description"`
	StatusCondition   string         `db:"status_condition"`
	StatusTimestamp   time.Time      `db:"status_timestamp"`
	CreationTimestamp time.Time      `db:"creation_timestamp"`
}

func (e *Runtime) GetID() string {
	return e.ID
}
