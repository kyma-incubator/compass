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

	testSuite, err = NewTestSuite(conf)

	code := m.Run()

	os.Exit(code)
}
