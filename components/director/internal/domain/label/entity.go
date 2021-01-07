package label

import (
	"database/sql"
)

type Entity struct {
	ID               string         `db:"id"`
	TenantID         string         `db:"tenant_id"`
	Key              string         `db:"key"`
	AppID            sql.NullString `db:"app_id"`
	RuntimeID        sql.NullString `db:"runtime_id"`
	RuntimeContextID sql.NullString `db:"runtime_context_id"`
	Value            string         `db:"value"`
}
type Collection []Entity

func (c Collection) Len() int {
	return len(c)
}
