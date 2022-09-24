package tests

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var (
	conf                     = &config.PairingAdapterConfig{}
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	config.ReadConfig(conf)

	ctx := context.Background()

	tenant.TestTenants.Init()

	cc, err := certloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, conf.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
