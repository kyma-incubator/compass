package apitests

import (
	"crypto/rsa"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/sirupsen/logrus"
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

	exitCode := m.Run()

	logrus.Info("Tests finished. Exit code: ", exitCode)

	os.Exit(exitCode)
}
