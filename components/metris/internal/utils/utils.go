package utils

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func TrackTime(name string, start time.Time, logger *zap.SugaredLogger) {
	elapsed := time.Since(start)
	logger.Debugf("%s took %s", name, elapsed)
}

// SetupSignalHandler registers for SIGTERM and SIGINT, a context is returned
// which is cancel when a signals is caught.
func SetupSignalHandler(shutdown func()) {
	sigbuf := 2
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	c := make(chan os.Signal, sigbuf)
	signal.Notify(c, signals...)

	go func() {
		<-c
		shutdown()
		<-c
		os.Exit(1)
	}()
}
