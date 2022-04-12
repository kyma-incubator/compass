package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	ExternalSvcMockURL             string `envconfig:"EXTERNAL_SERVICES_MOCK_BASE_URL"`
	TenantFetcherContainerName     string `envconfig:"TENANT_FETCHER_CONTAINER_NAME"`
	InternalDirectorGQLURL         string `envconfig:"INTERNAL_DIRECTOR_URL"`
	DirectorExternalCertSecuredURL string ``
	SkipSSLValidation              bool   `envconfig:"default=false"`
	CertLoaderConfig               certloader.Config
}

var (
	cfg                       Config
	certSecuredGraphQLClient  *graphql.Client
	directorInternalGQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&cfg)
	if err != nil {
		log.D().Fatal(err)
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, cfg.SkipSSLValidation)
	certSecuredGraphQLClient.Log = func(s string) {
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
	directorInternalGQLClient = graphql.NewClient(cfg.InternalDirectorGQLURL, graphql.WithHTTPClient(client))
	directorInternalGQLClient.Log = func(s string) {
		log.D().Info(s)
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}
