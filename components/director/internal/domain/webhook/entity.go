package webhook

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity is a webhook entity.
type Entity struct {
	ID                    string         `db:"id"`
	ApplicationID         sql.NullString `db:"app_id"`
	ApplicationTemplateID sql.NullString `db:"app_template_id"`
	RuntimeID             sql.NullString `db:"runtime_id"`
	IntegrationSystemID   sql.NullString `db:"integration_system_id"`
	FormationTemplateID   sql.NullString `db:"formation_template_id"`
	CollectionIDKey       sql.NullString `db:"correlation_id_key"`
	Type                  string         `db:"type"`
	Mode                  sql.NullString `db:"mode"`
	URL                   sql.NullString `db:"url"`
	Auth                  sql.NullString `db:"auth"`
	RetryInterval         sql.NullInt32  `db:"retry_interval"`
	Timeout               sql.NullInt32  `db:"timeout"`
	URLTemplate           sql.NullString `db:"url_template"`
	InputTemplate         sql.NullString `db:"input_template"`
	HeaderTemplate        sql.NullString `db:"header_template"`
	OutputTemplate        sql.NullString `db:"output_template"`
	StatusTemplate        sql.NullString `db:"status_template"`
	CreatedAt             *time.Time     `db:"created_at"`
}

// GetID returns the ID of the entity.
func (e *Entity) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.RuntimeID.Valid {
		return resource.Runtime, e.RuntimeID.String
	} else if e.ApplicationID.Valid {
		return resource.Application, e.ApplicationID.String
	} else if e.FormationTemplateID.Valid {
		return resource.FormationTemplate, e.FormationTemplateID.String
	}
	return "", ""
}

// Collection is a collection of webhook entities.
type Collection []Entity

// Len returns the number of entities in the collection.
func (c Collection) Len() int {
	return len(c)
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
