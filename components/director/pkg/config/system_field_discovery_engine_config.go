package config

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type SystemFieldDiscoveryEngineConfig struct {
	SaasRegSecretPath string `envconfig:"APP_SELF_REGISTER_SECRET_PATH"`
	OauthTokenPath    string `envconfig:"APP_SELF_REGISTER_OAUTH_TOKEN_PATH,optional"`

	SaasRegClientIDPath     string `envconfig:"APP_SELF_REGISTER_INSTANCE_CLIENT_ID_PATH"`
	SaasRegClientSecretPath string `envconfig:"APP_SELF_REGISTER_INSTANCE_CLIENT_SECRET_PATH"`
	SaasRegTokenURLPath     string `envconfig:"APP_SELF_REGISTER_INSTANCE_URL_PATH"`
	SaasRegURLPath          string `envconfig:"APP_SELF_REGISTER_SAAS_REGISTRY_URL_PATH"`

	RegionToSaasRegConfig map[string]SaasRegConfig `envconfig:"-"`
}

// PrepareConfiguration take cares to build the system field discovery engine configuration
func (c *SystemFieldDiscoveryEngineConfig) PrepareConfiguration() error {
	if err := c.MapSaasRegConfigs(); err != nil {
		return errors.Wrap(err, "while building region instances credentials")
	}

	return nil
}

// MapSaasRegConfigs parses the SaasRegConfigs json string to map with key: region name and value: SaasRegConfig for the instance in the region
func (c *SystemFieldDiscoveryEngineConfig) MapSaasRegConfigs() error {
	secretData, err := ReadConfigFile(c.SaasRegSecretPath)
	if err != nil {
		return errors.Wrapf(err, "while getting destinations secret")
	}

	bindingsMap, err := ParseConfigToJSONMap(secretData)
	if err != nil {
		return err
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
			return errors.Wrapf(err, "while validating saas reg config for region: %q", region)
		}

		c.RegionToSaasRegConfig[region] = s
	}

	return nil
}
