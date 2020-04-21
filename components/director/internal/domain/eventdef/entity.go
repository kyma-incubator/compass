package eventdef

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
)

type Entity struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	PkgID       string         `db:"package_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	GroupName   sql.NullString `db:"group_name"`
	version.Version
	EntitySpec
}

type EntitySpec struct {
	SpecData   sql.NullString `db:"spec_data"`
	SpecFormat sql.NullString `db:"spec_format"`
	SpecType   sql.NullString `db:"spec_type"`
}
