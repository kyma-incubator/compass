package tests

import (
	"os"
	"testing"

	pkgConfig "github.com/kyma-incubator/compass/tests/pkg/config"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	Auditlog                           pkgConfig.AuditlogConfig
	DefaultTestTenant                  string
	DirectorURL                        string
	ExternalServicesMockBaseURL        string
	ExternalServicesMockMTLSSecuredURL string `envconfig:"EXTERNAL_SERVICES_MOCK_MTLS_SECURED_URL"`
	BasicCredentialsUsername           string
	BasicCredentialsPassword           string
	AppClientID                        string
	AppClientSecret                    string
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
