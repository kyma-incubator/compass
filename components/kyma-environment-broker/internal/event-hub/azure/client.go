package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
)

type HyperscalerProvider interface {
	GetClient(config *Config) (EventhubsInterface, error)
}

var _ HyperscalerProvider = (*azureClient)(nil)

type azureClient struct{}

func NewAzureClient() HyperscalerProvider {
	return &azureClient{}
}

// GetClient gets a client for interacting with Azure EventHubs
func (ac *azureClient) GetClient(config *Config) (EventhubsInterface, error) {
	nsClient := eventhub.NewNamespacesClient(config.subscriptionID)

	authorizer, err := GetResourceManagementAuthorizer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authorizer with error: %v", err)
	}
	nsClient.Authorizer = authorizer

	if err = nsClient.AddToUserAgent(config.userAgent); err != nil {
		return nil, fmt.Errorf("failed to add use agent [%s] with error: %v", config.userAgent, err)
	}

	return NewNamespaceClient(nsClient), nil
}
