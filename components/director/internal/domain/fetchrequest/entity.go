package fetchrequest

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity represents a fetch request.
type Entity struct {
	ID              string         `db:"id"`
	URL             string         `db:"url"`
	SpecID          sql.NullString `db:"spec_id"`
	DocumentID      sql.NullString `db:"document_id"`
	Mode            string         `db:"mode"`
	Auth            sql.NullString `db:"auth"`
	Filter          sql.NullString `db:"filter"`
	StatusCondition string         `db:"status_condition"`
	StatusMessage   sql.NullString `db:"status_message"`
	StatusTimestamp time.Time      `db:"status_timestamp"`
}

// GetID returns the ID of the fetch request.
func (e *Entity) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(currentResourceType resource.Type) (resource.Type, string) {
	if e.SpecID.Valid {
		switch currentResourceType {
		case resource.APISpecFetchRequest:
			return resource.APISpecification, e.SpecID.String
		case resource.EventSpecFetchRequest:
			return resource.EventSpecification, e.SpecID.String
		case resource.CapabilitySpecFetchRequest:
			return resource.CapabilitySpecification, e.SpecID.String
		}
	}
	return resource.Document, e.DocumentID.String
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
