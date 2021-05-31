package tests

import (
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
	DefaultTenant               string
	DirectorURL                 string
	ExternalServicesMockBaseURL string
	BasicCredentialsUsername    string
	BasicCredentialsPassword    string
	AppClientID                 string
	AppClientSecret             string
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
