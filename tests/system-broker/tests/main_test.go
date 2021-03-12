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
	config.ReadConfig(&cfg)

	ctx, err := clients.NewSystemBrokerTestContext(cfg)
	if err != nil {
		logrus.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	testCtx=ctx
	testctx.Init()
	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
