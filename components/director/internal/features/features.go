package features

// Config missing godoc
type Config struct {
	ProtectedLabelPattern         string `envconfig:"default=.*_defaultEventing"`
	ImmutableLabelPattern         string `envconfig:"APP_SELF_REGISTER_LABEL_KEY_PATTERN,default=^xsappnameCMPClone$|^runtimeType$|^CMPSaaSAppName$"`
	SubscriptionProviderLabelKey  string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountLabelKey    string `envconfig:"APP_CONSUMER_SUBACCOUNT_LABEL_KEY,default=global_subaccount_id"`
	TokenPrefix                   string `envconfig:"APP_TOKEN_PREFIX,default=sb-"`
	RuntimeTypeLabelKey           string `envconfig:"APP_RUNTIME_TYPE_LABEL_KEY,default=runtimeType"`
	ApplicationTypeLabelKey       string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`
	KymaRuntimeTypeLabelValue     string `envconfig:"APP_KYMA_RUNTIME_TYPE_LABEL_VALUE,default=kyma"`
	KymaApplicationNamespaceValue string `envconfig:"APP_KYMA_APPLICATION_NAMESPACE_VALUE,default=sap.kyma"`
	KymaAdapterURL                string `envconfig:"APP_KYMA_ADAPTER_URL,default=https://compass-gateway-sap-mtls.local.kyma.dev:443"`
}
