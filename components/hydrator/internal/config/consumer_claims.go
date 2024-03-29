package config

type ConsumerClaimsKeysConfig struct {
	ClientIDKey string `envconfig:"APP_CONSUMER_CLAIMS_CLIENT_ID_KEY"`
	TenantIDKey string `envconfig:"APP_CONSUMER_CLAIMS_TENANT_ID_KEY"`
	UserNameKey string `envconfig:"APP_CONSUMER_CLAIMS_USER_NAME_KEY"`
}
