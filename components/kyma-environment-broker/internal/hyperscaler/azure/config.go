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

func azureRegions() map[string]bool {
	return map[string]bool{
		"australiacentral":   true,
		"australiacentral2":  true,
		"australiaeast":      true,
		"australiasoutheast": true,
		"brazilsouth":        true,
		"canadacentral":      true,
		"canadaeast":         true,
		"centralindia":       true,
		"centralus":          true,
		"eastasia":           true,
		"eastus":             true,
		"eastus2":            true,
		"francecentral":      true,
		"francesouth":        true,
		"germanynorth":       true,
		"germanywestcentral": true,
		"japaneast":          true,
		"japanwest":          true,
		"koreacentral":       true,
		"koreasouth":         true,
		"northcentralus":     true,
		"northeurope":        true,
		"norwayeast":         true,
		"norwaywest":         true,
		"southafricanorth":   true,
		"southafricawest":    true,
		"southcentralus":     true,
		"southeastasia":      true,
		"southindia":         true,
		"switzerlandnorth":   true,
		"switzerlandwest":    true,
		"uaecentral":         true,
		"uaenorth":           true,
		"uksouth":            true,
		"ukwest":             true,
		"westcentralus":      true,
		"westeurope":         true,
		"westindia":          true,
		"westus":             true,
		"westus2":            true,
	}
}

func gcpToAzureRegionMapping() map[string]string {
	return map[string]string{
		"asia-east1":              "eastasia",
		"asia-east2":              "eastasia",
		"asia-northeast1":         "eastasia",
		"asia-northeast2":         "eastasia",
		"asia-northeast3":         "eastasia",
		"asia-south1":             "southeastasia",
		"asia-southeast1":         "southeastasia",
		"australia-southeast1":    "australiasoutheast",
		"europe-north1":           "northeurope",
		"europe-west1":            "westeurope",
		"europe-west2":            "westeurope",
		"europe-west3":            "westeurope",
		"europe-west4":            "westeurope",
		"europe-west6":            "westeurope",
		"northamerica-northeast1": "canadacentral",
		"southamerica-east1":      "brazilsouth",
		"us-central1":             "centralus",
		"us-east1":                "eastus",
		"us-east4":                "eastus",
		"us-west1":                "westus",
		"us-west2":                "westus",
		"us-west3":                "westus",
	}
}
