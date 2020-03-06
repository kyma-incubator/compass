package azure

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/gobuffalo/envy"
)

func GetConfig(clientID, clientSecret, tenantID, subscriptionID string) *Config {
	config := NewDefaultConfig()

	azureEnv, _ := azure.EnvironmentFromName("AzurePublicCloud") // shouldn't fail
	config.authorizationServerURL = azureEnv.ActiveDirectoryEndpoint
	config.clientID = clientID
	config.clientSecret = clientSecret
	config.tenantID = tenantID
	config.subscriptionID = subscriptionID
	return config
}

// TODO(nachtmaar): delete me since GetConfig should be used in KEB later
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

	clientID, err := envy.MustGet("AZURE_CLIENT_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	clientSecret, err := envy.MustGet("AZURE_CLIENT_SECRET")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	tenantID, err := envy.MustGet("AZURE_TENANT_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	subscriptionID, err := envy.MustGet("AZURE_SUBSCRIPTION_ID")
	if err != nil {
		return nil, fmt.Errorf("expected env vars not provided: %s", err)
	}

	return GetConfig(clientID, clientSecret, tenantID, subscriptionID), nil
}
