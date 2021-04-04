package tenantloader

import (
	"reflect"

	"github.com/spf13/pflag"

	"github.com/kyma-incubator/compass/components/director/pkg/env"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type Validatable interface {
	Validate() error
}

func AddPFlags(set *pflag.FlagSet) {
	env.CreatePFlags(set, DefaultJobConfig())
	env.CreatePFlagsForConfigFile(set)
}

type JobConfig struct {
	DB  *persistence.DatabaseConfig `mapstructure:"db"`
	Log *log.Config
}

func DefaultJobConfig() *JobConfig {
	return &JobConfig{
		DB:  persistence.DefaultDatabaseConfig(),
		Log: log.DefaultConfig(),
	}
}

func New(env env.Environment) (*JobConfig, error) {
	config := DefaultJobConfig()
	if err := env.Unmarshal(config); err != nil {
		return nil, errors.Wrapf(err, "error loading cfg")
	}

	return config, nil
}

func (c *JobConfig) Validate() error {
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
