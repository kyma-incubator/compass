package config

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// SelfRegConfig is configuration for the runtime self-registration flow
type SelfRegConfig struct {
	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`
	SelfRegisterLabelKey            string `envconfig:"APP_SELF_REGISTER_LABEL_KEY,optional"`
	SelfRegisterLabelValuePrefix    string `envconfig:"APP_SELF_REGISTER_LABEL_VALUE_PREFIX,optional"`
	SelfRegisterResponseKey         string `envconfig:"APP_SELF_REGISTER_RESPONSE_KEY,optional"`
	SelfRegisterPath                string `envconfig:"APP_SELF_REGISTER_PATH,optional"`
	SelfRegisterNameQueryParam      string `envconfig:"APP_SELF_REGISTER_NAME_QUERY_PARAM,optional"`
	SelfRegisterTenantQueryParam    string `envconfig:"APP_SELF_REGISTER_TENANT_QUERY_PARAM,optional"`
	SelfRegisterRequestBodyPattern  string `envconfig:"APP_SELF_REGISTER_REQUEST_BODY_PATTERN,optional"`
	SelfRegisterSecretPath          string `envconfig:"APP_SELF_REGISTER_SECRET_PATH"`

	OAuthMode      oauth.AuthMode `envconfig:"APP_SELF_REGISTER_OAUTH_MODE,default=oauth-mtls"`
	OauthTokenPath string         `envconfig:"APP_SELF_REGISTER_OAUTH_TOKEN_PATH,optional"`

	SkipSSLValidation bool `envconfig:"APP_SELF_REGISTER_SKIP_SSL_VALIDATION,default=false"`

	ClientTimeout time.Duration `envconfig:"default=30s"`

	InstanceClientIDPath     string                    `envconfig:"APP_SELF_REGISTER_INSTANCE_CLIENT_ID_PATH"`
	InstanceClientSecretPath string                    `envconfig:"APP_SELF_REGISTER_INSTANCE_CLIENT_SECRET_PATH"`
	InstanceURLPath          string                    `envconfig:"APP_SELF_REGISTER_INSTANCE_URL_PATH"`
	InstanceTokenURLPath     string                    `envconfig:"APP_SELF_REGISTER_INSTANCE_TOKEN_URL_PATH"`
	InstanceCertPath         string                    `envconfig:"APP_SELF_REGISTER_INSTANCE_X509_CERT_PATH"`
	InstanceKeyPath          string                    `envconfig:"APP_SELF_REGISTER_INSTANCE_X509_KEY_PATH"`
	RegionToInstanceConfig   map[string]InstanceConfig `envconfig:"-"`
}

// MapInstanceConfigs parses the InstanceConfigs json string to map with key: region name and value: InstanceConfig for the instance in the region
func (c *SelfRegConfig) MapInstanceConfigs() error {
	secretData, err := ReadConfigFile(c.SelfRegisterSecretPath)
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

		if err := i.validate(c.OAuthMode); err != nil {
			c.RegionToInstanceConfig = nil
			return errors.Wrapf(err, "while validating instance for region: %q", region)
		}

		c.RegionToInstanceConfig[region] = i
	}

	return nil
}
