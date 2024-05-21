package config

import (
	directorcfg "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// SystemFieldDiscoveryEngineConfig is configuration for the system field discovery engine
type SystemFieldDiscoveryEngineConfig struct {
	SaasRegSecretPath string `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_SAAS_APP_SECRET_PATH"`
	OauthTokenPath    string `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_OAUTH_TOKEN_PATH,optional"`

	SaasRegClientIDPath     string `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_INSTANCE_CLIENT_ID_PATH"`
	SaasRegClientSecretPath string `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_CLIENT_SECRET_PATH"`
	SaasRegTokenURLPath     string `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_URL_PATH"`
	SaasRegURLPath          string `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_SAAS_REGISTRY_URL_PATH"`

	RegionToSaasRegConfig map[string]SaasRegConfig `envconfig:"-"`
}

// PrepareConfiguration take cares to build the system field discovery engine configuration
func (c SystemFieldDiscoveryEngineConfig) PrepareConfiguration() (*SystemFieldDiscoveryEngineConfig, error) {
	sfdCfg, err := c.MapSaasRegConfigs()
	if err != nil {
		return nil, errors.Wrap(err, "while building region instances credentials")
	}

	return sfdCfg, nil
}

// MapSaasRegConfigs parses the SaasRegConfigs json string to map with key: region name and value: SaasRegConfig for the instance in the region
func (c SystemFieldDiscoveryEngineConfig) MapSaasRegConfigs() (*SystemFieldDiscoveryEngineConfig, error) {
	secretData, err := directorcfg.ReadConfigFile(c.SaasRegSecretPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting destinations secret")
	}

	bindingsMap, err := directorcfg.ParseConfigToJSONMap(secretData)
	if err != nil {
		return nil, err
	}
	c.RegionToSaasRegConfig = make(map[string]SaasRegConfig)
	for region, config := range bindingsMap {
		s := SaasRegConfig{
			ClientID:        gjson.Get(config.String(), c.SaasRegClientIDPath).String(),
			ClientSecret:    gjson.Get(config.String(), c.SaasRegClientSecretPath).String(),
			TokenURL:        gjson.Get(config.String(), c.SaasRegTokenURLPath).String(),
			SaasRegistryURL: gjson.Get(config.String(), c.SaasRegURLPath).String(),
		}

		if err := s.validate(); err != nil {
			c.RegionToSaasRegConfig = nil
			return nil, errors.Wrapf(err, "while validating saas reg config for region: %q", region)
		}

		c.RegionToSaasRegConfig[region] = s
	}

	return &c, nil
}
