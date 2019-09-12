package apitests

import (
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
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Connector Test")

	cfg, err := testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	config = cfg
	internalClient = connector.NewInternalClient(config.InternalConnectorURL)
	hydratorClient = connector.NewHydratorClient(config.HydratorURL)
	connectorClient = connector.NewConnectorClient(config.ConnectorURL)

	exitCode := m.Run()

	logrus.Info("Tests finished. Exit code: ", exitCode)

	os.Exit(exitCode)
}
