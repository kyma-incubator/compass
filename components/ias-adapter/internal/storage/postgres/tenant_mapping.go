package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

const tenantMappingsTable = "tenant_mappings"

var (
	upsertTenantMappingQuery  = fmt.Sprintf("INSERT INTO %s (formation_id, ucl_application_id, value) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT pk DO UPDATE SET value=$3", tenantMappingsTable)
	selectTenantMappingsQuery = fmt.Sprintf("SELECT value FROM %s WHERE formation_id=$1", tenantMappingsTable)
	deleteTenantMappingQuery  = fmt.Sprintf("DELETE FROM %s WHERE formation_id=$1 AND ucl_application_id=$2", tenantMappingsTable)
)

func (c Connection) UpsertTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error {
	fields, err := tenantMappingFields(tenantMapping)
	if err != nil {
		return errors.Newf("failed to transform tenant mapping fields: %w", err)
	}
	if _, err := c.exec(ctx, upsertTenantMappingQuery, fields...); err != nil {
		return errors.Newf("failed to insert tenant mapping: %w", err)
	}
	return nil
}

func (c Connection) ListTenantMappings(ctx context.Context, formationID string) (map[string]types.TenantMapping, error) {
	tenantMappings := map[string]types.TenantMapping{}
	rows, err := c.query(ctx, selectTenantMappingsQuery, formationID)
	if err != nil {
		return tenantMappings, errors.Newf("failed to execute query: %w", err)
	}
	for rows.Next() {
		var (
			tenantMappingJSONString string
			tenantMapping           types.TenantMapping
		)
		if err := rows.Err(); err != nil {
			return tenantMappings, errors.Newf("failed to read db row: %w", err)
		}
		if err := rows.Scan(&tenantMappingJSONString); err != nil {
			return tenantMappings, errors.Newf("failed to scan row: %w", err)
		}
		if err := json.Unmarshal([]byte(tenantMappingJSONString), &tenantMapping); err != nil {
			return tenantMappings, errors.Newf("failed to unmarshal tenant mapping: %w", err)
		}
		if err := tenantMapping.AssignedTenant.SetConfiguration(ctx); err != nil {
			return tenantMappings, errors.Newf("failed to set tenant mapping assigned tenant configuration: %w", err)
		}
		tenantMappings[tenantMapping.AssignedTenant.AppID] = tenantMapping
	}

	return tenantMappings, nil
}

func (c Connection) DeleteTenantMapping(ctx context.Context, formationID, applicationID string) error {
	if _, err := c.exec(ctx, deleteTenantMappingQuery, formationID, applicationID); err != nil {
		return errors.Newf("failed to delete tenant mapping with formation_id '%s': %w", err)
	}
	return nil
}

func tenantMappingFields(tenantMapping types.TenantMapping) ([]any, error) {
	tenantMappingBytes, err := json.Marshal(tenantMapping)
	if err != nil {
		return []any{}, errors.Newf("failed to marshal tenant mapping to json: %w", err)
	}
	return []any{
		tenantMapping.FormationID,
		tenantMapping.AssignedTenant.AppID,
		string(tenantMappingBytes),
	}, nil
}
