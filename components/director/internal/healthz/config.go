package healthz

import "time"

const DefaultName = "default"

type Config struct {
	Indicators []IndicatorConfig
}

// IndicatorConfig implements IndicatorConfig interface
type IndicatorConfig struct {
	Name         string        `envconfig:"default=default"`
	Interval     time.Duration `envconfig:"default=5s"`
	Timeout      time.Duration `envconfig:"default=1s"`
	InitialDelay time.Duration `envconfig:"default=1s"`
	Threshold    int           `envconfig:"default=3"`
}

func NewDefaultConfig() IndicatorConfig {
	return IndicatorConfig{
		Name:         DefaultName,
		Interval:     5 * time.Second,
		Timeout:      time.Second,
		InitialDelay: time.Second,
		Threshold:    3,
	}
}
