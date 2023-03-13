package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

//go:generate mockery --name=HealthService --output=automock --outpkg=automock --case=underscore --disable-version-string

type HealthService interface {
	CheckHealth(ctx context.Context) (types.HealthStatus, error)
}

type HealthsHandler struct {
	Service HealthService
}

func (h HealthsHandler) Health(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	status, err := h.Service.CheckHealth(ctx)
	if err != nil {
		log.Err(err).Msg("Health check failed")
		ctx.JSON(http.StatusInternalServerError, status)
		return
	}

	ctx.JSON(http.StatusOK, status)
}
