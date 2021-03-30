package features

type Config struct {
	DefaultScenarioEnabled bool `mapstructure:"DEFAULT_SCENARIO_ENABLED"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultScenarioEnabled: true,
	}
}
