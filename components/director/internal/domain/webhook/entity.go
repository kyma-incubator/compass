package webhook

import "database/sql"

type Entity struct {
	ID       string         `db:"id"`
	TenantID string         `db:"tenant_id"`
	AppID    string         `db:"app_id"`
	Type     string         `db:"type"`
	URL      string         `db:"url"`
	Auth     sql.NullString `db:"auth"`
}

type Collection []Entity

func (c Collection) Len() int {
	return len(c)
}
