package azure

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provider"
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

func GetConfigFromHAPCredentialsAndProvisioningParams(credentials hyperscaler.Credentials, parameters internal.ProvisioningParameters) (*Config, error) {
	region, err := mapRegion(credentials, parameters)
	if err != nil {
		return nil, err
	}

	subscriptionID := string(credentials.CredentialData["subscriptionID"])
	clientID := string(credentials.CredentialData["clientID"])
	clientSecret := string(credentials.CredentialData["clientSecret"])
	tenantID := string(credentials.CredentialData["tenantID"])
	return GetConfig(clientID, clientSecret, tenantID, subscriptionID, region)
}

func mapRegion(credentials hyperscaler.Credentials, parameters internal.ProvisioningParameters) (string, error) {
	if credentials.HyperscalerType != hyperscaler.Azure {
		return "", fmt.Errorf("cannot use credential for hyperscaler of type %v on hyperscaler of type %v", credentials.HyperscalerType, hyperscaler.Azure)
	}
	if parameters.Parameters.Region == nil || *(parameters.Parameters.Region) == "" {
		return provider.DefaultAzureRegion, nil
	}
	region := *(parameters.Parameters.Region)
	switch parameters.PlanID {
	case broker.AzurePlanID:
		if !isInList(broker.AzureRegions(), region) {
			return "", fmt.Errorf("supplied region \"%v\" is not a valid region for Azure", region)
		}

	case broker.GCPPlanID:
		if azureRegion, mappingExists := gcpToAzureRegionMapping()[region]; mappingExists {
			region = azureRegion
			break
		}
		return "", fmt.Errorf("supplied gcp region \"%v\" cannot be mapped to Azure", region)
	default:
		return "", fmt.Errorf("cannot map from PlanID %v to azure regions", parameters.PlanID)
	}
	return region, nil
}

func isInList(list []string, item string) bool {
	for _, val := range list {
		if val == item {
			return true
		}
	}
	return false
}
