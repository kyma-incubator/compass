package oauth20

type Config struct {
	ClientEndpoint string `envconfig:"default=https://oauth2-admin.kyma.local/clients"`
	PublicAccessTokenEndpoint string `envconfig:"default=https://oauth2.kyma.local/oauth2/token"`
}