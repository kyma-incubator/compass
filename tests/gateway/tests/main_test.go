package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/server"
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

	var dexToken string

	tokenConfig := server.Config{}
	err = envconfig.InitWithPrefix(&tokenConfig, "APP")
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

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)

}
