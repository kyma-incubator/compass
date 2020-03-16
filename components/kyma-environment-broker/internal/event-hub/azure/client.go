package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	log "github.com/sirupsen/logrus"
)

type HyperscalerProvider interface {
	GetClientOrDie(config *Config) NamespaceClientInterface
}

var _ HyperscalerProvider = (*azureClient)(nil)

type azureClient struct{}

func NewAzureClient() HyperscalerProvider {
	return &azureClient{}
}

func (ac *azureClient) GetClientOrDie(config *Config) NamespaceClientInterface {
	// TODO(nachtmaar): don't die here, instead return error
	nsClient := eventhub.NewNamespacesClient(config.subscriptionID)

	authorizer, err := GetResourceManagementAuthorizer(config)
	if err != nil {
		log.Fatalf("Failed to initialize authorizer with error: %v", err)
	}
	nsClient.Authorizer = authorizer

	if err = nsClient.AddToUserAgent(config.userAgent); err != nil {
		log.Fatalf("Failed to add use agent [%s] with error: %v", config.userAgent, err)
	}

	return NewNamespaceClient(nsClient)
}
