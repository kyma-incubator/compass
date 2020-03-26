package avs

type Config struct {
	OauthTokenEndpoint string
	OauthUsername      string
	OauthPassword      string
	OauthClientId      string
	ApiEndpoint        string
	DefinitionType     string `envconfig:"default=BASIC"`
	ApiKey             string
}
