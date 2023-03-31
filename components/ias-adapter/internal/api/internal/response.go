package internal

import (
	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
)

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestID"`
}

func RespondWithError(ctx *gin.Context, statusCode int, err error) {
	log := logger.FromContext(ctx)
	log.Err(err).Send()
	requestID, _ := ctx.Get(logCtx.RequestIDCtxKey)
	ctx.AbortWithStatusJSON(statusCode, errorResponse{Error: err.Error(), RequestID: requestID.(string)})
}
