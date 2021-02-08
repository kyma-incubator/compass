package api

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
	Group       sql.NullString `db:"group_name"`
	TargetURL   string         `db:"target_url"`
	*repo.BaseEntity
	version.Version
}
