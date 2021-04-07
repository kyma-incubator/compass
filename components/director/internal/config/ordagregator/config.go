package ordagregator

import (
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/pkg/env"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/spf13/pflag"
)

type Validatable interface {
	Validate() error
}

func AddPFlags(set *pflag.FlagSet) {
	env.CreatePFlags(set, DefaultConfig())
	env.CreatePFlagsForConfigFile(set)
}

type Config struct {
	DB *persistence.DatabaseConfig `mapstructure:"db"`

	Log *log.Config `mapstructure:"log"`

	Features *features.Config

	ConfigurationFile       string        `mapstructure:"configuration_file"`
	ConfigurationFileReload time.Duration `mapstructure:"configuration_file_reload"`

	ClientTimeout time.Duration `mapstructure:"client_timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		DB:                      persistence.DefaultDatabaseConfig(),
		Log:                     log.DefaultConfig(),
		Features:                features.DefaultConfig(),
		ConfigurationFileReload: 1 * time.Minute,
		ClientTimeout:           60 * time.Second,
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
