package testkit

import (
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type Configuration struct {
	ConnectivityAdapterUrl     string `envconfig:"default=https://adapter-gateway.kyma.local"`
	ConnectivityAdapterMtlsUrl string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
	DirectorUrl                string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipSslVerify              bool   `envconfig:"default=true"`
	EventsBaseURL              string `envconfig:"default=https://events.com"`
	Tenant                     string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	DirectorHealthzUrl         string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/healthz"`
}

func ReadConfiguration() (Configuration, error) {
	var cfg Configuration

	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return Configuration{}, err
	}

	logrus.Infof("Configuration: %v", cfg)

	return cfg, nil
}
