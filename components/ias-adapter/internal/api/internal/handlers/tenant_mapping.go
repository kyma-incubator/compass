package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

//go:generate mockery --name=TenantMappingsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingsService interface {
	UpdateApplicationsConsumedAPIs(ctx context.Context, tenantMapping types.TenantMapping) error
}

type TenantMappingsHandler struct {
	Service TenantMappingsService
}

func (h TenantMappingsHandler) Patch(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	var tenantMapping types.TenantMapping
	if err := json.NewDecoder(ctx.Request.Body).Decode(&tenantMapping); err != nil {
		errMsg := "Failed to decode tenant mapping body"
		log.Err(err).Msg(errMsg)
		ctx.JSON(http.StatusUnprocessableEntity, errorResponse{Error: errMsg})
		return
	}

	if err := tenantMapping.Validate(); err != nil {
		errMsg := "Tenant mapping body is invalid"
		log.Err(err).Msg(errMsg)
		ctx.JSON(http.StatusUnprocessableEntity, errorResponse{Error: fmt.Sprintf("%s:%s", errMsg, err.Error())})
		return
	}

	if err := h.Service.UpdateApplicationsConsumedAPIs(ctx, tenantMapping); err != nil {
		log.Err(err).Msg("Failed to update applications consumed APIs")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}
