package document

import (
	"database/sql"
)

type Entity struct {
	ID           string         `db:"id"`
	TenantID     string         `db:"tenant_id"`
	AppID        string         `db:"app_id"`
	Title        string         `db:"title"`
	DisplayName  string         `db:"display_name"`
	Description  string         `db:"description"`
	Format       string         `db:"format"`
	Kind         sql.NullString `db:"kind"`
	Data         sql.NullString `db:"kind"`
	FetchRequest sql.NullString `db:"fetch_request"`
}
