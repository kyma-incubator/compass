package api

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type APIDefinition struct {
	ID                 string            `db:"id"`
	TenantID           string            `db:"tenant_id"`
	AppID              string            `db:"app_id"`
	Name               string            `db:"name"`
	Description        sql.NullString    `db:"description"`
	Group              sql.NullString    `db:"group_name"`
	TargetURL          string            `db:"target_url"`
	SpecData           sql.NullString    `db:"spec_data"`
	SpecFormat         model.SpecFormat  `db:"spec_format"`
	SpecType           model.APISpecType `db:"spec_type"`
	DefaultAuth        sql.NullString    `db:"default_auth"`
	SpecFetchRequestID sql.NullString    `db:"spec_fetch_request_id"`
	version.Version
}
