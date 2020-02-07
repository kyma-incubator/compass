package testkit

import (
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type Configuration struct {
	ConnectivityAdapterUrl string
	DirectorUrl            string
	SkipSslVerify          bool
}

func ReadConfiguration() (Configuration, error) {
	var cfg Configuration

	err := envconfig.InitWithPrefix(&cfg, "")
	if err != nil {
		return Configuration{}, nil
	}

	logrus.Infof("Configuration: %v", cfg)

	return cfg, nil
}
