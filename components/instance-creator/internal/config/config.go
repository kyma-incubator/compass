package config

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	pkgconfig "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// TenantInfo contains necessary configuration for determining the CMP tenant info
type TenantInfo struct {
	Endpoint           string        `envconfig:"APP_TENANT_INFO_ENDPOINT,default=localhost:8080/v1/info"`
	RequestTimeout     time.Duration `envconfig:"APP_TENANT_INFO_REQUEST_TIMEOUT,default=30s"`
	InsecureSkipVerify bool          `envconfig:"APP_TENANT_INFO_INSECURE_SKIP_VERIFY,default=false"`
}

// Config contains necessary configurations for the instance-creator to operate
type Config struct {
	APIRootPath            string        `envconfig:"APP_API_ROOT_PATH,default=/instance-creator"`
	Address                string        `envconfig:"APP_ADDRESS,default=localhost:8080"`
	SkipSSLValidation      bool          `envconfig:"APP_HTTP_CLIENT_SKIP_SSL_VALIDATION,default=false"`
	JWKSEndpoint           string        `envconfig:"APP_JWKS_ENDPOINT,default=file://hack/default-jwks.json"`
	ServerTimeout          time.Duration `envconfig:"APP_SERVER_TIMEOUT,default=110s"`
	ClientTimeout          time.Duration `envconfig:"APP_SM_CLIENT_TIMEOUT,default=105s"` // todo::: double check value
	AuthorizationHeaderKey string        `envconfig:"APP_AUTHORIZATION_HEADER_KEY,default=Authorization"`
	AllowJWTSigningNone    bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`

	SMInstancesSecretPath        string                    `envconfig:"APP_SM_INSTANCES_SECRET_PATH"`
	InstanceClientIDPath         string                    `envconfig:"APP_SM_INSTANCE_CLIENT_ID_PATH"`
	InstanceSMURLPath            string                    `envconfig:"APP_SM_INSTANCE_SM_URL_PATH"`
	InstanceTokenURLPath         string                    `envconfig:"APP_SM_INSTANCE_TOKEN_URL_PATH"`
	InstanceAppNamePath          string                    `envconfig:"APP_SM_INSTANCE_APP_NAME_PATH"`
	InstanceCertificatePath      string                    `envconfig:"APP_SM_INSTANCE_CERTIFICATE_PATH"`
	InstanceCertificateKeyPath   string                    `envconfig:"APP_SM_INSTANCE_CERTIFICATE_KEY_PATH"`
	ExternalClientCertSecretName string                    `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	OAuthTokenPath               string                    `envconfig:"APP_SM_INSTANCE_OAUTH_TOKEN_PATH"`
	RegionToInstanceConfig       map[string]InstanceConfig `envconfig:"-"`

	Log        log.Config
	TenantInfo TenantInfo
}

// InstanceConfig is a service instance config
type InstanceConfig struct {
	ClientID       string
	SMURL          string
	TokenURL       string
	AppName        string
	Certificate    string
	CertificateKey string
}

func (c *Config) PrepareConfiguration() error {
	if err := c.MapInstanceConfigs(); err != nil {
		return errors.Wrap(err, "while building region instances credentials")
	}

	return nil
}

// MapInstanceConfigs parses the InstanceConfigs json string to map with key: region name and value: InstanceConfig for the instance in the region
func (c *Config) MapInstanceConfigs() error {
	secretData, err := pkgconfig.ReadConfigFile(c.SMInstancesSecretPath)
	if err != nil {
		return errors.Wrapf(err, "while getting SM instances secret")
	}

	bindingsMap, err := pkgconfig.ParseConfigToJSONMap(secretData)
	if err != nil {
		return err
	}

	c.RegionToInstanceConfig = make(map[string]InstanceConfig)
	for region, config := range bindingsMap {
		i := InstanceConfig{
			ClientID:       gjson.Get(config.String(), c.InstanceClientIDPath).String(),
			SMURL:          gjson.Get(config.String(), c.InstanceSMURLPath).String(),
			TokenURL:       gjson.Get(config.String(), c.InstanceTokenURLPath).String(),
			AppName:        gjson.Get(config.String(), c.InstanceAppNamePath).String(),
			Certificate:    gjson.Get(config.String(), c.InstanceCertificatePath).String(),
			CertificateKey: gjson.Get(config.String(), c.InstanceCertificateKeyPath).String(),
		}

		if err := i.validate(); err != nil {
			c.RegionToInstanceConfig = nil
			return errors.Wrapf(err, "while validating instance for region: %q", region)
		}
		c.RegionToInstanceConfig[region] = i
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
	if i.SMURL == "" {
		errorMessages = append(errorMessages, "SM TokenURL is missing")
	}
	if i.TokenURL == "" {
		errorMessages = append(errorMessages, "TokenURL is missing")
	}
	if i.AppName == "" {
		errorMessages = append(errorMessages, "App Name is missing")
	}
	if i.Certificate == "" {
		errorMessages = append(errorMessages, "Certificate is missing")
	}
	if i.CertificateKey == "" {
		errorMessages = append(errorMessages, "Certificate Key is missing")
	}

	errorMsg := strings.Join(errorMessages, ", ")
	if errorMsg != "" {
		return errors.New(errorMsg)
	}

	return nil
}
