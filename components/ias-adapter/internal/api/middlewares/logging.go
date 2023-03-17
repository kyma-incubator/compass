package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

const CorrelationIDKey = "correlationID"

func Logging(ctx *gin.Context) {
	correlationID := uuid.NewString()
	ctxLogger := logger.Default().With().Str(CorrelationIDKey, correlationID).Logger()
	ctx.Set("logger", &ctxLogger)
	ctx.Set(CorrelationIDKey, correlationID)

	start := time.Now()
	status := ctx.Writer.Status()
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	ctx.Next()

	bodySize := ctx.Writer.Size()
	ctxLogger.Info().Msgf("%d %s %s %s %s %d", status, method, path, ctx.ClientIP(), time.Since(start), bodySize)
}
