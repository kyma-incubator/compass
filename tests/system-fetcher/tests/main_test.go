package tests

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	ExternalSvcMockURL             string `envconfig:"EXTERNAL_SERVICES_MOCK_BASE_URL"`
	SystemFetcherPageSize          int    `envconfig:"SYSTEM_FETCHER_PAGE_SIZE"`
	SystemFetcherContainerName     string `envconfig:"SYSTEM_FETCHER_CONTAINER_NAME"`
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool `envconfig:"default=false"`
	CertLoaderConfig               certloader.Config

	SelfRegDistinguishLabelKey   string
	SelfRegDistinguishLabelValue string
	SelfRegRegion                string
	ExternalClientCertSecretName string `envconfig:"EXTERNAL_CLIENT_CERT_SECRET_NAME"`

}

var (
	cfg                      Config
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&cfg)
	if err != nil {
		log.D().Fatal(err)
	}

	tenant.TestTenants.Init()

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, cc.Get()[cfg.ExternalClientCertSecretName].PrivateKey, cc.Get()[cfg.ExternalClientCertSecretName].Certificate, cfg.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
