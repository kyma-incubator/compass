package tests

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	config "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var (
	conf = &config.IstioConfig{}

	certCache                credloader.CertCache
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	var err error
	testctx.Tc, err = testctx.NewTestContext()
	if err != nil {
		log.D().Fatal(err)
	}

	config.ReadConfig(conf)

	ctx := context.Background()
	certCache, err = credloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err = credloader.WaitForCertCache(certCache); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, certCache.Get()[conf.ExternalClientCertSecretName].PrivateKey, certCache.Get()[conf.ExternalClientCertSecretName].Certificate, conf.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
