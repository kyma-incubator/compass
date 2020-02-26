package azure

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/gobuffalo/envy"
)

func GetConfigFromEnvironment(files ...string) (*Config, error) {
	config := NewDefaultConfig()

	err := envy.Load(files...)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %s", err)
	}

	azureEnv, _ := azure.EnvironmentFromName("AzurePublicCloud") // shouldn't fail
	config.authorizationServerURL = azureEnv.ActiveDirectoryEndpoint

	config.keepResources, err = strconv.ParseBool(envy.Get("AZURE_SAMPLES_KEEP_RESOURCES", "0"))
	if err != nil {
		log.Printf("invalid value specified for AZURE_SAMPLES_KEEP_RESOURCES, discarding")
		config.keepResources = false
	}

	config.clientID, err = envy.MustGet("AZURE_CLIENT_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	config.clientSecret, err = envy.MustGet("AZURE_CLIENT_SECRET")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	config.tenantID, err = envy.MustGet("AZURE_TENANT_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	config.subscriptionID, err = envy.MustGet("AZURE_SUBSCRIPTION_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	return config, nil
}
