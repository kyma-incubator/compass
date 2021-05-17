package bench

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

var (
	conf               = &config.DirectorConfig{}
	dexGraphQLClient   *graphql.Client
)

func TestMain(m *testing.M) {
	dbCfg := persistence.DatabaseConfig{}
	err := envconfig.Init(&dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	config.ReadConfig(conf)

	log.Info("Get Dex id_token")
	/*dexToken, err := idtokenprovider.GetDexToken()
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting dex token"))
	}*/
	dexToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0."
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()

	os.Exit(exitVal)
}
