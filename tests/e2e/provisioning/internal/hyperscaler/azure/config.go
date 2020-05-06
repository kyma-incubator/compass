package azure

import (
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/pkg/errors"
)

type Config struct {
	clientID               string
	clientSecret           string
	tenantID               string
	subscriptionID         string
	location               string
	authorizationServerURL string
	userAgent              string
	cloudName              string
	keepResources          bool
	environment            *azure.Environment
}

func NewDefaultConfig() *Config {
	return &Config{
		userAgent:     "kyma-environment-broker",
		cloudName:     "AzurePublicCloud",
		keepResources: false,
	}
}

func (c *Config) GetLocation() string {
	return c.location
}

func (c *Config) Environment() (*azure.Environment, error) {
	if c.environment != nil {
		return c.environment, nil
	}

	env, err := azure.EnvironmentFromName(c.cloudName)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating azure environment - invalid cloud name [%s]", c.cloudName)
	}
	c.environment = &env

	return c.environment, nil
}
