package env

import "github.com/vrischmann/envconfig"

var (
	Config EnvConfig
)

type EnvConfig struct {
	ServicePort int    `envconfig:"default=8000"`
	GraphqlURL  string `envconfig:"default=http://127.0.0.1:3000/graphql"`
	OIDC        struct {
		IssuerURL    string
		ClientID     string
		ClientSecret string
	}
}

func InitConfig() {
	err := envconfig.Init(&Config)
	if err != nil {
		panic(err)
	}
}
