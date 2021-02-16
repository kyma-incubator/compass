package tests

import (
	"os"
	"testing"

	cfg "github.com/kyma-incubator/compass/tests/pkg/system-broker-config"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/sirupsen/logrus"
)

var testCtx *testctx.TestContext

func TestMain(m *testing.M) {
	logrus.Info("Starting System Broker Tests")

	cfg, err := cfg.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	testCtx, err = testctx.NewTestContext(cfg)
	if err != nil {
		logrus.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
