package features

type Config struct {
	DefaultScenarioEnabled bool `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
}
