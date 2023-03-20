package data_product

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// "industry", "line_of_business", "product_type", "data_product_owner"

// Entity is a representation of a API in the database.
type Entity struct {
	ID               string         `db:"id"`
	ApplicationID    string         `db:"app_id"`
	OrdID            sql.NullString `db:"ord_id"`
	LocalID          sql.NullString `db:"local_id"`
	Title            sql.NullString `db:"title"`
	ShortDescription sql.NullString `db:"short_description"`
	Description      sql.NullString `db:"description"`
	Version          sql.NullString `db:"version"`
	ReleaseStatus    sql.NullString `db:"release_status"`
	Visibility       string         `db:"visibility"`
	OrdPackageID     sql.NullString `db:"package_id"`
	Tags             sql.NullString `db:"tags"`
	Industry         sql.NullString `db:"industry"`
	LineOfBusiness   sql.NullString `db:"line_of_business"`
	ProductType      sql.NullString `db:"product_type"`
	DataProductOwner sql.NullString `db:"data_product_owner"`
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
