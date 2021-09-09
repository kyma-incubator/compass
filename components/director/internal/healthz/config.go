package healthz

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// DefaultName missing godoc
const DefaultName = "default"

// IndicatorConfig missing godoc
type IndicatorConfig struct {
	Name         string        `envconfig:"default=default"`
	Interval     time.Duration `envconfig:"default=5s"`
	Timeout      time.Duration `envconfig:"default=1s"`
	InitialDelay time.Duration `envconfig:"default=1s"`
	Threshold    int           `envconfig:"default=3"`
}

// NewDefaultConfig missing godoc
func NewDefaultConfig() IndicatorConfig {
	return IndicatorConfig{
		Name:         DefaultName,
		Interval:     5 * time.Second,
		Timeout:      time.Second,
		InitialDelay: time.Second,
		Threshold:    3,
	}
}

// Validate missing godoc
func (ic *IndicatorConfig) Validate() error {
	if ic.Interval <= 0 {
		return errors.New("interval could not be <= 0")
	}
	if ic.Timeout <= 0 {
		return errors.New("timeout could not be <= 0")
	}
	if ic.InitialDelay < 0 {
		return errors.New("initial delay could not be < 0")
	}
	if ic.Threshold < 0 {
		return errors.New("threshold could not be < 0")
	}
	return nil
}

// Config missing godoc
type Config struct {
	Indicators []IndicatorConfig
}

// Validate missing godoc
func (c *Config) Validate() error {
	for _, ind := range c.Indicators {
		if err := ind.Validate(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s indicator", ind.Name))
		}
	}
	return nil
}
