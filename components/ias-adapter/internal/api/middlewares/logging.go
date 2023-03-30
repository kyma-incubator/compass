package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/paths"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	logCtx "github.com/kyma-incubator/compass/components/ias-adapter/internal/logger/context"
)

func Logging(ctx *gin.Context) {
	requestID := getRequestID(ctx)
	ctxLogger := logger.Default().With().Str(logCtx.RequestIDCtxKey, requestID).Logger()
	ctx.Set(logCtx.LoggerCtxKey, &ctxLogger)
	ctx.Set(logCtx.RequestIDCtxKey, requestID)

	start := time.Now()
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	ctx.Next()

	status := ctx.Writer.Status()
	if status == http.StatusOK {
		if strings.HasPrefix(path, paths.HealthPath) || strings.HasPrefix(path, paths.ReadyPath) {
			return
		}
	}
	bodySize := ctx.Writer.Size()

	ctxLogger.Info().Msgf("%d %s %s %s %d", status, method, path, time.Since(start), bodySize)
}

func getRequestID(ctx *gin.Context) string {
	requestID := ctx.GetHeader(logCtx.RequestIDHeader)
	if requestID != "" {
		return requestID
	}
	return uuid.NewString()
}
