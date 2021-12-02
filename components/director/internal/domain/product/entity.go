package product

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity represents a product entity.
type Entity struct {
	ID               string         `db:"id"`
	OrdID            string         `db:"ord_id"`
	ApplicationID    string         `db:"app_id"`
	Title            string         `db:"title"`
	ShortDescription string         `db:"short_description"`
	Vendor           string         `db:"vendor"`
	Parent           sql.NullString `db:"parent"`
	CorrelationIDs   sql.NullString `db:"correlation_ids"`
	Labels           sql.NullString `db:"labels"`
}

// GetID returns the product ID.
func (e *Entity) GetID() string {
	return e.ID
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
