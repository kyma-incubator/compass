package metrics

import (
	"fmt"
	"time"
)

type Config struct {
	Port            int           `mapstructure:"port"`
	Timeout         time.Duration `mapstructure:"server_timeout"`
	ShutDownTimeout time.Duration `mapstructure:"shut_down_timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:            5002,
		Timeout:         time.Second * 115,
		ShutDownTimeout: time.Second * 10,
	}
}

func (s *Config) Validate() error {
	if s.Port <= 0 {
		return fmt.Errorf("validate Settings: Invalid Port")
	}
	if s.Timeout == 0 {
		return fmt.Errorf("validate Settings: Timeout missing")
	}
	if s.ShutDownTimeout == 0 {
		return fmt.Errorf("validate Settings: ShutDownTimeout missing")
	}
	return nil
}
