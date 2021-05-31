package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/testctx/broker"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	log "github.com/sirupsen/logrus"
)

var testCtx *broker.SystemBrokerTestContext

func TestMain(m *testing.M) {
	log.Info("Starting System Broker Tests")

	cfg := config.SystemBrokerTestConfig{}
	config.ReadConfig(&cfg)

	var dexToken string

	tokenConfig := server.Config{}
	err := envconfig.InitWithPrefix(&tokenConfig, "APP")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Get Dex id_token")
	if tokenConfig.IsWithToken {
		tokenConfig.Log = log.Infof
		ts := server.New(&tokenConfig)
		dexToken = server.WaitForToken(ts)
	} else {
		dexToken, err = idtokenprovider.GetDexToken()
		if err != nil {
			log.Fatal(errors.Wrap(err, "while getting dex token"))
		}
	}

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
