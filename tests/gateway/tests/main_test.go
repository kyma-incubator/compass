package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	DefaultTenant string
	Domain        string
	DirectorURL   string
}

var (
	testConfig       config
	dexGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}
	testConfig.DirectorURL = fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", testConfig.Domain)

	log.Info("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting dex token"))
	}

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)

}
