package healthz

import "time"

const DefaultName = "default"

type Config struct {
	Indicators []IndicatorConfig
}

// IndicatorConfig implements IndicatorConfig interface
type IndicatorConfig struct {
	Name         string        `envconfig:"default=default"`
	Interval     time.Duration `envconfig:"default=3s"`
	Timeout      time.Duration `envconfig:"default=1s"`
	InitialDelay time.Duration `envconfig:"default=0s"`
	Threshold    int           `envconfig:"default=3"`
}

func NewDefaultConfig() IndicatorConfig {
	return IndicatorConfig{
		Name:         DefaultName,
		Interval:     3 * time.Second,
		Timeout:      1 * time.Second,
		InitialDelay: time.Second,
		Threshold:    3,
	}
}
