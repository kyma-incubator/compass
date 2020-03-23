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

func azureRegions() map[string]struct{} {
	return map[string]struct{}{
		"australiacentral":   {},
		"australiacentral2":  {},
		"australiaeast":      {},
		"australiasoutheast": {},
		"brazilsouth":        {},
		"canadacentral":      {},
		"canadaeast":         {},
		"centralindia":       {},
		"centralus":          {},
		"eastasia":           {},
		"eastus":             {},
		"eastus2":            {},
		"francecentral":      {},
		"francesouth":        {},
		"germanynorth":       {},
		"germanywestcentral": {},
		"japaneast":          {},
		"japanwest":          {},
		"koreacentral":       {},
		"koreasouth":         {},
		"northcentralus":     {},
		"northeurope":        {},
		"norwayeast":         {},
		"norwaywest":         {},
		"southafricanorth":   {},
		"southafricawest":    {},
		"southcentralus":     {},
		"southeastasia":      {},
		"southindia":         {},
		"switzerlandnorth":   {},
		"switzerlandwest":    {},
		"uaecentral":         {},
		"uaenorth":           {},
		"uksouth":            {},
		"ukwest":             {},
		"westcentralus":      {},
		"westeurope":         {},
		"westindia":          {},
		"westus":             {},
		"westus2":            {},
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
