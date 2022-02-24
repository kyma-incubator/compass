package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	config "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

var (
	conf = &config.IstioConfig{}

	certCache                certloader.Cache
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	dbCfg := persistence.DatabaseConfig{}
	err := envconfig.Init(&dbCfg)
	if err != nil {
		log.D().Fatal(err)
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	testctx.Tc, err = testctx.NewTestContext()
	if err != nil {
		log.D().Fatal(err)
	}

	config.ReadConfig(conf)

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	for cc.Get() == nil {
		log.D().Info("Waiting for certificate cache to load, sleeping for 1 second")
		time.Sleep(1 * time.Second)
	}
	certCache = cc

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, certCache.Get().PrivateKey, cc.Get().Certificate, conf.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
