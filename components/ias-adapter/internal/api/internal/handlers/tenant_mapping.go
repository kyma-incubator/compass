package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

//go:generate mockery --name=TenantMappingsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsService interface {
	ProcessTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error
}

type TenantMappingsHandler struct {
	Service TenantMappingsService
}

func (h TenantMappingsHandler) Patch(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	var tenantMapping types.TenantMapping
	if err := json.NewDecoder(ctx.Request.Body).Decode(&tenantMapping); err != nil {
		err = errors.Newf("failed to decode tenant mapping body: %w", err)
		internal.RespondWithError(ctx, http.StatusUnprocessableEntity, err)
		return
	}
	if err := tenantMapping.AssignedTenants[0].SetConfiguration(); err != nil {
		err = errors.Newf("failed to set assigned tenant configuration: %w", err)
		log.Info().Msg(err.Error())
	}

	if err := tenantMapping.Validate(); err != nil {
		err = errors.Newf("tenant mapping body is invalid: %w", err)
		internal.RespondWithError(ctx, http.StatusUnprocessableEntity, err)
		return
	}

	logProcessing(ctx, tenantMapping)
	if err := h.Service.ProcessTenantMapping(ctx, tenantMapping); err != nil {
		err = errors.Newf("failed to process tenant mapping notification: %w", err)
		internal.RespondWithError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func logProcessing(ctx context.Context, tenantMapping types.TenantMapping) {
	log := logger.FromContext(ctx)
	tenantMapping.AssignedTenants[0].Parameters = types.AssignedTenantParameters{}
	tenantMapping.AssignedTenants[0].Config = types.AssignedTenantConfiguration{}
	tenantMapping.AssignedTenants[0].Configuration = types.AssignedTenantConfiguration{}
	log.Info().Msgf("Processing tenant mapping notification: %+v", tenantMapping)
}
