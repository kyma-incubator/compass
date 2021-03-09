package tests

import (
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg"
	config "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"os"
	"testing"
	"time"

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
	pkg.TestTenants.Cleanup()
	fmt.Println("cleaned up")
	time.Sleep(20*time.Second)
	pkg.TestTenants.Init()
	defer pkg.TestTenants.Cleanup()

	testctx.Tc, err = testctx.NewTestContext()
	if err != nil {
		log.Fatal(err)
	}

	config.ReadConfig(conf)

	exitVal := m.Run()

	os.Exit(exitVal)
}
