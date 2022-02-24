package bench

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

var (
	certSecuredGraphQLClient *graphql.Client
)

type config struct {
	DirectorURL                    string
	DirectorExternalCertSecuredURL string
	ORDServiceURL                  string
	ORDServiceDefaultResponseType  string
	SkipSSLValidation              bool
	CertLoaderConfig               certloader.Config
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, testConfig.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	for cc.Get() == nil {
		log.D().Info("Waiting for certificate cache to load, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, testConfig.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
