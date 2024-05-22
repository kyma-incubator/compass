package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	ExternalSvcMockURL             string `envconfig:"EXTERNAL_SERVICES_MOCK_BASE_URL"`
	SystemFetcherURL               string `envconfig:"SYSTEM_FETCHER_URL"`
	SystemFetcherPageSize          int    `envconfig:"SYSTEM_FETCHER_PAGE_SIZE"`
	SystemFetcherContainerName     string `envconfig:"SYSTEM_FETCHER_CONTAINER_NAME"`
	DirectorExternalCertSecuredURL string
	GatewayOauth                   string `envconfig:"APP_GATEWAY_OAUTH"`
	SkipSSLValidation              bool   `envconfig:"default=false"`
	CertLoaderConfig               credloader.CertConfig
	InternalDirectorGQLURL         string `envconfig:"INTERNAL_DIRECTOR_URL"`

	SelfRegDistinguishLabelKey   string
	SelfRegDistinguishLabelValue string
	SelfRegRegion                string
	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`

	SystemInformationSourceKey string
	TemplateLabelFilter        string
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

	ctx := context.Background()
	cc, err := credloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err = credloader.WaitForCertCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, cc.Get()[cfg.ExternalClientCertSecretName].PrivateKey, cc.Get()[cfg.ExternalClientCertSecretName].Certificate, cfg.SkipSSLValidation)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	saTransport := httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")
	
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
