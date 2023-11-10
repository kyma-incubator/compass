package bench

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	cfg "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type config struct {
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool   `envconfig:"default=false"`
	ApplicationTypeLabelKey        string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`
	GatewayOauth                   string `envconfig:"APP_GATEWAY_OAUTH"`
	CertLoaderConfig               credloader.CertConfig
	ExternalClientCertSecretName   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

var (
	conf                     config
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	tenant.TestTenants.Init()

	cfg.ReadConfig(&conf)
	ctx := context.Background()

	cc, err := credloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := credloader.WaitForCertCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, cc.Get()[conf.ExternalClientCertSecretName].PrivateKey, cc.Get()[conf.ExternalClientCertSecretName].Certificate, conf.SkipSSLValidation)

	exitVal := m.Run()

	os.Exit(exitVal)
}
