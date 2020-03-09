package api

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

func TestMain(m *testing.M) {
	db := persistence.DatabaseConfig{}
	err := envconfig.Init(&db)
	if err != nil {
		log.Fatal(err)
	}

	connString := persistence.GetConnString(db)
	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), connString)

	defer func() {
		err := closeFunc()
		if err != nil {
			log.Fatal(err)
		}
	}()

	testTenants.InitializeDB(transact)

	tc, err = newTestContext()
	if err != nil {
		log.Fatal(err)
	}

	exitVal := m.Run()

	testTenants.CleanupDB(transact)

	os.Exit(exitVal)
}
