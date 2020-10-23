package director

import "github.com/pkg/errors"

type Config struct {
	PageSize int `mapstructure:"page_size"`
}

func DefaultConfig() *Config {
	return &Config{
		PageSize: 100,
	}
}

func (c *Config) Validate() error {
	if c.PageSize < 1 {
		return errors.New("graphql page size must be a positive number")
	}

	return nil
}
