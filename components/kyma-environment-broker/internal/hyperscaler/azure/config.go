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

// see internal/broker/plans.go for list of available gcp regions
func gcpToAzureRegionMapping() map[string]string {
	return map[string]string{
		"asia-east1":              "westeurope",
		"asia-east2":              "westeurope",
		"asia-northeast1":         "westeurope",
		"asia-northeast2":         "westeurope",
		"asia-south1":             "westeurope",
		"asia-southeast1":         "westeurope",
		"australia-southeast1":    "westeurope",
		"europe-north1":           "westeurope",
		"europe-west1":            "westeurope",
		"europe-west2":            "westeurope",
		"europe-west3":            "westeurope",
		"europe-west4":            "westeurope",
		"europe-west6":            "westeurope",
		"northamerica-northeast1": "westus2",
		"southamerica-east1":      "westus2",
		"us-central1":             "westus2",
		"us-east1":                "westus2",
		"us-east4":                "westus2",
		"us-west1":                "westus2",
		"us-west2":                "westus2",
	}
}
