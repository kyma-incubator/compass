package bench

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

var (
	certSecuredGraphQLClient *graphql.Client
)

type config struct {
	DirectorExternalCertSecuredURL string
	ORDServiceURL                  string
	ORDServiceDefaultResponseType  string
	SkipSSLValidation              bool
	CertLoaderConfig               credloader.CertConfig
	ExternalClientCertSecretName   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()

	ctx := context.Background()
	cc, err := credloader.StartCertLoader(ctx, testConfig.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := credloader.WaitForCertCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, cc.Get()[testConfig.ExternalClientCertSecretName].PrivateKey, cc.Get()[testConfig.ExternalClientCertSecretName].Certificate, testConfig.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
