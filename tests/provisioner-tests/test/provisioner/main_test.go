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
	var err error
	testSuite, err = NewTestSuite(config)
	if err != nil {
		logrus.Errorf("Failed to create new test suite: %s", err.Error())
		return 1
	}
	defer testSuite.Cleanup()

	if err = testSuite.Setup(); err != nil {
		logrus.Errorf("Failed to setup tests environment: %s", err.Error())
		return 1
	}
	return m.Run()
}
