package testkit

import (
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type Configuration struct {
	DirectorUrl string `envconfig:"default=http://127.0.0.1:3000/graphql"`
}

func ReadConfiguration() (Configuration, error) {
	var cfg Configuration

	err := envconfig.InitWithPrefix(&cfg, "")
	if err != nil {
		return Configuration{}, err
	}

	logrus.Infof("Configuration: %v", cfg)

	return cfg, nil
}
