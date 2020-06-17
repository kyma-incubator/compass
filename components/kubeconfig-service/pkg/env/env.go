package env

import (
	"github.com/vrischmann/envconfig"
)

var (
	Config EnvConfig
)

type EnvConfig struct {
	Port struct {
		Service int `envconfig:"default=8000"`
		Health  int `envconfig:"default=9000"`
	}
	GraphqlURL string `envconfig:"default=http://127.0.0.1:3000/graphql"`
	OIDC       struct {
		Kubeconfig struct {
			IssuerURL    string
			ClientID     string
			ClientSecret string
		}
		IssuerURL string
		ClientID  string
		CA        string `envconfig:"optional"`
		Claim     struct {
			Username string `envconfig:"default=email"`
			Groups   string `envconfig:"default=groups"`
		}
		Prefix struct {
			Username string `envconfig:"optional"`
			Groups   string `envconfig:"optional"`
		}
		SupportedSigningAlgs []string `envconfig:"default=RS256"`
	}
}

func InitConfig() {
	err := envconfig.Init(&Config)
	if err != nil {
		panic(err)
	}
}
