package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/testctx/broker"
)

var (
	cfg     = &config.SystemBrokerTestConfig{}
	testCtx *broker.SystemBrokerTestContext
)

func TestMain(m *testing.M) {
	log.D().Info("Starting System Broker Tests")

	config.ReadConfig(&cfg)

	ctx, err := broker.NewSystemBrokerTestContext(*cfg)
	if err != nil {
		log.D().Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	testCtx = ctx
	exitCode := m.Run()
	log.D().Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
