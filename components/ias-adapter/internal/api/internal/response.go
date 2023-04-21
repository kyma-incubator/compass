package internal

import (
	"net/url"

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
	errorMessage := url.QueryEscape(err.Error())
	ctx.AbortWithStatusJSON(statusCode, errorResponse{Error: errorMessage, RequestID: requestID.(string)})
}
