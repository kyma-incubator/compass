package tests

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/subscription"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type DirectorConfig struct {
	BaseDirectorConfig
	DirectorUrl                    string
	HealthUrl                      string `envconfig:"default=https://director.kyma.local/healthz"`
	WebhookUrl                     string `envconfig:"default=https://kyma-project.io"`
	InfoUrl                        string `envconfig:"APP_INFO_API_ENDPOINT,default=https://director.kyma.local/v1/info"`
	CertIssuer                     string `envconfig:"APP_INFO_CERT_ISSUER"`
	CertSubject                    string `envconfig:"APP_INFO_CERT_SUBJECT"`
	DefaultScenarioEnabled         bool   `envconfig:"default=true"`
	DefaultNormalizationPrefix     string `envconfig:"default=mp-"`
	GatewayOauth                   string
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool   `envconfig:"default=false"`
	ConsumerID                     string `envconfig:"APP_INFO_CERT_CONSUMER_ID"`
	CertLoaderConfig               certloader.Config
	certprovider.ExternalCertProviderConfig
	SubscriptionConfig               subscription.Config
	TestProviderSubaccountID         string
	TestConsumerSubaccountID         string
	TestConsumerTenantID             string
	ExternalServicesMockBaseURL      string
	TokenPath                        string
	SubscriptionProviderAppNameValue string
	ConsumerSubaccountLabelKey       string
	SubscriptionLabelKey             string
	RuntimeTypeLabelKey              string
	KymaRuntimeTypeLabelValue        string
	ConsumerTokenURL                 string
	ProviderClientID                 string
	ProviderClientSecret             string
	BasicUsername                    string
	BasicPassword                    string
	ExternalCertTestCN               string
}

type BaseDirectorConfig struct {
	DefaultScenario string `envconfig:"default=DEFAULT"`
}

var (
	conf                     = &DirectorConfig{}
	certSecuredGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	config.ReadConfig(conf)

	ctx := context.Background()

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
