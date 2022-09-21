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
	DirectorUrl                    string
	HealthUrl                      string `envconfig:"default=https://director.kyma.local/healthz"`
	WebhookUrl                     string `envconfig:"default=https://kyma-project.io"`
	InfoUrl                        string `envconfig:"APP_INFO_API_ENDPOINT,default=https://director.kyma.local/v1/info"`
	DefaultNormalizationPrefix     string `envconfig:"default=mp-"`
	GatewayOauth                   string
	DirectorExternalCertSecuredURL string
	SkipSSLValidation              bool   `envconfig:"default=false"`
	ConsumerID                     string `envconfig:"APP_INFO_CERT_CONSUMER_ID"`
	CertLoaderConfig               certloader.Config
	certprovider.ExternalCertProviderConfig
	SubscriptionConfig                              subscription.Config
	TestProviderAccountID                           string
	TestProviderSubaccountID                        string
	TestConsumerAccountID                           string
	TestConsumerSubaccountID                        string
	TestConsumerTenantID                            string
	TestProviderSubaccountIDRegion2                 string
	ExternalServicesMockBaseURL                     string
	ExternalServicesMockMtlsSecuredURL              string
	TokenPath                                       string
	SubscriptionProviderAppNameValue                string
	ConsumerSubaccountLabelKey                      string
	SubscriptionLabelKey                            string
	RuntimeTypeLabelKey                             string
	ApplicationTypeLabelKey                         string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`
	KymaRuntimeTypeLabelValue                       string
	SaaSAppNameLabelKey                             string `envconfig:"APP_SELF_REGISTER_SAAS_APP_LABEL_KEY,default=CMPSaaSAppName"`
	ConsumerTokenURL                                string
	ProviderClientID                                string
	ProviderClientSecret                            string
	BasicUsername                                   string
	BasicPassword                                   string
	ExternalCertCommonName                          string `envconfig:"EXTERNAL_CERT_COMMON_NAME"`
	CertSvcInstanceTestIntSystemSecretName          string `envconfig:"CERT_SVC_INSTANCE_TEST_INTEGRATION_SYSTEM_SECRET_NAME"`
	ExternalCertTestIntSystemOUSubaccount           string `envconfig:"APP_EXTERNAL_CERT_TEST_INTEGRATION_SYSTEM_OU_SUBACCOUNT"`
	ExternalCertTestIntSystemCommonName             string `envconfig:"APP_EXTERNAL_CERT_TEST_INTEGRATION_SYSTEM_CN"`
	ExternalClientCertExpectedIssuerLocalityRegion2 string `envconfig:"APP_EXTERNAL_CLIENT_CERT_EXPECTED_ISSUER_LOCALITY_REGION2"`
	SupportedORDApplicationType                     string `envconfig:"APP_SUPPORTED_ORD_APPLICATION_TYPE"`
}

var (
	conf                     = &DirectorConfig{}
	certSecuredGraphQLClient *graphql.Client
	cc                       certloader.Cache
)

func TestMain(m *testing.M) {
	tenant.TestTenants.Init()

	config.ReadConfig(conf)

	ctx := context.Background()

	var err error
	cc, err = certloader.StartCertLoader(ctx, conf.CertLoaderConfig)
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
