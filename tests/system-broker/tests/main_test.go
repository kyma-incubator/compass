package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/system-broker/pkg"

	"github.com/sirupsen/logrus"
)

var testCtx *pkg.TestContext

func TestMain(m *testing.M) {
	logrus.Info("Starting System Broker Tests")

	cfg, err := pkg.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	testCtx, err = pkg.NewTestContext(cfg)
	if err != nil {
		logrus.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
