package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/testctx/broker"

	"github.com/sirupsen/logrus"
)

var testCtx *broker.SystemBrokerTestContext

func TestMain(m *testing.M) {
	logrus.Info("Starting System Broker Tests")

	cfg := config.SystemBrokerTestConfig{}
	config.ReadConfig(&cfg)

	ctx, err := broker.NewSystemBrokerTestContext(cfg)
	if err != nil {
		logrus.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	testCtx = ctx
	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
