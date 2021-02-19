package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

func TestMain(m *testing.M) {
	dbCfg := persistence.DatabaseConfig{}
	err := envconfig.Init(&dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	transact, closeFunc, err := persistence.Configure(context.TODO(), dbCfg)

	defer func() {
		err := closeFunc()
		if err != nil {
			log.Fatal(err)
		}
	}()

	pkg.TestTenants.InitializeDB(transact)

	pkg.Tc, err = pkg.NewTestContext()
	if err != nil {
		log.Fatal(err)
	}

	exitVal := m.Run()

	pkg.TestTenants.CleanupDB(transact)

	os.Exit(exitVal)
}
