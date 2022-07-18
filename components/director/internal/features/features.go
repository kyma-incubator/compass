package features

// Config missing godoc
type Config struct {
	DefaultScenarioEnabled       bool   `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
	ProtectedLabelPattern        string `envconfig:"default=.*_defaultEventing"`
	ImmutableLabelPattern        string `envconfig:"APP_SELF_REGISTER_LABEL_KEY_PATTERN,default=^xsappnameCMPClone$|^runtimeType$"`
	SubscriptionProviderLabelKey string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountLabelKey   string `envconfig:"APP_CONSUMER_SUBACCOUNT_LABEL_KEY,default=consumer_subaccount_id"`
	GlobalSubaccountLabelKey     string `envconfig:"APP_GLOBAL_SUBACCOUNT_LABEL_KEY,default=global_subaccount_id"`
	TokenPrefix                  string `envconfig:"APP_TOKEN_PREFIX,default=sb-"`
	RuntimeTypeLabelKey          string `envconfig:"APP_RUNTIME_TYPE_LABEL_KEY,default=runtimeType"`
	KymaRuntimeTypeLabelValue    string `envconfig:"APP_KYMA_RUNTIME_TYPE_LABEL_VALUE,default=kyma"`
}
