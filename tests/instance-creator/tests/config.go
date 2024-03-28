package tests

import (
	"time"

	directorcfg "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
)

type InstanceCreatorConfig struct {
	DirectorInternalGatewayUrl     string
	GatewayOauth                   string
	CompassExternalMTLSGatewayURL  string `envconfig:"APP_COMPASS_EXTERNAL_MTLS_GATEWAY_URL"`
	DirectorExternalCertSecuredURL string
	CertLoaderConfig               credloader.CertConfig
	certprovider.ExternalCertProviderConfig
	SubscriptionConfig                  subscription.Config
	DestinationsConfig                  directorcfg.DestinationsConfig
	ProviderDestinationConfig           config.ProviderDestinationConfig
	TestProviderSubaccountID            string
	TestConsumerAccountID               string
	TestConsumerSubaccountID            string
	TestConsumerTenantID                string
	ExternalServicesMockBaseURL         string
	ExternalServicesMockMtlsSecuredURL  string
	TokenPath                           string
	SubscriptionProviderAppNameValue    string
	GlobalSubaccountIDLabelKey          string        `envconfig:"APP_GLOBAL_SUBACCOUNT_ID_LABEL_KEY"`
	SubscriptionProviderAppNameProperty string        `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_APP_NAME_PROPERTY"`
	CertSubjectMappingResyncInterval    time.Duration `envconfig:"APP_CERT_SUBJECT_MAPPING_RESYNC_INTERVAL"`
	InstanceCreatorRegion               string        `envconfig:"APP_INSTANCE_CREATOR_REGION,default=eu-1"`
	SkipSSLValidation                   bool          `envconfig:"default=false"`
}
