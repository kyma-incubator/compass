package avs

type Config struct {
	OauthTokenEndpoint     string
	OauthUsername          string
	OauthPassword          string
	OauthClientId          string
	ApiEndpoint            string
	DefinitionType         string `envconfig:"default=BASIC"`
	ApiKey                 string
	Disabled               bool `envconfig:"default=false"`
	InternalTesterAccessId int64
	GroupId                int64
	ExternalTesterAccessId int64
}
