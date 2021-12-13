package tests

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

var dexGraphQLClient *graphql.Client

type config struct {
	ConnectivityAdapterUrl     string `envconfig:"default=https://adapter-gateway.kyma.local"`
	ConnectivityAdapterMtlsUrl string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
	DirectorUrl                string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipSSLValidation          bool   `envconfig:"default=true"`
	EventsBaseURL              string `envconfig:"default=https://events.com"`
	Tenant                     string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	DirectorReadyzUrl          string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/readyz"`
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.InitWithPrefix(&testConfig, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	dexToken := server.Token()
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)
}
