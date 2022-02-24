package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	ExternalSvcMockURL             string `envconfig:"EXTERNAL_SERVICES_MOCK_BASE_URL"`
	SystemFetcherPageSize          int    `envconfig:"SYSTEM_FETCHER_PAGE_SIZE"`
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool `envconfig:"default=false"`
	CertLoaderConfig               certloader.Config
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
	defer tenant.TestTenants.Cleanup()

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	for cc.Get() == nil {
		log.D().Info("Waiting for certificate cache to load, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, cfg.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
