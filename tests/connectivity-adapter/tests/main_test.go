package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

var (
	certCache                certloader.Cache
	certSecuredGraphQLClient *graphql.Client
)

type config struct {
	ConnectivityAdapterUrl         string `envconfig:"default=https://adapter-gateway.kyma.local"`
	ConnectivityAdapterMtlsUrl     string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
	DirectorUrl                    string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	DirectorExternalCertSecuredURL string `envconfig:"default=http://compass-director-external-mtls.compass-system.svc.cluster.local:3000/graphql"`
	SkipSSLValidation              bool   `envconfig:"default=true"`
	EventsBaseURL                  string `envconfig:"default=https://events.com"`
	Tenant                         string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	DirectorReadyzUrl              string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/readyz"`
	CertLoaderConfig               certloader.Config
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.InitWithPrefix(&testConfig, "APP")
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

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, certCache.Get().PrivateKey, cc.Get().Certificate, testConfig.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)
}
