package subscription

type Config struct {
	URL                                   string `envconfig:"APP_SUBSCRIPTION_CONFIG_URL"`
	TokenURL                              string `envconfig:"APP_SUBSCRIPTION_CONFIG_TOKEN_URL"`
	ClientID                              string `envconfig:"APP_SUBSCRIPTION_CONFIG_CLIENT_ID"`
	ClientSecret                          string `envconfig:"APP_SUBSCRIPTION_CONFIG_CLIENT_SECRET"`
	SelfRegDistinguishLabelKey            string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REG_DISTINGUISH_LABEL_KEY"`
	SelfRegDistinguishLabelValue          string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REG_DISTINGUISH_LABEL_VALUE"`
	SelfRegRegion                         string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REG_REGION"`
	SelfRegRegion2                        string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REG_REGION2"`
	SelfRegisterSubdomainPlaceholderValue string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REGISTER_SUBDOMAIN_PLACEHOLDER_VALUE"`
	SelfRegisterLabelKey                  string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REGISTER_LABEL_KEY"`
	SelfRegisterLabelValuePrefix          string `envconfig:"APP_SUBSCRIPTION_CONFIG_SELF_REGISTER_LABEL_VALUE_PREFIX"`
	PropagatedProviderSubaccountHeader    string `envconfig:"APP_SUBSCRIPTION_CONFIG_PROPAGATED_PROVIDER_SUBACCOUNT_HEADER"`
	SubscriptionsLabelKey                 string `envconfig:"APP_SUBSCRIPTION_CONFIG_SUBSCRIPTIONS_LABEL_KEY"`
	// StandardFlow subscribes to saas-instance which has CMP declared as dependency.
	// The participants in the scenario and their relationships are: saas-instance -> CMP
	StandardFlow string `envconfig:"APP_SUBSCRIPTION_CONFIG_STANDARD_FLOW"`
	// DirectDependencyFlow is used in subscription tests for subscribing to saas-direct-dependency-instance which has CMP declared as dependencya.
	// The participants in the scenario and their relationships are saas-root-instance -> saas-direct-dependency-instance -> CMP
	DirectDependencyFlow string `envconfig:"APP_SUBSCRIPTION_CONFIG_DIRECT_DEPENDENCY_FLOW"`
	// IndirectDependencyFlow is used in subscription tests for subscribing to saas-root-instance that have CMP as indirect dependency.
	// The participants in the scenario and their relationships are saas-root-instance -> saas-direct-dependency-instance -> CMP.
	// As saas-root-instance has saas-direct-dependency-instance declared as dependency when subscribing to saas-root-instance, saas-direct-dependency-instance
	// will also receive a subscription callback.
	IndirectDependencyFlow    string `envconfig:"APP_SUBSCRIPTION_CONFIG_INDIRECT_DEPENDENCY_FLOW"`
	SubscriptionFlowHeaderKey string `envconfig:"APP_SUBSCRIPTION_CONFIG_SUBSCRIPTION_FLOW_HEADER_KEY"`
}

type TenantFetcherConfig struct {
	URL                     string `envconfig:"APP_TF_CONFIG_URL"`
	RootAPI                 string `envconfig:"APP_TF_CONFIG_ROOT_API"`
	RegionalHandlerEndpoint string `envconfig:"APP_TF_CONFIG_REGIONAL_HANDLER_ENDPOINT"`
	TenantPathParam         string `envconfig:"APP_TF_CONFIG_TENANT_PATH_PARAM"`
	RegionPathParam         string `envconfig:"APP_TF_CONFIG_REGION_PATH_PARAM"`
	FullRegionalURL         string `envconfig:"-"`
}
