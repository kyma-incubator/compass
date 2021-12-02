package ordpackage

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity represents the ORD package entity.
type Entity struct {
	ID                string         `db:"id"`
	ApplicationID     string         `db:"app_id"`
	OrdID             string         `db:"ord_id"`
	Vendor            sql.NullString `db:"vendor"`
	Title             string         `db:"title"`
	ShortDescription  string         `db:"short_description"`
	Description       string         `db:"description"`
	Version           string         `db:"version"`
	PackageLinks      sql.NullString `db:"package_links"`
	Links             sql.NullString `db:"links"`
	LicenseType       sql.NullString `db:"licence_type"`
	Tags              sql.NullString `db:"tags"`
	Countries         sql.NullString `db:"countries"`
	Labels            sql.NullString `db:"labels"`
	PolicyLevel       string         `db:"policy_level"`
	CustomPolicyLevel sql.NullString `db:"custom_policy_level"`
	PartOfProducts    sql.NullString `db:"part_of_products"`
	LineOfBusiness    sql.NullString `db:"line_of_business"`
	Industry          sql.NullString `db:"industry"`
	ResourceHash      sql.NullString `db:"resource_hash"`
}

// GetID returns the ID of the entity.
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
