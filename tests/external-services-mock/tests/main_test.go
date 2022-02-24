package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	pkgConfig "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Auditlog                           pkgConfig.AuditlogConfig
	DefaultTestTenant                  string
	DirectorURL                        string
	DirectorExternalCertSecuredURL     string
	ExternalServicesMockBaseURL        string
	ExternalServicesMockMTLSSecuredURL string `envconfig:"EXTERNAL_SERVICES_MOCK_MTLS_SECURED_URL"`
	BasicCredentialsUsername           string
	BasicCredentialsPassword           string
	AppClientID                        string
	AppClientSecret                    string
	SkipSSLValidation                  bool
	CertLoaderConfig                   certloader.Config
	ConsumerID                         string
}

var (
	testConfig               config
	consumerID               string
	certCache                certloader.Cache
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, testConfig.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	for cc.Get() == nil {
		log.D().Info("Waiting for certificate cache to load, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}
	certCache = cc

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, certCache.Get().PrivateKey, certCache.Get().Certificate, testConfig.SkipSSLValidation)

	consumerID = testConfig.ConsumerID

	exitVal := m.Run()
	os.Exit(exitVal)
}
