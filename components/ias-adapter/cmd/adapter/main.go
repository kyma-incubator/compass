package main

import (
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/app"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

func main() {
	log := logger.Default()

	cfg, err := config.New()
	if err != nil {
		log.Fatal().Msgf("Failed to init app config: %s", err)
	}
	log.Info().Msgf("Initialized config: %+v", cfg)

	app.Start(cfg)
}
