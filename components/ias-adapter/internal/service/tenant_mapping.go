package service

import (
	"context"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ias"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/storage/postgres"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

//go:generate mockery --name=TenantMappingsStorage --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsStorage interface {
	UpsertTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
	ListTenantMappings(ctx context.Context, formationID string) (map[string]types.TenantMapping, error)
	DeleteTenantMapping(ctx context.Context, formationID, applicationID string) error
}

//go:generate mockery --name=IASService --output=automock --outpkg=automock --case=underscore --disable-version-string
type IASService interface {
	GetApplicationByClientID(ctx context.Context, iasHost, clientID, appTenantID string) (types.Application, error)
	GetApplicationByName(ctx context.Context, iasHost, name string) (types.Application, error)
	CreateApplication(ctx context.Context, iasHost string, app *types.Application) (string, error)
	UpdateApplicationConsumedAPIs(ctx context.Context, data ias.UpdateData) error
}

type TenantMappingsService struct {
	Storage    TenantMappingsStorage
	IASService IASService
}

func (s TenantMappingsService) CanSafelyRemoveTenantMapping(ctx context.Context, formationID string) (bool, error) {
	tenantMappingsFromDB, err := s.Storage.ListTenantMappings(ctx, formationID)
	if err != nil {
		return false, errors.Newf("failed to get tenant mappings for formation '%s': %w", formationID, postgres.Error(err))
	}

	return len(tenantMappingsFromDB) < 2, nil
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

func (s TenantMappingsService) RemoveTenantMapping(
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

func (s TenantMappingsService) handleAssign(ctx context.Context,
	tenantMapping types.TenantMapping, tenantMappingsFromDB map[string]types.TenantMapping) error {

	formationID := tenantMapping.FormationID
	assignedTenant := tenantMapping.AssignedTenants[0]
	uclAppID := assignedTenant.UCLApplicationID

	_, tenantMappingAlreadyInDB := tenantMappingsFromDB[uclAppID]

	if assignedTenant.UCLApplicationType == types.S4ApplicationType && !tenantMappingAlreadyInDB {
		appID, err := s.createIfNotExistsIASApp(ctx, tenantMapping)
		if err != nil {
			logger.FromContext(ctx).Err(err).Msgf("Failed to create/find suitable IAS application")
			return errors.Newf("could not create/find suitable IAS application: %w", err)
		}
		tenantMapping.AssignedTenants[0].Parameters.IASApplicationID = appID
	}

	if tenantMappingAlreadyInDB && len(assignedTenant.Configuration.ConsumedAPIs) == 0 {
		// Safeguard for empty consumedAPIs
		logger.FromContext(ctx).Warn().Msgf(
			"Received additional tenant mapping for app '%s' in formation '%s'. Skipping upsert.",
			uclAppID, formationID)
	} else {
		if err := s.upsertTenantMappingInDB(ctx, tenantMapping); err != nil {
			return err
		}
		tenantMappingsFromDB[uclAppID] = tenantMapping
	}

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

	iasApps, err := s.getIASApps(ctx, triggerOperation, tenantMappings)
	if err != nil {
		return errors.Newf("Failed to get IAS applications during %s operation: %w", triggerOperation, err)
	}

	// could only be in the Unassign case
	if len(iasApps) < tenantMappingsCount {
		log.Warn().Msgf("Not all IAS applications are still present, skipping consumed APIs update")
		return nil
	}

	for idx, consumerApp := range iasApps {
		tenantMapping := tenantMappings[idx]
		uclAppID := tenantMapping.AssignedTenants[0].UCLApplicationID
		providerAppID := iasApps[abs(idx-1)].ID

		log.Info().Msgf(
			"Updating application '%s' consumed APIs with provider app id '%s' for UCL app '%s' in formation '%s'",
			consumerApp.ID, providerAppID, uclAppID, tenantMapping.FormationID)

		updateData := ias.UpdateData{
			Operation:             triggerOperation,
			TenantMapping:         tenantMapping,
			ConsumerApplication:   consumerApp,
			ProviderApplicationID: providerAppID,
		}

		if err := s.IASService.UpdateApplicationConsumedAPIs(ctx, updateData); err != nil {
			return errors.Newf("error occurred during IAS consumed APIs update", err)
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
	return s.RemoveTenantMapping(ctx, tenantMapping)
}

func (s TenantMappingsService) getIASApplication(
	ctx context.Context, tenantMapping types.TenantMapping) (types.Application, error) {
	iasHost := tenantMapping.ReceiverTenant.ApplicationURL
	tenantMappingUCLApplicationID := tenantMapping.AssignedTenants[0].UCLApplicationID
	clientID := tenantMapping.AssignedTenants[0].Parameters.ClientID
	localTenantID := tenantMapping.AssignedTenants[0].LocalTenantID
	iasAppID := tenantMapping.AssignedTenants[0].Parameters.IASApplicationID

	if iasAppID != "" {
		return types.Application{ID: iasAppID}, nil
	}

	iasApplication, err := s.IASService.GetApplicationByClientID(ctx, iasHost, clientID, localTenantID)
	if err != nil {
		return iasApplication, errors.Newf(
			"failed to get IAS application with clientID '%s' and tenantID '%s' for UCL App ID '%s': %w",
			clientID, localTenantID, tenantMappingUCLApplicationID, err)
	}
	return iasApplication, nil
}

func (s TenantMappingsService) getIASApps(ctx context.Context, triggerOperation types.Operation,
	tenantMappings []types.TenantMapping) ([]types.Application, error) {

	iasApps := make([]types.Application, 0, len(tenantMappings))
	for _, tenantMapping := range tenantMappings {
		iasApp, err := s.getIASApplication(ctx, tenantMapping)
		if err != nil {
			// allow missing IAS applications for unassign
			if errors.Is(err, errors.IASApplicationNotFound) && triggerOperation == types.OperationUnassign {
				logger.FromContext(ctx).Warn().Msgf("Application missing during unassign: %s", err.Error())
				continue
			}
			return nil, err
		}
		iasApps = append(iasApps, iasApp)
	}
	return iasApps, nil
}

func (s TenantMappingsService) createIfNotExistsIASApp(ctx context.Context, tenantMapping types.TenantMapping) (string, error) {
	iasHost := tenantMapping.ReceiverTenant.ApplicationURL
	s4Certificate := tenantMapping.AssignedTenants[0].Configuration.Credentials.InboundCommunicationCredentials.OAuth2mTLSAuthentication.Certificate
	if s4Certificate == "" {
		return "", errors.S4CertificateNotFound
	}
	s4AppName := string(types.S4ApplicationType) + "-" + tenantMapping.AssignedTenants[0].LocalTenantID

	existingS4App, err := s.IASService.GetApplicationByName(ctx, iasHost, s4AppName)
	if err == nil {
		logger.FromContext(ctx).Info().Msgf("Found existing IAS application with name: %s", s4AppName)
		return existingS4App.ID, nil
	}
	if !errors.Is(err, errors.IASApplicationNotFound) {
		return "", err
	}

	s4App := types.Application{
		Name: s4AppName,
		Authentication: types.ApplicationAuthentication{
			APICertificates: []types.ApiCertificateData{
				{Base64Certificate: s4Certificate},
			},
		},
	}

	return s.IASService.CreateApplication(ctx, iasHost, &s4App)
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
