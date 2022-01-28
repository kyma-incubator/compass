package tests

import (
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	ExternalServicesMockURL string
	ClientID                string
	ClientSecret            string
	DefaultTestTenant       string
	Domain                  string
	DirectorURL             string
	AdapterURL              string
	IsLocalEnv              bool
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

	tenant.TestTenants.Init()

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func getToken(t *testing.T,) string {
	if testConfig.IsLocalEnv {
		return getTokenFromExternalSVCMock(t)
	}

	return getTokenFromClone(t)
}

func getTokenFromExternalSVCMock(t *testing.T,) string {
	claims := map[string]interface{}{
		"ns-adapter-test": "ns-adapter-flow",
		"ext_attr": map[string]interface{}{
			"subaccountid": "08b6da37-e911-48fb-a0cb-fa635a6c4321",
		},
		"scope":     []string{},
		"client_id": "test_client_id",
		"tenant":    testConfig.DefaultTestTenant,
		"identity":  "nsadapter-flow-identity",
		"iss":       testConfig.ExternalServicesMockURL,
		"exp":       time.Now().Unix() + int64(time.Minute.Seconds()*10),
	}
	return token.FromExternalServicesMock(t, testConfig.ExternalServicesMockURL, testConfig.ClientID, testConfig.ClientSecret, claims)
}

func getTokenFromClone(t *testing.T,) string {
return ""
}
