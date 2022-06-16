package application

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity missing godoc
type Entity struct {
	ApplicationTemplateID sql.NullString `db:"app_template_id"`
	Name                  string         `db:"name"`
	ProviderName          sql.NullString `db:"provider_name"`
	Description           sql.NullString `db:"description"`
	StatusCondition       string         `db:"status_condition"`
	StatusTimestamp       time.Time      `db:"status_timestamp"`
	HealthCheckURL        sql.NullString `db:"healthcheck_url"`
	IntegrationSystemID   sql.NullString `db:"integration_system_id"`
	BaseURL               sql.NullString `db:"base_url"`
	SystemNumber          sql.NullString `db:"system_number"`
	LocalTenantID         sql.NullString `db:"local_tenant_id"`
	Labels                sql.NullString `db:"labels"`
	CorrelationIDs        sql.NullString `db:"correlation_ids"`
	SystemStatus          sql.NullString `db:"system_status"`

	DocumentationLabels sql.NullString `db:"documentation_labels"`
	*repo.BaseEntity
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (a EntityCollection) Len() int {
	return len(a)
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
