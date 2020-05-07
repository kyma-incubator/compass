package azure

import (
	autorestazure "github.com/Azure/go-autorest/autorest/azure"

	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/hyperscaler"
)

// GetConfig returns Azure config
func GetConfig(clientID, clientSecret, tenantID, subscriptionID, location string) (*Config, error) {
	config := NewDefaultConfig()
	azureEnv, err := autorestazure.EnvironmentFromName("AzurePublicCloud") // shouldn't fail
	if err != nil {
		return nil, err
	}

	config.authorizationServerURL = azureEnv.ActiveDirectoryEndpoint
	config.clientID = clientID
	config.clientSecret = clientSecret
	config.tenantID = tenantID
	config.subscriptionID = subscriptionID
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
