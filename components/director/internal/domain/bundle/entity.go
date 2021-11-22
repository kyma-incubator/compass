package bundle

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity missing godoc
type Entity struct {
	ApplicationID                 string         `db:"app_id"`
	Name                          string         `db:"name"`
	Description                   sql.NullString `db:"description"`
	InstanceAuthRequestJSONSchema sql.NullString `db:"instance_auth_request_json_schema"`
	DefaultInstanceAuth           sql.NullString `db:"default_instance_auth"`
	OrdID                         sql.NullString `db:"ord_id"`
	ShortDescription              sql.NullString `db:"short_description"`
	Links                         sql.NullString `db:"links"`
	Labels                        sql.NullString `db:"labels"`
	CredentialExchangeStrategies  sql.NullString `db:"credential_exchange_strategies"`
	CorrelationIDs                sql.NullString `db:"correlation_ids"`
	*repo.BaseEntity
}

func (e *Entity) GetParentID() string {
	return e.ApplicationID
}
