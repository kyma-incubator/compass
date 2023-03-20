package port

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity is a representation of a Port in the database.
type Entity struct {
	ID                  string         `db:"id"`
	DataProductID       string         `db:"data_product_id"`
	ApplicationID       string         `db:"app_id"`
	Name                sql.NullString `db:"name"`
	PortType            sql.NullString `db:"port_type"`
	Description         sql.NullString `db:"description"`
	ProducerCardinality sql.NullString `db:"producer_cardinality"`
	Disabled            sql.NullBool   `db:"disabled"`
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

// GetID returns the product ID.
func (e *Entity) GetID() string {
	return e.ID
}
