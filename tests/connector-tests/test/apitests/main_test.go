package apitests

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/sirupsen/logrus"
)

const (
	apiAccessTimeout  = 60 * time.Second
	apiAccessInterval = 2 * time.Second
)

var (
	config          testkit.TestConfig
	internalClient  *connector.InternalClient
	hydratorClient  *connector.HydratorClient
	connectorClient *connector.TokenSecuredClient

	clientKey *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Connector Test")

	cfg, err := testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	config = cfg
	clientKey, err = testkit.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	internalClient = connector.NewInternalClient(config.InternalConnectorURL)
	hydratorClient = connector.NewHydratorClient(config.HydratorURL)
	connectorClient = connector.NewConnectorClient(config.ConnectorURL)

	// Wait for sidecar to initialize
	err = testkit.WaitForFunction(apiAccessInterval, apiAccessTimeout, func() bool {
		resp, err := http.Get(fmt.Sprintf("%s/%s", config.HydratorURL, "health"))
		if err != nil {
			logrus.Infof("Failed to access health endpoint, retrying in %f: %s", apiAccessInterval.Seconds(), err.Error())
			return false
		}

		if resp.StatusCode != http.StatusOK {
			logrus.Infof("Health endpoint responded with %s status, retrying in %f", resp.Status, apiAccessInterval.Seconds())
			return false
		}

		return true
	})
	if err != nil {
		logrus.Errorf("Error while waiting for access to API: %s", err.Error())
		os.Exit(1)
	}

	exitCode := m.Run()

	logrus.Info("Tests finished. Exit code: ", exitCode)

	os.Exit(exitCode)
}
