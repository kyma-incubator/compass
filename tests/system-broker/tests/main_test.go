package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/system-broker/pkg"

	"github.com/sirupsen/logrus"
)

var testCtx *pkg.TestContext
var config pkg.Config

func TestMain(m *testing.M) {
	logrus.Info("Starting System Broker Tests")

	var err error
	config, err = pkg.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	testCtx, err = pkg.NewTestContext(config)
	if err != nil {
		logrus.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
