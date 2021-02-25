package tests

import (
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

var testCtx *clients.SystemBrokerTestContext

func TestMain(m *testing.M) {
	logrus.Info("Starting System Broker Tests")

	cfg := config.SystemBrokerTestConfig{}
	err := config.ReadConfig(&cfg)
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	testCtx, err = clients.NewSystemBrokerTestContext(cfg)
	if err != nil {
		logrus.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}
	testctx.Init()
	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
