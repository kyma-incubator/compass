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

func (s TenantMappingsService) ProcessTenantMapping(ctx context.Context, tenantMapping1 types.TenantMapping) error {
	log := logger.FromContext(ctx)

	formationID := tenantMapping1.FormationID
	tenantMappings, err := s.Storage.ListTenantMappings(ctx, formationID)
	if err != nil {
		log.Err(err).Msgf("Failed to get tenant mappings for formation '%s'", formationID)
		return errors.Newf("failed to get tenant mappings for formation '%s': %w", formationID, postgres.Error(err))
	}

	applicationID1 := tenantMapping1.AssignedTenants[0].UCLApplicationID
	_, tenantMapping1Found := tenantMappings[applicationID1]
	tenantMapping2 := getTenantMapping2(tenantMappings, applicationID1)

	switch tenantMapping1.AssignedTenants[0].Operation {
	case types.OperationAssign:
		if len(tenantMappings) > 0 && !tenantMapping1Found {
			if err := s.updateApplicationsConsumedAPIs(ctx, tenantMapping1, tenantMapping2); err != nil {
				return errors.Newf("failed to update applications consumed APIs in formation '%s': %w", formationID, err)
			}
		}
		if err := s.Storage.UpsertTenantMapping(ctx, tenantMapping1); err != nil {
			log.Err(err).Msgf("Failed to upsert tenant mapping for formation '%s'", formationID)
			return errors.Newf("failed to upsert tenant mapping for formation '%s': %w", formationID, postgres.Error(err))
		}
	case types.OperationUnassign:
		if len(tenantMappings) > 1 && tenantMapping1Found {
			if err := s.updateApplicationsConsumedAPIs(ctx, tenantMapping1, tenantMapping2); err != nil {
				return errors.Newf("failed to update applications consumed APIs in formation '%s': %w", formationID, err)
			}
		}
		if err := s.Storage.DeleteTenantMapping(ctx, formationID, applicationID1); err != nil {
			log.Err(err).Msgf("Failed to clean up tenant mapping for formation '%s'", formationID)
			return errors.Newf("failed to clean up tenant mapping for formation '%s': %w", formationID, postgres.Error(err))
		}
	}
	return nil
}

func (s TenantMappingsService) updateApplicationsConsumedAPIs(ctx context.Context, tenantMapping1, tenantMapping2 types.TenantMapping) error {
	log := logger.FromContext(ctx)

	formationID := tenantMapping1.FormationID
	log.Info().Msgf("Updating consumed APIs for applications in formation '%s', triggered by '%s' operation",
		formationID, tenantMapping1.AssignedTenants[0].Operation)

	tenantMapping1ConsumedAPIs := tenantMapping1.AssignedTenants[0].Configuration.ConsumedAPIs
	tenantMapping2ConsumedAPIs := tenantMapping2.AssignedTenants[0].Configuration.ConsumedAPIs
	if len(tenantMapping1ConsumedAPIs) == 0 && len(tenantMapping2ConsumedAPIs) == 0 {
		log.Info().Msgf("No APIs to configure for applications in formation '%s'", formationID)
		return nil
	}

	iasHost := tenantMapping1.ReceiverTenant.ApplicationURL
	tenantMapping1UCLApplicationID := tenantMapping1.AssignedTenants[0].UCLApplicationID
	iasApplication1, err := s.IASService.GetApplication(ctx, iasHost,
		tenantMapping1.AssignedTenants[0].Parameters.ClientID, tenantMapping1.AssignedTenants[0].LocalTenantID)
	if err != nil {
		return errors.Newf("failed to get IAS application with UCL ID '%s': %w", tenantMapping1UCLApplicationID, err)
	}

	tenantMapping2UCLApplicationID := tenantMapping2.AssignedTenants[0].UCLApplicationID
	iasApplication2, err := s.IASService.GetApplication(ctx, iasHost,
		tenantMapping2.AssignedTenants[0].Parameters.ClientID, tenantMapping2.AssignedTenants[0].LocalTenantID)
	if err != nil {
		return errors.Newf("failed to get IAS application with UCL ID '%s': %w", tenantMapping2UCLApplicationID, err)
	}

	if len(tenantMapping1ConsumedAPIs) != 0 {
		log.Info().Msgf("Updating consumed APIs for application with UCL ID '%s'", tenantMapping1UCLApplicationID)
		updateData := ias.UpdateData{
			TenantMapping:         tenantMapping1,
			ConsumerApplication:   iasApplication1,
			ProviderApplicationID: iasApplication2.ID,
		}
		if err := s.IASService.UpdateApplicationConsumedAPIs(ctx, updateData); err != nil {
			return err
		}
	}
	if len(tenantMapping2ConsumedAPIs) != 0 {
		log.Info().Msgf("Updating consumed APIs for application with UCL ID '%s'", tenantMapping2UCLApplicationID)
		updateData := ias.UpdateData{
			TenantMapping:         tenantMapping2,
			ConsumerApplication:   iasApplication2,
			ProviderApplicationID: iasApplication1.ID,
		}
		if err := s.IASService.UpdateApplicationConsumedAPIs(ctx, updateData); err != nil {
			return err
		}
	}

	return nil
}

func getTenantMapping2(tenantMappings map[string]types.TenantMapping, uclApplicationID1 string) types.TenantMapping {
	for uclApplicationID, tenantMapping := range tenantMappings {
		if uclApplicationID != uclApplicationID1 {
			return tenantMapping
		}
	}
	return types.TenantMapping{}
}
