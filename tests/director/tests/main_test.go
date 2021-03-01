package tests

import (
	"github.com/kyma-incubator/compass/tests/pkg"
	config "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

var conf = &config.DirectorConfig{}

func TestMain(m *testing.M) {
	dbCfg := persistence.DatabaseConfig{}
	err := envconfig.Init(&dbCfg)
	if err != nil {
		log.Fatal(err)
	}

	pkg.TestTenants.Init()

	testctx.Tc, err = testctx.NewTestContext()
	if err != nil {
		log.Fatal(err)
	}

	config.ReadConfig(conf)

	exitVal := m.Run()

	pkg.TestTenants.Cleanup()

	os.Exit(exitVal)
}
