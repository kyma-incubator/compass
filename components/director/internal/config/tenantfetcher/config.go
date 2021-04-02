package tenantfetcher

import (
	"reflect"
	"time"

	"github.com/spf13/pflag"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/pkg/env"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type Validatable interface {
	Validate() error
}

func AddPFlags(set *pflag.FlagSet) {
	env.CreatePFlags(set, DefaultConfig())
	env.CreatePFlagsForConfigFile(set)
}

type Config struct {
	Database         *persistence.DatabaseConfig `mapstructure:"db"`
	KubernetesConfig *tenantfetcher.KubeConfig
	OAuthConfig      tenantfetcher.OAuth2Config
	APIConfig        tenantfetcher.APIConfig
	QueryConfig      tenantfetcher.QueryConfig
	FieldMapping     tenantfetcher.TenantFieldMapping

	Log *log.Config

	TenantProvider      string `mapstructure:"tenant_provider"`
	MetricsPushEndpoint string `mapstructure:"metrics_push_endpoint"` //optional

	ClientTimeout time.Duration `mapstructure:"client_timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Database:         persistence.DefaultDatabaseConfig(),
		KubernetesConfig: tenantfetcher.DefaultKubeConfig(),
		OAuthConfig:      tenantfetcher.OAuth2Config{},
		APIConfig:        tenantfetcher.APIConfig{},
		QueryConfig:      tenantfetcher.QueryConfig{},
		FieldMapping:     tenantfetcher.TenantFieldMapping{},
		Log:              log.DefaultConfig(),
		ClientTimeout:    60 * time.Second,
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
