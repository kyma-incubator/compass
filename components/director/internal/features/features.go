package features

type Config struct {
	DefaultScenarioEnabled   bool `envconfig:"default=true,APP_DEFAULT_SCENARIO_ENABLED"`
	NameNormalizationEnabled bool `envconfig:"default=false,APP_NAME_NORMALIZATION_ENABLED"`
}
