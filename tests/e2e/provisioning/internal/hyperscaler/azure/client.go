package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type AzureInterface interface {
	GetResourceGroup(ctx context.Context, name string) (resources.Group, error)
	GetEHNamespace(ctx context.Context, resourceGroupName,namespaceName string) (eventhub.EHNamespace, error)
}

var _ AzureInterface = (*AzureClient)(nil)

type AzureClient struct {
	eventhubNamespaceClient eventhub.NamespacesClient
	resourcegroupClient     resources.GroupsClient
}

func NewAzureClient(namespaceClient eventhub.NamespacesClient, resourcegroupClient resources.GroupsClient) *AzureClient {
	return &AzureClient{
		eventhubNamespaceClient: namespaceClient,
		resourcegroupClient:     resourcegroupClient,
	}
}

func (nc *AzureClient) GetResourceGroup(ctx context.Context, name string) (resources.Group, error) {
	return nc.resourcegroupClient.Get(ctx, name)
}

func (nc *AzureClient) GetEHNamespace(ctx context.Context, resourceGroupName,namespaceName string) (eventhub.EHNamespace, error) {
	return nc.eventhubNamespaceClient.Get(ctx, resourceGroupName, namespaceName)
}
