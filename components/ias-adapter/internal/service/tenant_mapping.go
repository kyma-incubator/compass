package service

import (
	"context"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ias"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/storage/postgres"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

type TenantMappingsStorage interface {
	UpsertTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
	ListTenantMappings(ctx context.Context, formationID string) (map[string]types.TenantMapping, error)
	DeleteTenantMapping(ctx context.Context, formationID, applicationID string) error
}

type IASService interface {
	GetApplication(ctx context.Context, iasHost, clientID, appTenantID string) (types.Application, error)
	UpdateApplicationConsumedAPIs(ctx context.Context, data ias.UpdateData) error
}

type TenantMappingsService struct {
	Storage    TenantMappingsStorage
	IASService IASService
}

func (s TenantMappingsService) ProcessTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error {
	formationID := tenantMapping.FormationID
	tenantMappingsFromDB, err := s.Storage.ListTenantMappings(ctx, formationID)
	if err != nil {
		logger.FromContext(ctx).Err(err).Msgf("Failed to get tenant mappings for formation '%s'", formationID)
		return errors.Newf("failed to get tenant mappings for formation '%s': %w", formationID, postgres.Error(err))
	}
	operation := tenantMapping.AssignedTenants[0].Operation

	switch operation {
	case types.OperationAssign:
		return s.handleAssign(ctx, tenantMapping, tenantMappingsFromDB)
	case types.OperationUnassign:
		return s.handleUnassign(ctx, tenantMapping, tenantMappingsFromDB)
	default:
		panic(errors.Newf("invalid tenant mapping operation %s", operation))
	}
}

func (s TenantMappingsService) handleAssign(ctx context.Context,
	tenantMapping types.TenantMapping, tenantMappingsFromDB map[string]types.TenantMapping) error {

	formationID := tenantMapping.FormationID
	uclAppID := tenantMapping.AssignedTenants[0].UCLApplicationID

	if _, ok := tenantMappingsFromDB[uclAppID]; ok {
		// Safeguard for empty consumedAPIs
		logger.FromContext(ctx).Warn().Msgf(
			"Received additional tenant mapping for app '%s' in formation '%s'. Skipping processing",
			uclAppID, formationID)
		return nil
	}

	if err := s.upsertTenantMappingInDB(ctx, tenantMapping); err != nil {
		return err
	}
	tenantMappingsFromDB[uclAppID] = tenantMapping
	if len(tenantMappingsFromDB) == 2 {
		if err := s.updateIASAppsConsumedAPIs(ctx, types.OperationAssign, tenantMappingsFromDB); err != nil {
			return errors.Newf("failed to update applications consumed APIs in formation '%s': %w", formationID, err)
		}
	}
	return nil
}

func (s TenantMappingsService) upsertTenantMappingInDB(ctx context.Context, tenantMapping types.TenantMapping) error {
	formationID := tenantMapping.FormationID
	if err := s.Storage.UpsertTenantMapping(ctx, tenantMapping); err != nil {
		logger.FromContext(ctx).Err(err).Msgf("Failed to upsert first tenant mapping for formation '%s'", formationID)
		return errors.Newf("failed to upsert first tenant mapping for formation '%s': %w",
			formationID, postgres.Error(err))
	}
	return nil
}

func (s TenantMappingsService) updateIASAppsConsumedAPIs(ctx context.Context,
	triggerOperation types.Operation, tenantMappingsMap map[string]types.TenantMapping) error {
	log := logger.FromContext(ctx)

	tenantMappingsCount := len(tenantMappingsMap)
	if tenantMappingsCount != 2 {
		panic(errors.Newf("tenantMappingsCount must be 2, got %d", tenantMappingsCount))
	}

	tenantMappings := toArray(tenantMappingsMap)
	log.Info().Msgf("Updating consumed APIs for applications in formation '%s' triggered by %s operation",
		tenantMappings[0].FormationID, triggerOperation)

	iasApps, err := s.getIASApps(ctx, tenantMappings)
	if err != nil {
		return errors.Newf("Failed to get IAS applications during %s operation: %w", triggerOperation, err)
	}

	for idx, tenantMapping := range tenantMappings {
		uclAppID := tenantMapping.AssignedTenants[0].UCLApplicationID
		log.Info().Msgf("Updating consumed APIs for application with UCL ID '%s' in formation '%s'",
			uclAppID, tenantMapping.FormationID)
		updateData := ias.UpdateData{
			Operation:             triggerOperation,
			TenantMapping:         tenantMapping,
			ConsumerApplication:   iasApps[idx],
			ProviderApplicationID: iasApps[abs(idx-1)].ID,
		}
		if err := s.IASService.UpdateApplicationConsumedAPIs(ctx, updateData); err != nil {
			return errors.Newf("error ocurred during IAS consumed APIs update", err)
		}
	}
	return nil
}

func (s TenantMappingsService) handleUnassign(ctx context.Context,
	tenantMapping types.TenantMapping, tenantMappingsFromDB map[string]types.TenantMapping) error {
	formationID := tenantMapping.FormationID
	if len(tenantMappingsFromDB) == 2 {
		if err := s.updateIASAppsConsumedAPIs(ctx, types.OperationUnassign, tenantMappingsFromDB); err != nil {
			return errors.Newf("failed to remove applications consumed APIs in formation '%s': %w", formationID, err)
		}
	}
	return s.removeTenantMappingFromDB(ctx, tenantMapping)
}

func (s TenantMappingsService) removeTenantMappingFromDB(
	ctx context.Context, tenantMapping types.TenantMapping) error {
	formationID := tenantMapping.FormationID
	err := s.Storage.DeleteTenantMapping(ctx, formationID, tenantMapping.AssignedTenants[0].UCLApplicationID)
	if err != nil {
		logger.FromContext(ctx).Err(err).Msgf("Failed to clean up tenant mapping for formation '%s'", formationID)
		return errors.Newf("failed to clean up tenant mapping for formation '%s': %w",
			formationID, postgres.Error(err))
	}
	return nil
}

func (s TenantMappingsService) getIASApplication(
	ctx context.Context, tenantMapping types.TenantMapping) (types.Application, error) {
	iasHost := tenantMapping.ReceiverTenant.ApplicationURL
	tenantMappingUCLApplicationID := tenantMapping.AssignedTenants[0].UCLApplicationID
	iasApplication, err := s.IASService.GetApplication(ctx, iasHost,
		tenantMapping.AssignedTenants[0].Parameters.ClientID, tenantMapping.AssignedTenants[0].LocalTenantID)
	if err != nil {
		return iasApplication, errors.Newf("failed to get IAS application with UCL ID '%s': %w",
			tenantMappingUCLApplicationID, err)
	}
	return iasApplication, nil
}

func (s TenantMappingsService) getIASApps(
	ctx context.Context, tenantMappings []types.TenantMapping) ([]types.Application, error) {
	iasApps := make([]types.Application, 0, len(tenantMappings))
	for _, tenantMapping := range tenantMappings {
		iasApp, err := s.getIASApplication(ctx, tenantMapping)
		if err != nil {
			return nil, err
		}
		iasApps = append(iasApps, iasApp)
	}
	return iasApps, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func toArray(tenantMappingsMap map[string]types.TenantMapping) []types.TenantMapping {
	tenantMappings := make([]types.TenantMapping, 0, len(tenantMappingsMap))
	for _, tenantMapping := range tenantMappingsMap {
		tenantMappings = append(tenantMappings, tenantMapping)
	}
	return tenantMappings
}
