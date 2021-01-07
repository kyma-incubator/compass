package apitests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/system-broker/tests/testkit"
	"github.com/sirupsen/logrus"
)

var testCtx *testkit.TestContext

func TestMain(m *testing.M) {
	logrus.Info("Starting System Broker Tests")

	cfg, err := testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	testCtx = testkit.NewTestContext(cfg)

	exitCode := m.Run()
	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
