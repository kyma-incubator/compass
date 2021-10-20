package features

// Config missing godoc
type Config struct {
	DefaultScenarioEnabled        bool   `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
	ProtectedLabelPattern         string `envconfig:"default=.*_defaultEventing,APP_PROTECTED_LABEL_PATTERN"`
	SubscriptionProviderLabelKey  string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountIDsLabelKey string `envconfig:"APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY,default=consumer_subaccount_ids"`
}
