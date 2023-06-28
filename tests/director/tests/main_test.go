package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

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
	DirectorUrl                                 string
	DirectorInternalGatewayUrl                  string
	HealthUrl                                   string `envconfig:"default=https://director.kyma.local/healthz"`
	WebhookUrl                                  string `envconfig:"default=https://kyma-project.io"`
	InfoUrl                                     string `envconfig:"APP_INFO_API_ENDPOINT,default=https://director.kyma.local/v1/info"`
	DefaultNormalizationPrefix                  string `envconfig:"default=mp-"`
	GatewayOauth                                string
	DirectorExternalCertSecuredURL              string
	DirectorExternalCertFAAsyncStatusURL        string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASSIGNMENT_ASYNC_STATUS_URL"`
	DirectorExternalCertFormationAsyncStatusURL string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASYNC_STATUS_URL"`
	SkipSSLValidation                           bool   `envconfig:"default=false"`
	ConsumerID                                  string `envconfig:"APP_INFO_CERT_CONSUMER_ID"`
	CertLoaderConfig                            certloader.Config
	certprovider.ExternalCertProviderConfig
	SubscriptionConfig                              subscription.Config
	TestProviderAccountID                           string
	TestProviderSubaccountID                        string
	TestConsumerAccountID                           string
	TestConsumerSubaccountID                        string
	TestConsumerAccountIDTenantHierarchy            string
	TestConsumerSubaccountIDTenantHierarchy         string
	TestConsumerTenantID                            string
	TestProviderSubaccountIDRegion2                 string
	SelfRegisterSubdomainPlaceholderValue           string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REGISTER_SUBDOMAIN_PLACEHOLDER_VALUE"`
	ExternalServicesMockBaseURL                     string
	ExternalServicesMockMtlsSecuredURL              string
	TokenPath                                       string
	SubscriptionProviderAppNameValue                string
	ConsumerSubaccountLabelKey                      string
	SubscriptionLabelKey                            string
	RuntimeTypeLabelKey                             string
	ApplicationTypeLabelKey                         string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`
	KymaRuntimeTypeLabelValue                       string
	KymaApplicationNamespaceValue                   string
	SaaSAppNameLabelKey                             string `envconfig:"APP_SELF_REGISTER_SAAS_APP_LABEL_KEY,default=CMPSaaSAppName"`
	ConsumerTokenURL                                string
	ProviderClientID                                string
	ProviderClientSecret                            string
	BasicUsername                                   string
	BasicPassword                                   string
	CertSvcInstanceSecretName                       string        `envconfig:"CERT_SVC_INSTANCE_SECRET_NAME"`
	ExternalCertTestIntSystemOUSubaccount           string        `envconfig:"APP_EXTERNAL_CERT_TEST_INTEGRATION_SYSTEM_OU_SUBACCOUNT"`
	ExternalCertTestIntSystemCommonName             string        `envconfig:"APP_EXTERNAL_CERT_TEST_INTEGRATION_SYSTEM_CN"`
	ExternalClientCertExpectedIssuerLocalityRegion2 string        `envconfig:"APP_EXTERNAL_CLIENT_CERT_EXPECTED_ISSUER_LOCALITY_REGION2"`
	SupportedORDApplicationType                     string        `envconfig:"APP_SUPPORTED_ORD_APPLICATION_TYPE"`
	FormationMappingAsyncResponseDelay              int64         `envconfig:"APP_FORMATION_MAPPING_ASYNC_RESPONSE_DELAY"`
	SubscriptionProviderAppNameProperty             string        `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_APP_NAME_PROPERTY"`
	CertSubjectMappingResyncInterval                time.Duration `envconfig:"APP_CERT_SUBJECT_MAPPING_RESYNC_INTERVAL"`
	ApplicationTemplateProductLabel                 string        `envconfig:"APP_APPLICATION_TEMPLATE_PRODUCT_LABEL"`
	DefaultTenantRegion                             string        `envconfig:"APP_DEFAULT_TENANT_REGION,default=eu-1"`
}

var (
	conf                      = &DirectorConfig{}
	certSecuredGraphQLClient  *graphql.Client
	directorInternalGQLClient *graphql.Client
	cc                        certloader.Cache
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
