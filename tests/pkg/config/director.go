package config

type DirectorConfig struct {
	WebhookUrl                 string `envconfig:"default=https://kyma-project.io"`
	DefaultScenario            string `envconfig:"default=DEFAULT"`
	DefaultScenarioEnabled     bool   `envconfig:"default=true"`
	DefaultNormalizationPrefix string `envconfig:"default=mp-"`
	GatewayOauth               string
}
