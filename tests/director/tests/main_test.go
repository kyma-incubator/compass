package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var (
	conf                      = &DirectorConfig{}
	certSecuredGraphQLClient  *graphql.Client
	directorInternalGQLClient *graphql.Client
	cc                        credloader.CertCache
)

func TestMain(m *testing.M) {
	tenant.TestTenants.Init()

	config.ReadConfig(conf)

	ctx := context.Background()

	if err := conf.DestinationsConfig.MapInstanceConfigs(); err != nil {
		log.D().Fatal(errors.Wrap(err, "while loading destination instances config"))
	}

	var err error
	cc, err = credloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err = credloader.WaitForCertCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, cc.Get()[conf.ExternalClientCertSecretName].PrivateKey, cc.Get()[conf.ExternalClientCertSecretName].Certificate, conf.SkipSSLValidation)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	saTransport := httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")
	client := &http.Client{
		Transport: saTransport,
		Timeout:   time.Second * 30,
	}
	directorInternalGQLClient = graphql.NewClient(conf.DirectorInternalGatewayUrl, graphql.WithHTTPClient(client))
	directorInternalGQLClient.Log = func(s string) {
		log.D().Info(s)
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}
