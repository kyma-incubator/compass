package config

import (
	"fmt"
	"reflect"
	"time"

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

type ApiConfig struct {
	APIEndpoint                   string `envconfig:"default=/graphql"`
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	RuntimeMappingEndpoint        string `envconfig:"default=/runtime-mapping"`
	AuthenticationMappingEndpoint string `envconfig:"default=/authn-mapping"`
	OperationPath                 string `envconfig:"default=/operation"`
	LastOperationPath             string `envconfig:"default=/last_operation"`
	PlaygroundAPIEndpoint         string `envconfig:"default=/graphql"`
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
	Client time.Duration `envconfig:"default=105s"`
	Server time.Duration `envconfig:"default=110s"`
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
	Endpoint            string        `envconfig:"default=file://hack/default-jwks.json"`
	SyncPeriod          time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone bool          `envconfig:"default=true"`
	RuntimeCachePeriod  time.Duration `envconfig:"default=5m"`
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
	UsersSrc  string `envconfig:"default=/data/static-users.yaml"`
	GroupsSrc string `envconfig:"default=/data/static-groups.yaml"`
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
	Address                string `envconfig:"default=127.0.0.1:3000"`
	InternalGraphQLAddress string `envconfig:"default=127.0.0.1:3001"`
	HydratorAddress        string `envconfig:"default=127.0.0.1:8080"`

	InternalAddress string `envconfig:"default=127.0.0.1:3002"`
	AppURL          string `envconfig:"APP_URL"`

	Timeouts *TimeoutsConfig

	Database                persistence.DatabaseConfig
	API                     *ApiConfig
	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	Log *log.Config

	MetricsAddress string `envconfig:"default=127.0.0.1:3003"`

	JWKS *JWKSConfig

	ClientIDHttpHeaderKey string `envconfig:"default=client_user,APP_CLIENT_ID_HTTP_HEADER"`

	Static *StaticConfig

	PairingAdapterSrc string `envconfig:"optional"`

	OneTimeToken onetimetoken.Config
	OAuth20      oauth20.Config

	Features features.Config

	ProtectedLabelPattern string `envconfig:"default=.*_defaultEventing"`
	OperationsNamespace   string `envconfig:"default=compass-system"`

	DisableAsyncMode bool `envconfig:"default=false"`
}

func DefaultConfig() *Config {
	return &Config{
		Address:                 "127.0.0.1:3000",
		InternalGraphQLAddress:  "127.0.0.1:3001",
		HydratorAddress:         "127.0.0.1:8080",
		InternalAddress:         "127.0.0.1:3002",
		Timeouts:                DefaultTimeoutsConfig(),
		API:                     DefaultAPIConfig(),
		ConfigurationFileReload: 1 * time.Minute,
		Log:                     log.DefaultConfig(),
		MetricsAddress:          "127.0.0.1:3003",
		JWKS:                    DefaultJWKSConfig(),
		ClientIDHttpHeaderKey:   "client_user,APP_CLIENT_ID_HTTP_HEADER",
		Static:                  DefaultStaticConfig(),
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
