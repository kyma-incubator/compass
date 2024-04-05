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

const (
	locationHeader = "Location"
)

//go:generate mockery --name=TenantMappingsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsService interface {
	CanSafelyRemoveTenantMapping(ctx context.Context, formationID string) (bool, error)
	ProcessTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
	RemoveTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
}

//go:generate mockery --name=AsyncProcessor --output=automock --outpkg=automock --case=underscore --disable-version-string
type AsyncProcessor interface {
	ProcessTMRequest(ctx context.Context, tenantMapping types.TenantMapping)
}

type TenantMappingsHandler struct {
	Service        TenantMappingsService
	AsyncProcessor AsyncProcessor
}

func (h TenantMappingsHandler) Patch(ctx *gin.Context) {
	var tenantMapping types.TenantMapping
	if err := json.NewDecoder(ctx.Request.Body).Decode(&tenantMapping); err != nil {
		err = errors.Newf("failed to decode tenant mapping body: %w", err)
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := tenantMapping.AssignedTenant.SetConfiguration(ctx); err != nil {
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
	ctx.Set(locationHeader, ctx.GetHeader(locationHeader))
	h.AsyncProcessor.ProcessTMRequest(ctx, tenantMapping)
}

func (h TenantMappingsHandler) handleValidateError(ctx *gin.Context, err error, tenantMapping *types.TenantMapping) {
	operation := tenantMapping.Operation
	if operation != types.OperationUnassign ||
		errors.Is(err, types.ErrInvalidFormationID) ||
		errors.Is(err, types.ErrInvalidAssignedTenantAppID) {
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
