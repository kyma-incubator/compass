package oauth20

type Config struct {
	ClientEndpoint            string `envconfig:"APP_OAUTH20_CLIENT_ENDPOINT"`
	PublicAccessTokenEndpoint string `envconfig:"APP_OAUTH20_PUBLIC_ACCESS_TOKEN_ENDPOINT"`
}
