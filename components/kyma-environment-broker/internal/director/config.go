package director

type Config struct {
	URL                        string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	Namespace                  string `envconfig:"default=compass-system"`
	OauthCredentialsSecretName string `envconfig:"default=compass-broker-registration"`

	DefaultTenant        string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	SkipCertVerification bool   `envconfig:"default=false"`
}
