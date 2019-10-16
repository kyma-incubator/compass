package provisioner

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/sirupsen/logrus"
)

var (
	config    testkit.TestConfig
	testSuite *TestSuite
)

func TestMain(m *testing.M) {
	var err error

	config, err = testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read configuration: %s", err.Error())
		os.Exit(1)
	}

	code := runTests(m)
	os.Exit(code)
}

func runTests(m *testing.M) int {
	testSuite, err := NewTestSuite(config)
	if err != nil {
		logrus.Errorf("Failed to setup tests environment: %s", err.Error())
		return 1
	}

	err = testSuite.Setup()
	if err != nil {
		logrus.Errorf("Failed to setup tests environment: %s", err.Error())
		return 1
	}
	defer testSuite.Cleanup()

	return m.Run()
}
