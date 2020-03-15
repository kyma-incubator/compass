package azure

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
)

func GetConfig(clientID, clientSecret, tenantID, subscriptionID, location string) (*Config, error) {
	config := NewDefaultConfig()

	azureEnv, err := azure.EnvironmentFromName("AzurePublicCloud") // shouldn't fail
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

func GetConfigfromHAPCredentialsAndProvisioningParams(credentials hyperscaler.Credentials, parameters internal.ProvisioningParameters) (*Config, error) {
	// TODO(nachtmaar): validate in early http request
	if parameters.Parameters.Region == nil {
		return nil, fmt.Errorf("region is a required parameter")
	}
	// region := *parameters.Parameters.Region
	region := "westeurope"
	// TODO(nachtmaar): set location https://github.com/kyma-incubator/compass/issues/968
	subscriptionID := string(credentials.CredentialData["subscriptionID"])
	clientID := string(credentials.CredentialData["clientID"])
	clientSecret := string(credentials.CredentialData["clientSecret"])
	tenantID := string(credentials.CredentialData["tenantID"])
	return GetConfig(clientID, clientSecret, tenantID, subscriptionID, region)
}
