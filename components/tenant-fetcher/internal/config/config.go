package config

import (
	"reflect"
	"time"

	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant"

	"github.com/spf13/pflag"

	"github.com/kyma-incubator/compass/components/director/pkg/env"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type Validatable interface {
	Validate() error
}

func AddPFlags(set *pflag.FlagSet) {
	env.CreatePFlags(set, DefaultConfig())
	env.CreatePFlagsForConfigFile(set)
}

type Config struct {
	Address string `mapstructure:"address"`

	ServerTimeout   time.Duration `mapstructure:"server_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	Log *log.Config

	RootAPI string `mapstructure:"root_api"`

	Handler *tenant.Config

	DB *persistence.DatabaseConfig `mapstructure:"db"`
}

func DefaultConfig() *Config {
	return &Config{
		Address:         "127.0.0.1:8080",
		ServerTimeout:   110 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		Log:             log.DefaultConfig(),
		RootAPI:         "/tenants",
		Handler:         tenant.DefaultConfig(),
		DB:              persistence.DefaultDatabaseConfig(),
	}
}

func New(env env.Environment) (*Config, error) {
	config := DefaultConfig()
	if err := env.Unmarshal(config); err != nil {
		return nil, errors.Wrapf(err, "error loading cfg")
	}

	return config, nil
}

func (c *Config) Validate() error {
	validatableFields := make([]Validatable, 0, 0)

	v := reflect.ValueOf(*c)
	for i := 0; i < v.NumField(); i++ {
		field, ok := v.Field(i).Interface().(Validatable)
		if ok {
			validatableFields = append(validatableFields, field)
		}
	}

	for _, item := range validatableFields {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}
