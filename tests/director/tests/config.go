package tests

import (
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/authenticator"

	directorcfg "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
)

type DirectorConfig struct {
	DirectorUrl                                       string
	DirectorInternalGatewayUrl                        string
	HealthUrl                                         string `envconfig:"default=https://director.kyma.local/healthz"`
	WebhookUrl                                        string `envconfig:"default=https://kyma-project.io"`
	InfoUrl                                           string `envconfig:"APP_INFO_API_ENDPOINT,default=https://director.kyma.local/v1/info"`
	DefaultNormalizationPrefix                        string `envconfig:"default=mp-"`
	GatewayOauth                                      string
	CompassExternalMTLSGatewayURL                     string `envconfig:"APP_COMPASS_EXTERNAL_MTLS_GATEWAY_URL"`
	DirectorUserNameAuthenticatorURL                  string `envconfig:"APP_DIRECTOR_USER_NAME_AUTHENTICATOR_URL"`
	DirectorExternalCertSecuredURL                    string
	DirectorExternalCertFAAsyncStatusURL              string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASSIGNMENT_ASYNC_STATUS_URL"`
	DirectorExternalCertFAAsyncStatusExternalTokenURL string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASSIGNMENT_ASYNC_STATUS_EXTERNAL_TOKEN_URL"`
	DirectorExternalCertFAAsyncResetStatusURL         string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASSIGNMENT_ASYNC_RESET_STATUS_URL"`
	DirectorExternalCertFormationAsyncStatusURL       string `envconfig:"APP_DIRECTOR_EXTERNAL_CERT_FORMATION_ASYNC_STATUS_URL"`
	ORDExternalCertSecuredServiceURL                  string `envconfig:"APP_ORD_EXTERNAL_CERT_SECURED_SERVICE_URL"`
	SkipSSLValidation                                 bool   `envconfig:"default=false"`
	ConsumerID                                        string `envconfig:"APP_INFO_CERT_CONSUMER_ID"`
	UsernameAuthCfg                                   authenticator.Config
	CertLoaderConfig                                  credloader.CertConfig
	certprovider.ExternalCertProviderConfig
	SubscriptionConfig                                 subscription.Config
	DestinationAPIConfig                               clients.DestinationServiceAPIConfig
	DestinationsConfig                                 directorcfg.DestinationsConfig
	ProviderDestinationConfig                          config.ProviderDestinationConfig
	DestinationConsumerSubdomainMtls                   string `envconfig:"APP_DESTINATION_CONSUMER_SUBDOMAIN_MTLS"`
	DestinationConsumerSubdomain                       string `envconfig:"APP_DESTINATION_CONSUMER_SUBDOMAIN"`
	TestDestinationInstanceID                          string `envconfig:"APP_TEST_DESTINATION_INSTANCE_ID"`
	TestCostObjectID                                   string
	TestProviderAccountID                              string
	TestProviderSubaccountID                           string
	TestConsumerAccountID                              string
	TestConsumerSubaccountID                           string
	TestConsumerAccountIDTenantHierarchy               string
	TestConsumerSubaccountIDTenantHierarchy            string
	TestConsumerTenantID                               string
	TestProviderSubaccountIDRegion2                    string
	SelfRegisterDirectDependencyDistinguishLabelValue  string `envconfig:"APP_SELF_REG_DIRECT_DEPENDENCY_DISTINGUISH_LABEL_VALUE"`
	SelfRegisterSubdomainPlaceholderValue              string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REGISTER_SUBDOMAIN_PLACEHOLDER_VALUE"`
	ExternalServicesMockBaseURL                        string
	ExternalServicesMockMtlsSecuredURL                 string
	TokenPath                                          string
	SubscriptionProviderAppNameValue                   string
	IndirectDependencySubscriptionProviderAppNameValue string `envconfig:"APP_INDIRECT_DEPENDENCY_SUBSCRIPTION_PROVIDER_APP_NAME_VALUE"`
	DirectDependencySubscriptionProviderAppNameValue   string `envconfig:"APP_DIRECT_DEPENDENCY_SUBSCRIPTION_PROVIDER_APP_NAME_VALUE"`
	GlobalSubaccountIDLabelKey                         string `envconfig:"APP_GLOBAL_SUBACCOUNT_ID_LABEL_KEY"`
	SubscriptionLabelKey                               string
	RuntimeTypeLabelKey                                string
	ApplicationTypeLabelKey                            string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`
	KymaRuntimeTypeLabelValue                          string
	KymaApplicationNamespaceValue                      string
	SaaSAppNameLabelKey                                string `envconfig:"APP_SELF_REGISTER_SAAS_APP_LABEL_KEY,default=CMPSaaSAppName"`
	ConsumerTokenURL                                   string
	ProviderClientID                                   string
	ProviderClientSecret                               string
	BasicUsername                                      string
	BasicPassword                                      string
	CertSvcInstanceSecretName                          string        `envconfig:"CERT_SVC_INSTANCE_SECRET_NAME"`
	ExternalCertTestIntSystemOUSubaccount              string        `envconfig:"APP_EXTERNAL_CERT_TEST_INTEGRATION_SYSTEM_OU_SUBACCOUNT"`
	ExternalCertTestIntSystemCommonName                string        `envconfig:"APP_EXTERNAL_CERT_TEST_INTEGRATION_SYSTEM_CN"`
	ExternalClientCertExpectedIssuerLocalityRegion2    string        `envconfig:"APP_EXTERNAL_CLIENT_CERT_EXPECTED_ISSUER_LOCALITY_REGION2"`
	SupportedORDApplicationType                        string        `envconfig:"APP_SUPPORTED_ORD_APPLICATION_TYPE"`
	TenantMappingAsyncResponseDelay                    int64         `envconfig:"APP_TENANT_MAPPING_ASYNC_RESPONSE_DELAY"`
	SubscriptionProviderAppNameProperty                string        `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_APP_NAME_PROPERTY"`
	CertSubjectMappingResyncInterval                   time.Duration `envconfig:"APP_CERT_SUBJECT_MAPPING_RESYNC_INTERVAL"`
	ApplicationTemplateProductLabel                    string        `envconfig:"APP_APPLICATION_TEMPLATE_PRODUCT_LABEL"`
	DefaultTenantRegion                                string        `envconfig:"APP_DEFAULT_TENANT_REGION,default=eu-1"`
	ORDServiceDefaultResponseType                      string        `envconfig:"APP_ORD_SERVICE_DEFAULT_RESPONSE_TYPE"`
}
