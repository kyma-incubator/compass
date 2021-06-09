package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	DefaultTestTenant           string
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

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)
}
