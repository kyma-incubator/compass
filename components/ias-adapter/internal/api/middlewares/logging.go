package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

func Logging(ctx *gin.Context) {
	ctxLogger := logger.Default().With().Str("correlationID", uuid.NewString()).Logger()
	ctx.Set("logger", &ctxLogger)

	start := time.Now()
	status := ctx.Writer.Status()
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	ctx.Next()

	bodySize := ctx.Writer.Size()
	ctxLogger.Info().Msgf("%d %s %s %s %s %d", status, method, path, ctx.ClientIP(), time.Since(start), bodySize)
}
