package config

type DirectorConfig struct {
	HealthUrl                  string `envconfig:"default=https://director.kyma.local/healthz"`
	WebhookUrl                 string `envconfig:"default=https://kyma-project.io"`
	DefaultScenario            string `envconfig:"default=DEFAULT"`
	DefaultScenarioEnabled     bool   `envconfig:"default=true"`
	DefaultNormalizationPrefix string `envconfig:"default=mp-"`
	GatewayOauth               string
}
