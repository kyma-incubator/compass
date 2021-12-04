package features

// Config missing godoc
type Config struct {
	DefaultScenarioEnabled        bool   `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
	ProtectedLabelPattern         string `envconfig:"default=.*_defaultEventing|^consumer_subaccount_ids$"`
	ImmutableLabelPattern         string `envconfig:"APP_SELF_REGISTER_LABEL_KEY,optional"`
	SubscriptionProviderLabelKey  string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountIDsLabelKey string `envconfig:"APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY,default=consumer_subaccount_ids"`
}
