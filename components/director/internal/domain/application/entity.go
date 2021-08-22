		package application

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type Entity struct {
	TenantID              string         `db:"tenant_id"`
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
	Labels                sql.NullString `db:"labels"`
	CorrelationIds        sql.NullString `db:"correlation_ids"`
	*repo.BaseEntity
}

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}
