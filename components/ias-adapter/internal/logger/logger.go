package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
)

const componentName = "UCL IAS Adapter"

var logger zerolog.Logger

func init() {
	var logWriter io.Writer
	if os.Getenv("ENV") == "dev" {
		logWriter = zerolog.NewConsoleWriter()
	} else {
		logWriter = os.Stdout
	}
	zerolog.TimestampFieldName = "timestamp"
	logger = zerolog.New(logWriter).With().Timestamp().Str("component", componentName).Logger()
}

func Default() *zerolog.Logger {
	return &logger
}

func FromContext(ctx context.Context) *zerolog.Logger {
	log := ctx.Value("logger")
	return log.(*zerolog.Logger)
}
