package oauth20

type Config struct {
	ClientCreationEndpoint string `envconfig:"default="`
	AccessTokenEndpoint string `envconfig:"default=https://oauth2.kyma.local/oauth2/token"`
}