package config

type ConnectivityAdapterTestConfig struct {
	ConnectivityAdapterUrl     string `envconfig:"default=https://adapter-gateway.kyma.local"`
	ConnectivityAdapterMtlsUrl string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
	DirectorUrl                string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipSslVerify              bool   `envconfig:"default=true"`
	EventsBaseURL              string `envconfig:"default=https://events.com"`
	Tenant                     string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	DirectorHealthzUrl         string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/healthz"`
}
