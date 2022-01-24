package tests

import (
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/pkg/errors"
	"os"
	"testing"

	pkgConfig "github.com/kyma-incubator/compass/tests/pkg/config"

	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

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
	staticUser       string
	dexGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	dexToken := server.Token()
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	claims := claims{}
	_, err = jwt.ParseWithClaims(dexToken, &claims, nil)
	if err != nil {
		panic(err)
	}

	staticUser = claims.Name
	fmt.Println("Static user from token:", staticUser)

	exitVal := m.Run()
	os.Exit(exitVal)
}

type claims struct {
	Name string `json:"name"`
}

func (c *claims) Valid() error {
	if c.Name == "" {
		return errors.New("missing \"name\" claim value")
	}
	return nil
}
