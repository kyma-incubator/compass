package api

import (
	"database/sql"
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
)

type Entity struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	AppID       string         `db:"app_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Group       sql.NullString `db:"group_name"`
	TargetURL   string         `db:"target_url"`
	DefaultAuth sql.NullString `db:"default_auth"`
	EntitySpec
	version.Version
}

type EntitySpec struct {
	SpecData   sql.NullString `db:"spec_data"`
	SpecFormat sql.NullString `db:"spec_format"`
	SpecType   sql.NullString `db:"spec_type"`
}

// sqlx cannot fetch value from embedded struct, thats why we need to help sqlx.
func (es *EntitySpec) Value() (driver.Value, error) {
	if es == nil {
		return nil, nil
	}
	return es, nil
}
