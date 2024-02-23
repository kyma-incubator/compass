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
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

//go:generate mockery --name=TenantMappingsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsService interface {
	CanSafelyRemoveTenantMapping(ctx context.Context, formationID string) (bool, error)
	ProcessTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
	RemoveTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
}

type TenantMappingsHandler struct {
	Service TenantMappingsService
}

func (h TenantMappingsHandler) Patch(ctx *gin.Context) {
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

	reverseAssignmentState := tenantMapping.AssignedTenants[0].ReverseAssignmentState
	if tenantMapping.AssignedTenants[0].Operation == types.OperationAssign {
		if reverseAssignmentState != types.StateInitial && reverseAssignmentState != types.StateReady {
			errMsgf := "skipped processing tenant mapping notification with $.assignedTenants[0].reverseAssignmentState '%s'"
			err := errors.Newf(errMsgf, reverseAssignmentState)
			internal.RespondWithError(ctx, internal.IncompleteStatusCode, err)
			return
		}
	}
	if err := h.Service.ProcessTenantMapping(ctx, tenantMapping); err != nil {
		err = errors.Newf("failed to process tenant mapping notification: %w", err)
		operation := tenantMapping.AssignedTenants[0].Operation

		if operation == types.OperationAssign && errors.Is(err, errors.IASApplicationNotFound) {
			internal.RespondWithError(ctx, internal.NotFoundStatusCode, err)
			return
		}

		internal.RespondWithError(ctx, internal.ErrorStatusCode, err)
		return
	}

	ctx.Status(http.StatusOK)
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
