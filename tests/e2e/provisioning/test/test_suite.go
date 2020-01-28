package test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/broker"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/kyma"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Broker broker.Config

	ServiceManager struct {
		ClassName string
	}
	Properties struct {
		ProvisionTimeout time.Duration
		ProvisionGCP     bool
	}
}

type Suite struct {
	log          logrus.FieldLogger
	brokerClient *broker.Client
	kymaClient   *kyma.Client
	config       *Config
}

func newTestSuite(t *testing.T) *Suite {
	cfg := &Config{}
	err := envconfig.InitWithPrefix(cfg, "APP")
	require.NoError(t, err)

	log := logrus.New()
	brokerClient := broker.NewClient(cfg.Broker, cfg.Properties.ProvisionGCP, cfg.Properties.ProvisionTimeout, log)
	kymaClient := kyma.NewClient(log)

	return &Suite{
		config:       cfg,
		log:          log,
		kymaClient:   kymaClient,
		brokerClient: brokerClient,
	}
}
