package eventdef

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type Entity struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	BndlID      string         `db:"bundle_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	GroupName   sql.NullString `db:"group_name"`
	*repo.BaseEntity
	version.Version
}
