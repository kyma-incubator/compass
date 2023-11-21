package config

import (
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/server"
	pkgconfig "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/vrischmann/envconfig"
	"strings"
	"time"
)

const envPrefix = "APP"

type Config struct {
	Server            *server.Config
	Log               *log.Config
	HTTPClient        HTTPClient
	ServiceManagerCfg ServiceManagerConfig
}

type HTTPClient struct {
	Timeout           time.Duration `envconfig:"APP_TM_ADAPTER_CLIENT_TIMEOUT"`
	SkipSSLValidation bool          `envconfig:"APP_TM_ADAPTER_CLIENT_SKIP_SSL_VALIDATION"`
}

type ServiceManagerConfig struct {
	URL                      string                    `envconfig:"-"`
	InstancesSecretPath      string                    `envconfig:"APP_SM_INSTANCES_SECRET_PATH"`
	InstanceClientIDPath     string                    `envconfig:"APP_SM_INSTANCE_CLIENT_ID_PATH"`
	InstanceClientSecretPath string                    `envconfig:"APP_SM_INSTANCE_CLIENT_SECRET_PATH"`
	InstanceURLPath          string                    `envconfig:"APP_SM_INSTANCE_URL_PATH"`
	InstanceTokenURLPath     string                    `envconfig:"APP_SM_INSTANCE_TOKEN_URL_PATH"`
	InstanceOAuthTokenPath   string                    `envconfig:"APP_SM_INSTANCE_OAUTH_TOKEN_PATH"`
	RegionToInstanceConfig   map[string]InstanceConfig `envconfig:"-"`
}

// InstanceConfig is a service instance config
type InstanceConfig struct {
	ClientID       string
	ClientSecret   string
	OAuthURL       string
	OAuthTokenPath string
	SMURL          string
}

func New() (Config, error) {
	cfg := Config{}
	return cfg, envconfig.InitWithPrefix(&cfg, envPrefix)
}

// MapInstanceConfigs parses the InstanceConfigs json string to map with key: region name and value: InstanceConfig for the instance in the region
func (c *Config) MapInstanceConfigs() error {
	secretData, err := pkgconfig.ReadConfigFile(c.ServiceManagerCfg.InstancesSecretPath)
	if err != nil {
		return errors.Wrapf(err, "while getting SM instances secret")
	}

	bindingsMap, err := pkgconfig.ParseConfigToJSONMap(secretData)
	if err != nil {
		return err
	}

	c.ServiceManagerCfg.RegionToInstanceConfig = make(map[string]InstanceConfig)
	for region, config := range bindingsMap {
		i := InstanceConfig{
			ClientID:       gjson.Get(config.String(), c.ServiceManagerCfg.InstanceClientIDPath).String(),
			ClientSecret:   gjson.Get(config.String(), c.ServiceManagerCfg.InstanceClientSecretPath).String(),
			SMURL:          gjson.Get(config.String(), c.ServiceManagerCfg.InstanceURLPath).String(),
			OAuthURL:       gjson.Get(config.String(), c.ServiceManagerCfg.InstanceTokenURLPath).String(),
			OAuthTokenPath: c.ServiceManagerCfg.InstanceOAuthTokenPath,
		}

		if err := i.validate(); err != nil {
			c.ServiceManagerCfg.RegionToInstanceConfig = nil
			return errors.Wrapf(err, "while validating instance for region: %q", region)
		}
		c.ServiceManagerCfg.RegionToInstanceConfig[region] = i
	}
	return nil
}

// validate checks if all required fields are populated.
// In the end, the error message is aggregated by joining all error messages.
func (i *InstanceConfig) validate() error {
	errorMessages := make([]string, 0)

	if i.ClientID == "" {
		errorMessages = append(errorMessages, "Client ID is missing")
	}
	if i.ClientSecret == "" {
		errorMessages = append(errorMessages, "Client secret is missing")
	}
	if i.OAuthURL == "" {
		errorMessages = append(errorMessages, "OAuth token URL is missing")
	}
	if i.OAuthTokenPath == "" {
		errorMessages = append(errorMessages, "OAuth token path is missing")
	}
	errorMsg := strings.Join(errorMessages, ", ")
	if errorMsg != "" {
		return errors.New(errorMsg)
	}

	return nil
}
