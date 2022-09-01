package config

import (
	"encoding/base64"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// DestinationsConfig destination service configuration
type DestinationsConfig struct {
	InstanceClientIDPath     string                    `envconfig:"APP_DESTINATION_INSTANCE_CLIENT_ID_PATH,default=clientid"`
	InstanceClientSecretPath string                    `envconfig:"APP_DESTINATION_INSTANCE_CLIENT_SECRET_PATH,default=clientsecret"`
	InstanceURLPath          string                    `envconfig:"APP_DESTINATION_INSTANCE_URL_PATH,default=uri"`
	InstanceTokenURLPath     string                    `envconfig:"APP_DESTINATION_INSTANCE_TOKEN_URL_PATH,default=certurl"`
	InstanceCertPath         string                    `envconfig:"APP_DESTINATION_INSTANCE_X509_CERT_PATH,default=certificate"`
	InstanceKeyPath          string                    `envconfig:"APP_DESTINATION_INSTANCE_X509_KEY_PATH,default=key"`
	DestinationSecretPath    string                    `envconfig:"APP_DESTINATION_SECRET_PATH"`
	RegionToInstanceConfig   map[string]InstanceConfig `envconfig:"-"`
	OAuthMode                oauth.AuthMode            `envconfig:"APP_DESTINATION_OAUTH_MODE,default=oauth-mtls"`
}

// MapInstanceConfigs creates region to destination configuration map
func (c *DestinationsConfig) MapInstanceConfigs() error {
	secretData, err := ReadConfigFile(c.DestinationSecretPath)
	if err != nil {
		return errors.Wrapf(err, "while getting destinations secret")
	}

	bindingsMap, err := ParseConfigToJSONMap(secretData)
	if err != nil {
		return err
	}

	c.RegionToInstanceConfig = make(map[string]InstanceConfig)
	for region, config := range bindingsMap {
		i := InstanceConfig{
			ClientID:     gjson.Get(config.String(), c.InstanceClientIDPath).String(),
			ClientSecret: gjson.Get(config.String(), c.InstanceClientSecretPath).String(),
			URL:          gjson.Get(config.String(), c.InstanceURLPath).String(),
			TokenURL:     gjson.Get(config.String(), c.InstanceTokenURLPath).String(),
			Cert:         gjson.Get(config.String(), c.InstanceCertPath).String(),
			Key:          gjson.Get(config.String(), c.InstanceKeyPath).String(),
		}

		if i.Cert != "" {
			decodeCert, err := base64.StdEncoding.DecodeString(i.Cert)
			if err != nil {
				return errors.Wrap(err, "could not base64 decode client certificate")
			}
			i.Cert = string(decodeCert)
		}

		if i.Key != "" {
			decodeKey, err := base64.StdEncoding.DecodeString(i.Key)
			if err != nil {
				return errors.Wrap(err, "could not base64 decode client certificate")
			}
			i.Key = string(decodeKey)
		}

		if err := i.validate(c.OAuthMode); err != nil {
			c.RegionToInstanceConfig = nil
			return errors.Wrapf(err, "while validating instance for region: %q", region)
		}
		c.RegionToInstanceConfig[region] = i
	}

	return nil
}
