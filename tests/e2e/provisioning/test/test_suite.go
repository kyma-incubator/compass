package test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/runtime"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/broker"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/kyma"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Broker  broker.Config
	Runtime runtime.Config

	TenantID string `default:"3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
}

// Suite provides set of clients able to provision and test Kyma runtime
type Suite struct {
	t *testing.T

	log           logrus.FieldLogger
	brokerClient  *broker.Client
	kymaClient    *kyma.Client
	runtimeClient *runtime.Client
}

func (ts *Suite) TearDown() {
	ts.log.Info("Cleaning up...")
	err := ts.runtimeClient.TearDown()
	assert.NoError(ts.t, err)
	err = ts.brokerClient.DeprovisionRuntime()
	require.NoError(ts.t, err)
}

func newTestSuite(t *testing.T) *Suite {
	cfg := &Config{}
	err := envconfig.InitWithPrefix(cfg, "APP")
	require.NoError(t, err)

	client := newHTTPClient(true)

	log := logrus.New()
	runtimeID := uuid.New().String()
	kymaClient := kyma.NewClient(*client, log)
	brokerClient := broker.NewClient(cfg.Broker, cfg.TenantID, runtimeID, *client, log)
	runtimeClient := runtime.NewClient(cfg.Runtime, cfg.TenantID, runtimeID, *client, log)

	return &Suite{
		t:             t,
		log:           log,
		kymaClient:    kymaClient,
		brokerClient:  brokerClient,
		runtimeClient: runtimeClient,
	}
}

func newHTTPClient(insecureSkipVerify bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			},
		},
	}
}
