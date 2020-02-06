package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/thanhpk/randstr"

	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/broker"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/gardener"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/kyma"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Broker   broker.Config
	Gardener gardener.Config

	WorkingNamespace string
}

type Suite struct {
	t *testing.T

	log            logrus.FieldLogger
	brokerClient   *broker.Client
	kymaClient     *kyma.Client
	gardenerClient *gardener.Client
}

func (ts *Suite) TearDown() {
	ts.log.Info("Cleaning up...")
	err := ts.gardenerClient.RuntimeTearDown()
	assert.NoError(ts.t, err)
	err = ts.brokerClient.DeprovisionRuntime()
	require.NoError(ts.t, err)
}

func newTestSuite(t *testing.T) *Suite {
	cfg := &Config{}
	err := envconfig.InitWithPrefix(cfg, "APP")
	require.NoError(t, err)

	runtimeID := uuid.New().String()
	clusterName := fmt.Sprintf("%s-%s", "e2e-provisioning", strings.ToLower(randstr.String(10)))

	log := logrus.New()
	kymaClient := kyma.NewClient(log)
	brokerClient := broker.NewClient(cfg.Broker, clusterName, runtimeID, log)
	gardenerClient, err := gardener.NewClient(cfg.Gardener, runtimeID, log)
	require.NoError(t, err)

	return &Suite{
		t:              t,
		log:            log,
		brokerClient:   brokerClient,
		kymaClient:     kymaClient,
		gardenerClient: gardenerClient,
	}
}
