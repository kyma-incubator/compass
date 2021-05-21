package config

type DirectorConfig struct {
	WebhookUrl                 string `envconfig:"default=https://kyma-project.io"`
	DefaultScenario            string `envconfig:"default=DEFAULT"`
	DefaultNormalizationPrefix string `envconfig:"default=mp-"`
	GatewayOauth               string
}
