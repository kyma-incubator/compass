package document

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type Entity struct {
	TenantID    string         `db:"tenant_id"`
	BndlID      string         `db:"bundle_id"`
	Title       string         `db:"title"`
	DisplayName string         `db:"display_name"`
	Description string         `db:"description"`
	Format      string         `db:"format"`
	Kind        sql.NullString `db:"kind"`
	Data        sql.NullString `db:"data"`
	*repo.BaseEntity
}

type Collection []Entity

func (r Collection) Len() int {
	return len(r)
}
