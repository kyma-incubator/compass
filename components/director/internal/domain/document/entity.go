package document

import (
	"database/sql"
)

type Entity struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	AppID       sql.NullString `db:"app_id"`
	PkgID       sql.NullString `db:"package_id"`
	Title       string         `db:"title"`
	DisplayName string         `db:"display_name"`
	Description string         `db:"description"`
	Format      string         `db:"format"`
	Kind        sql.NullString `db:"kind"`
	Data        sql.NullString `db:"data"`
}

type Collection []Entity

func (r Collection) Len() int {
	return len(r)
}
