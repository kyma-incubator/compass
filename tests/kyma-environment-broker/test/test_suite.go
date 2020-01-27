package test

import (
	"github.com/kyma-incubator/compass/tests/kyma-environment-broker/pkg/broker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	"testing"
	"time"
)

type Config struct {
	Broker struct {
		URL            string
		ServiceClassID string `default:"4deee563-e5ec-4731-b9b1-53b42d855f0c"`
	}
	ServiceManager struct {
		SecretName      string
		SecretNamespace string
		ClassName       string
	}
	Properties struct {
		ProvisionTimeout time.Duration
	}
}

type Suite struct {
	log          logrus.Logger
	brokerClient broker.Client
	config       Config
}

func newTestSuite(t *testing.T) *Suite {
	cfg := &Config{}
	err := envconfig.InitWithPrefix(cfg, "TEST")
	require.NoError(t, err)

	log := logrus.New()
	brokerClient := broker.NewClient(cfg.Broker.URL, *log)

	return &Suite{
		log:          *log,
		brokerClient: *brokerClient,
		config:       *cfg,
	}
}
