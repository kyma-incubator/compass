package oauth20

type Config struct {
	ClientCreationEndpoint string `envconfig:"default=/clients"` // TODO: Update it
	ClientDeletionEndpoint string `envconfig:"default=/clients"` // TODO: Update it
	PublicAccessTokenEndpoint string `envconfig:"default=https://oauth2.kyma.local/oauth2/token"`
}