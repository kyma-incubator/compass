package config

type DirectorConfig struct {
	BaseDirectorConfig
	HealthUrl                  string `envconfig:"default=https://director.kyma.local/healthz"`
	WebhookUrl                 string `envconfig:"default=https://kyma-project.io"`
	InfoUrl                    string `envconfig:"APP_INFO_API_ENDPOINT,default=https://director.kyma.local/v1/info"`
	CertIssuer                 string `envconfig:"APP_INFO_CERT_ISSUER"`
	CertSubject                string `envconfig:"APP_INFO_CERT_SUBJECT"`
	DefaultScenarioEnabled     bool   `envconfig:"default=true"`
	DefaultNormalizationPrefix string `envconfig:"default=mp-"`
	GatewayOauth               string
}

type BaseDirectorConfig struct {
	DefaultScenario string `envconfig:"default=DEFAULT"`
}
