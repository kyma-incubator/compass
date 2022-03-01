package tests

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	pkgConfig "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"os"
	"testing"
)

type config struct {
	Auditlog                           pkgConfig.AuditlogConfig
	DefaultTestTenant                  string
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

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, testConfig.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
