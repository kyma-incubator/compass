package azure

import (
	"fmt"

	autorestazure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/kyma-project/control-plane/tests/e2e/provisioning/internal/hyperscaler"
)

// GetConfig returns Azure config
func GetConfig(clientID, clientSecret, tenantID, subscriptionID, location string) (*Config, error) {
	config := NewDefaultConfig()
	azureEnv, err := autorestazure.EnvironmentFromName("AzurePublicCloud")
	if err != nil {
		return nil, err
	}

	if azureEnv.ActiveDirectoryEndpoint == "" {
		return nil, fmt.Errorf("failed to initialize Config as ActiveDirectoryEndpoint is empty")
	}
	config.authorizationServerURL = azureEnv.ActiveDirectoryEndpoint

	if clientID == "" {
		return nil, fmt.Errorf("failed to initialize Config as clientID is empty")
	}
	config.clientID = clientID

	if clientSecret == "" {
		return nil, fmt.Errorf("failed to initialize Config as clientSecret is empty")
	}
	config.clientSecret = clientSecret

	if tenantID == "" {
		return nil, fmt.Errorf("failed to initialize Config as tenantID is empty")
	}
	config.tenantID = tenantID

	if subscriptionID == "" {
		return nil, fmt.Errorf("failed to initialize Config as subscriptionID is empty")
	}
	config.subscriptionID = subscriptionID

	if location == "" {
		return nil, fmt.Errorf("failed to initialize Config as location is empty")
	}
	config.location = location
	return config, nil
}

// GetConfigFromHAPCredentialsAndProvisioningParams returns Azure config from HAPCredentials
func GetConfigFromHAPCredentialsAndProvisioningParams(credentials hyperscaler.Credentials, region string) (*Config, error) {
	subscriptionID := string(credentials.CredentialData["subscriptionID"])
	clientID := string(credentials.CredentialData["clientID"])
	clientSecret := string(credentials.CredentialData["clientSecret"])
	tenantID := string(credentials.CredentialData["tenantID"])
	return GetConfig(clientID, clientSecret, tenantID, subscriptionID, region)
}
