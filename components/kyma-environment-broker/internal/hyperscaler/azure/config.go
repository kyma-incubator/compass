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

// see internal/broker/plans.go for list of available azure regions
// important" ! filter non evnehub regions out
// filtered entries: "uaenorth"
// list used: https://docs.microsoft.com/en-us/rest/api/eventhub/regions/listbysku

func azureRegions() map[string]struct{} {
	return map[string]struct{}{
		"EastUS2EUAP":   {},
		"FranceCentral": {},
		"centralus":     {},
		"eastus":        {},
		"eastus2":       {},
		"japaneast":     {},
		"northeurope":   {},
		"southeastasia": {},
		"uksouth":       {},
		"westeurope":    {},
		"westus2":       {},
	}
}

// see internal/broker/plans.go for list of available gcp regions
func gcpToAzureRegionMapping() map[string]string {
	return map[string]string{
		"asia-east1":              "southeastasia",
		"asia-east2":              "southeastasia",
		"asia-northeast1":         "southeastasia",
		"asia-northeast2":         "southeastasia",
		"asia-south1":             "southeastasia",
		"asia-southeast1":         "southeastasia",
		"australia-southeast1":    "southeastasia",
		"europe-north1":           "northeurope",
		"europe-west1":            "westeurope",
		"europe-west2":            "westeurope",
		"europe-west3":            "westeurope",
		"europe-west4":            "westeurope",
		"europe-west6":            "westeurope",
		"northamerica-northeast1": "eastus",
		"southamerica-east1":      "westus2",
		"us-central1":             "centralus",
		"us-east1":                "eastus",
		"us-east4":                "eastus",
		"us-west1":                "westus",
		"us-west2":                "westus",
	}
}
