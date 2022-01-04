package bundle

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a bundle entity
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
	DocumentationLabels           sql.NullString `db:"documentation_labels"`
	*repo.BaseEntity
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	return resource.Application, e.ApplicationID
}

// DecorateWithTenantID decorates the entity with the given tenant ID.
func (e *Entity) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*Entity
		TenantID string `db:"tenant_id"`
	}{
		Entity:   e,
		TenantID: tenant,
	}
}
