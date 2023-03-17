package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/middlewares"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

type errorResponse struct {
	Error         string `json:"error"`
	CorrelationID string `json:"correlationID"`
}

func respondWithError(ctx *gin.Context, statusCode int, err error) {
	log := logger.FromContext(ctx)
	log.Err(err).Send()
	correlationID, _ := ctx.Get(middlewares.CorrelationIDKey)
	ctx.JSON(statusCode, errorResponse{Error: err.Error(), CorrelationID: correlationID.(string)})
}
