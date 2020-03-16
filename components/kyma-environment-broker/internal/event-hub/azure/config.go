package azure

import (
	"log"

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

// TODO(nachtmaar): do not panic
func (c *Config) Environment() *azure.Environment {
	if c.environment != nil {
		return c.environment
	}

	env, err := azure.EnvironmentFromName(c.cloudName)
	if err != nil {
		log.Panicf("invalid cloud name [%s]", c.cloudName)
	}
	c.environment = &env

	return c.environment
}
