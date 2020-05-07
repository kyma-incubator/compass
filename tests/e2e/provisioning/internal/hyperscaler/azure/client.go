package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type AzureInterface interface {
	ListResourceGroup(ctx context.Context, filter string, top *int32) (resources.GroupListResultPage, error)
	ListEHNamespaceByResourceGroup(ctx context.Context, resourceGroupName string) (eventhub.EHNamespaceListResultPage, error)
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

func (nc *AzureClient) ListResourceGroup(ctx context.Context, filter string, top *int32) (resources.GroupListResultPage, error) {
	return nc.resourcegroupClient.List(ctx, filter, top)
}

func (nc *AzureClient) ListEHNamespaceByResourceGroup(ctx context.Context, resourceGroupName string) (eventhub.EHNamespaceListResultPage, error) {
	return nc.eventhubNamespaceClient.ListByResourceGroup(ctx, resourceGroupName)
}
