package application

import (
	"database/sql"
	"time"
)

type Entity struct {
	ID                  string         `db:"id"`
	TenantID            string         `db:"tenant_id"`
	Name                string         `db:"name"`
	ProviderName        sql.NullString `db:"provider_name"`
	Description         sql.NullString `db:"description"`
	StatusCondition     string         `db:"status_condition"`
	StatusTimestamp     time.Time      `db:"status_timestamp"`
	HealthCheckURL      sql.NullString `db:"healthcheck_url"`
	IntegrationSystemID sql.NullString `db:"integration_system_id"`
}

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}
