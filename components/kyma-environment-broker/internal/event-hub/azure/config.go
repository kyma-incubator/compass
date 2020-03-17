package azure

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"
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
		return nil, fmt.Errorf("invalid cloud name [%s], error: %v", c.cloudName, err)
	}
	c.environment = &env

	return c.environment, nil
}
