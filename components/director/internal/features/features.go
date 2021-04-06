package features

type Config struct {
	DefaultScenarioEnabled bool `mapstructure:"default_scenario_enabled"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultScenarioEnabled: true,
	}
}
