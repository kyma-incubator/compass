package director

import (
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/pflag"

	"github.com/kyma-incubator/compass/components/director/pkg/env"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/features"
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

type ApiConfig struct {
	APIEndpoint                   string `mapstructure:"api_endpoint"`
	TenantMappingEndpoint         string `mapstructure:"tenant_mapping_endpoint"`
	RuntimeMappingEndpoint        string `mapstructure:"runtime_mapping_endpoint"`
	AuthenticationMappingEndpoint string `mapstructure:"authentication_mapping_endpoint`
	OperationPath                 string `mapstructure:"operation_path"`
	LastOperationPath             string `mapstructure:"last_operation_path"`
	PlaygroundAPIEndpoint         string `mapstructure:"playground_api_endpoint"`
}

func DefaultAPIConfig() *ApiConfig {
	return &ApiConfig{
		APIEndpoint:                   "/graphql",
		TenantMappingEndpoint:         "/tenant-mapping",
		RuntimeMappingEndpoint:        "/runtime-mapping",
		AuthenticationMappingEndpoint: "/authn-mapping",
		OperationPath:                 "/operation",
		LastOperationPath:             "/last_operation",
		PlaygroundAPIEndpoint:         "/graphql",
	}
}

func (s *ApiConfig) Validate() error {
	if len(s.APIEndpoint) == 0 {
		return fmt.Errorf("validate Settings: API endpoint missing")
	}
	if len(s.TenantMappingEndpoint) == 0 {
		return fmt.Errorf("validate Settings: tennant mapping endpoint missing")
	}
	if len(s.RuntimeMappingEndpoint) == 0 {
		return fmt.Errorf("validate Settings: runtime mapping endpoint missing")
	}
	if len(s.AuthenticationMappingEndpoint) == 0 {
		return fmt.Errorf("validate Settings: authentication mapping endpoint missing")
	}
	if len(s.OperationPath) == 0 {
		return fmt.Errorf("validate Settings: operation path missing")
	}
	if len(s.LastOperationPath) == 0 {
		return fmt.Errorf("validate Settings: last operation path missing")
	}
	if len(s.PlaygroundAPIEndpoint) == 0 {
		return fmt.Errorf("validate Settings: playground api endpoint missing")
	}
	return nil
}

type TimeoutsConfig struct {
	Client time.Duration `mapstructure:"client_timeout"`
	Server time.Duration `mapstructure:"server_timeout"`
}

func DefaultTimeoutsConfig() *TimeoutsConfig {
	return &TimeoutsConfig{
		Client: 105 * time.Second,
		Server: 110 * time.Second,
	}
}

func (s *TimeoutsConfig) Validate() error {
	if s.Client == 0 {
		return fmt.Errorf("validate Settings: Client timeout missing")
	}
	if s.Server == 0 {
		return fmt.Errorf("validate Settings: Server timeout missing")
	}
	return nil
}

type JWKSConfig struct {
	Endpoint            string        `mapstructure:"jwks_endpoint"`
	SyncPeriod          time.Duration `mapstructure:"jwks_sync_period"`
	AllowJWTSigningNone bool          `mapstructure:"allow_jwts_signing_none"`
	RuntimeCachePeriod  time.Duration `mapstructure:"runtime_jwks_cache_period"`
}

func DefaultJWKSConfig() *JWKSConfig {
	return &JWKSConfig{
		Endpoint:            "file://hack/default-jwks.json",
		SyncPeriod:          5 * time.Minute,
		AllowJWTSigningNone: true,
		RuntimeCachePeriod:  5 * time.Minute,
	}
}

func (s *JWKSConfig) Validate() error {
	if len(s.Endpoint) == 0 {
		return fmt.Errorf("validate Settings: JWKS endpoint missing")
	}
	if s.SyncPeriod <= 0 {
		return fmt.Errorf("validate Settings: JWKS sync period must be greater than zero")
	}
	if s.RuntimeCachePeriod <= 0 {
		return fmt.Errorf("validate Settings: JWKS runtime cache period must be greater than zero")
	}
	return nil
}

type StaticConfig struct {
	UsersSrc  string `mapstructure:"users_src"`
	GroupsSrc string `mapstructure:"groups_src"`
}

func DefaultStaticConfig() *StaticConfig {
	return &StaticConfig{
		UsersSrc:  "/data/static-users.yaml",
		GroupsSrc: "/data/static-groups.yaml",
	}
}

func (s *StaticConfig) Validate() error {
	if len(s.UsersSrc) == 0 {
		return fmt.Errorf("validate Settings: Static users source file missing")
	}
	if len(s.GroupsSrc) == 0 {
		return fmt.Errorf("validate Settings: Static groups source file missing")
	}
	return nil
}

type Config struct {
	Address                string `mapstructure:"address"`
	InternalGraphQLAddress string `mapstructure:"internal_graphql_address"`
	HydratorAddress        string `mapstructure:"hydrator_address"`

	InternalAddress string `mapstructure:"internal_address"`
	AppURL          string `mapstructure:"app_url"`

	Timeouts *TimeoutsConfig `mapstructure:",squash"`

	DB                      *persistence.DatabaseConfig `mapstructure:"db"`
	API                     *ApiConfig                  `mapstructure:",squash"`
	ConfigurationFile       string                      `mapstructure:"configuration_file"`
	ConfigurationFileReload time.Duration               `mapstructure:"configuration_file_reload"`

	Log *log.Config `mapstructure:"log"`

	MetricsAddress string `mapstructure:"metrics_address"`

	JWKS *JWKSConfig `mapstructure:",squash"`

	ClientIDHttpHeaderKey string `mapstructure:"client_id_http_header"`

	Static *StaticConfig `mapstructure:"static"`

	PairingAdapterSrc string `mapstructure:"pairing_adapter_src"`

	OneTimeToken *onetimetoken.Config
	OAuth20      *oauth20.Config `mapstructure:"oauth20"`

	Features *features.Config

	ProtectedLabelPattern string `mapstructure:"protected_label_pattern"`
	OperationsNamespace   string `mapstructure:"operations_namespace"`

	DisableAsyncMode bool `mapstructure:"disable_async_mode"`
}

func DefaultConfig() *Config {
	return &Config{
		Address:                 "127.0.0.1:3000",
		InternalGraphQLAddress:  "127.0.0.1:3001",
		HydratorAddress:         "127.0.0.1:8080",
		InternalAddress:         "127.0.0.1:3002",
		AppURL:                  "",
		Timeouts:                DefaultTimeoutsConfig(),
		DB:                      persistence.DefaultDatabaseConfig(),
		API:                     DefaultAPIConfig(),
		ConfigurationFile:       "/config/config.yaml",
		ConfigurationFileReload: 1 * time.Minute,
		Log:                     log.DefaultConfig(),
		MetricsAddress:          "127.0.0.1:3003",
		JWKS:                    DefaultJWKSConfig(),
		ClientIDHttpHeaderKey:   "client_user",
		Static:                  DefaultStaticConfig(),
		PairingAdapterSrc:       "",
		OneTimeToken:            onetimetoken.DefaultConfig(),
		OAuth20:                 oauth20.DefaultConfig(),
		Features:                features.DefaultConfig(),
		ProtectedLabelPattern:   ".*_defaultEventing",
		OperationsNamespace:     "compass-system",
		DisableAsyncMode:        false,
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
