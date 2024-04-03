package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ucl"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

const (
	S4SAPManagedCommunicationScenario = "SAP_COM_1002"

	locationHeader = "Location"
)

//go:generate mockery --name=TenantMappingsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsService interface {
	CanSafelyRemoveTenantMapping(ctx context.Context, formationID string) (bool, error)
	ProcessTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
	RemoveTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
}

//go:generate mockery --name=UCLService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UCLService interface {
	ReportStatus(ctx context.Context, url string, statusReport ucl.StatusReport) error
}

type TenantMappingsHandler struct {
	Service TenantMappingsService
	UCLService
}

func (h TenantMappingsHandler) Patch(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	var tenantMapping types.TenantMapping
	if err := json.NewDecoder(ctx.Request.Body).Decode(&tenantMapping); err != nil {
		err = errors.Newf("failed to decode tenant mapping body: %w", err)
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := tenantMapping.AssignedTenants[0].SetConfiguration(ctx); err != nil {
		err = errors.Newf("failed to set assigned tenant configuration: %w", err)
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}
	logProcessing(ctx, tenantMapping)

	if err := tenantMapping.Validate(); err != nil {
		err = errors.Newf("tenant mapping body is invalid: %w", err)

		h.handleValidateError(ctx, err, &tenantMapping)
		return
	}
	if !strings.HasPrefix(tenantMapping.ReceiverTenant.ApplicationURL, "http") {
		tenantMapping.ReceiverTenant.ApplicationURL = "https://" + tenantMapping.ReceiverTenant.ApplicationURL
	}

	ctx.AbortWithStatus(http.StatusAccepted)

	reverseAssignmentState := tenantMapping.AssignedTenants[0].ReverseAssignmentState
	if tenantMapping.AssignedTenants[0].Operation == types.OperationAssign {
		if reverseAssignmentState != types.StateInitial && reverseAssignmentState != types.StateReady {
			log.Warn().Msgf("skipping processing tenant mapping notification with $.assignedTenants[0].reverseAssignmentState '%s'",
				reverseAssignmentState)
			h.reportStatus(ctx, ucl.StatusReport{State: types.StateConfigPending})
			return
		}
	}

	operation := tenantMapping.AssignedTenants[0].Operation

	if err := h.Service.ProcessTenantMapping(ctx, tenantMapping); err != nil {
		err = errors.Newf("failed to process tenant mapping notification: %w", err)

		if operation == types.OperationAssign {
			if errors.Is(err, errors.IASApplicationNotFound) {
				h.reportStatus(ctx, ucl.StatusReport{State: errorState(operation), Error: err.Error()})
				return
			}

			if errors.Is(err, errors.S4CertificateNotFound) {
				logger.FromContext(ctx).Info().Msgf("S/4 certificate not provided. Responding with CONFIG_PENDING.")
				s4Config := &types.TenantMappingConfiguration{
					Credentials: types.Credentials{
						OutboundCommunicationCredentials: types.CommunicationCredentials{
							OAuth2mTLSAuthentication: types.OAuth2mTLSAuthentication{
								CorrelationIds: []string{S4SAPManagedCommunicationScenario},
							},
						},
					},
				}
				h.reportStatus(ctx, ucl.StatusReport{State: types.StateConfigPending, Configuration: s4Config})
				return
			}
		}

		h.reportStatus(ctx, ucl.StatusReport{State: errorState(operation)})
		return
	}

	h.reportStatus(ctx, ucl.StatusReport{State: readyState(operation)})
}

func readyState(operation types.Operation) types.State {
	if operation == types.OperationAssign {
		return types.StateCreateReady
	}
	return types.StateDeleteReady
}

func errorState(operation types.Operation) types.State {
	if operation == types.OperationAssign {
		return types.StateCreateError
	}
	return types.StateDeleteError
}

func (h TenantMappingsHandler) reportStatus(ctx *gin.Context, statusReport ucl.StatusReport) {
	log := logger.FromContext(ctx)
	statusReportURL := ctx.GetHeader(locationHeader)

	if err := h.ReportStatus(ctx, statusReportURL, statusReport); err != nil {
		log.Error().Msgf("failed to report status to '%s': %s", statusReportURL, err)
	}
}

func (h TenantMappingsHandler) handleValidateError(ctx *gin.Context, err error, tenantMapping *types.TenantMapping) {
	operation := tenantMapping.AssignedTenants[0].Operation
	if operation != types.OperationUnassign ||
		errors.Is(err, types.ErrInvalidFormationID) ||
		errors.Is(err, types.ErrInvalidAssignedTenantID) {
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}

	okToRemoveTenantMapping, errCheck := h.Service.CanSafelyRemoveTenantMapping(ctx, tenantMapping.FormationID)

	if errCheck != nil {
		internal.RespondWithError(
			ctx, http.StatusInternalServerError, fmt.Errorf("%w, failed to check formation: %s", err, errCheck.Error()))
		return
	}

	if !okToRemoveTenantMapping {
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Service.RemoveTenantMapping(ctx, *tenantMapping); err != nil {
		internal.RespondWithError(
			ctx, http.StatusInternalServerError, fmt.Errorf("failed to remove tenant mapping: %w", err))
		return
	}

	logger.FromContext(ctx).Info().Msgf("%s. Responding OK as assignment is safe to remove", err.Error())
	ctx.Status(http.StatusOK)
}

func logProcessing(ctx context.Context, tenantMapping types.TenantMapping) {
	log := logger.FromContext(ctx)
	log.Info().Msgf("Processing tenant mapping notification (%s)", tenantMapping)
}
