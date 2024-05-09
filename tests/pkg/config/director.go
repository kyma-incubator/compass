package config

// ProviderDestinationConfig holds a provider's destination service configuration
type ProviderDestinationConfig struct {
	ClientID     string `envconfig:"APP_PROVIDER_DESTINATION_CLIENT_ID"`
	ClientSecret string `envconfig:"APP_PROVIDER_DESTINATION_CLIENT_SECRET"`
	TokenURL     string `envconfig:"APP_PROVIDER_DESTINATION_TOKEN_URL"`
	TokenPath    string `envconfig:"APP_PROVIDER_DESTINATION_TOKEN_PATH"`
	ServiceURL   string `envconfig:"APP_PROVIDER_DESTINATION_SERVICE_URL"`
	Dependency   string `envconfig:"APP_PROVIDER_DESTINATION_DEPENDENCY"`
}

// ConsumerClaimsKeysConfig holds customer claims keys
type ConsumerClaimsKeysConfig struct {
	ClientIDKey  string `envconfig:"APP_CONSUMER_CLAIMS_CLIENT_ID_KEY"`
	TenantIDKey  string `envconfig:"APP_CONSUMER_CLAIMS_TENANT_ID_KEY"`
	UserNameKey  string `envconfig:"APP_CONSUMER_CLAIMS_USER_NAME_KEY"`
	SubdomainKey string `envconfig:"APP_CONSUMER_CLAIMS_SUBDOMAIN_KEY"`
}
