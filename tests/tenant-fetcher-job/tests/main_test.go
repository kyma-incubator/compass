package tests

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	ExternalSvcMockURL     string `envconfig:"EXTERNAL_SERVICES_MOCK_BASE_URL"`
	InternalDirectorGQLURL string `envconfig:"INTERNAL_DIRECTOR_URL"`
}

var (
	cfg                       Config
	dexGraphQLClient          *graphql.Client
	directorInternalGQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&cfg)
	if err != nil {
		log.D().Fatal(err)
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)
	dexGraphQLClient.Log = func(s string) {
		log.D().Info(s)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	saTransport := httputil.NewServiceAccountTokenTransportWithHeader(tr, "Authorization")
	client := &http.Client{
		Transport: saTransport,
		Timeout:   time.Second * 30,
	}
	directorInternalGQLClient = gcli.NewClient(cfg.InternalDirectorGQLURL, gcli.WithHTTPClient(client))
	directorInternalGQLClient.Log = func(s string) {
		log.D().Info(s)
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}
