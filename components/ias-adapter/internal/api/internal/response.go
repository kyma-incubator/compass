package internal

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

const (
	SuccessStatusCode    = http.StatusOK
	IncompleteStatusCode = http.StatusUnprocessableEntity
	NotFoundStatusCode   = http.StatusNotFound
	ErrorStatusCode      = http.StatusInternalServerError
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

func RespondWithConfigPending(ctx *gin.Context, config *types.TenantMappingConfiguration) {
	tenantMappingResponse := types.TenantMappingResponse{
		State: types.StateConfigPending,
	}
	if config != nil {
		tenantMappingResponse.Configuration = *config
	}
	ctx.AbortWithStatusJSON(http.StatusOK, tenantMappingResponse)
}
