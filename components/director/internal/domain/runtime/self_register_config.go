package runtime

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
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
	InstanceConfigs          string                    `envconfig:"APP_SELF_REGISTER_AGGREGATED_INSTANCE_CONFIGS"`
	RegionToInstanceConfig   map[string]InstanceConfig `envconfig:"-"`
}

// InstanceConfig is configuration for communication with specific instance in the runtime self-registration flow
type InstanceConfig struct {
	ClientID     string
	ClientSecret string
	URL          string
	TokenURL     string
	Cert         string
	Key          string
}

// MapInstanceConfigs parses the InstanceConfigs json string to map with key: regin_name and value: InstanceConfig for the instance in the region
func (c *SelfRegConfig) MapInstanceConfigs() {
	bindingsResult := gjson.Parse(c.InstanceConfigs)
	bindingsMap := bindingsResult.Map()
	c.RegionToInstanceConfig = make(map[string]InstanceConfig)
	for k, v := range bindingsMap {
		i := InstanceConfig{
			gjson.Get(v.String(), c.InstanceClientIDPath).String(),
			gjson.Get(v.String(), c.InstanceClientSecretPath).String(),
			gjson.Get(v.String(), c.InstanceURLPath).String(),
			gjson.Get(v.String(), c.InstanceTokenURLPath).String(),
			gjson.Get(v.String(), c.InstanceCertPath).String(),
			gjson.Get(v.String(), c.InstanceKeyPath).String(),
		}
		c.RegionToInstanceConfig[k] = i
	}
}
