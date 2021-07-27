package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/testctx/broker"

	log "github.com/sirupsen/logrus"
)

var testCtx *broker.SystemBrokerTestContext

func TestMain(m *testing.M) {
	log.Info("Starting System Broker Tests")

	cfg := config.SystemBrokerTestConfig{}
	config.ReadConfig(&cfg)

	dexToken := server.Token()

	ctx, err := broker.NewSystemBrokerTestContext(cfg, dexToken)
	if err != nil {
		log.Errorf("Failed to create test context: %s", err.Error())
		os.Exit(1)
	}

	testCtx = ctx
	exitCode := m.Run()
	log.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}
