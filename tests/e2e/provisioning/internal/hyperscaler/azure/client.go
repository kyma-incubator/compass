package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type Interface interface {
	ListResourceGroup(ctx context.Context, filter string, top *int32) (resources.GroupListResultPage, error)
	ListEHNamespaceByResourceGroup(ctx context.Context, resourceGroupName string) (eventhub.EHNamespaceListResultPage, error)
}

var _ Interface = (*Client)(nil)

type Client struct {
	eventHubNamespaceClient eventhub.NamespacesClient
	resourceGroupClient     resources.GroupsClient
}

func NewAzureClient(namespaceClient eventhub.NamespacesClient, resourceGroupClient resources.GroupsClient) *Client {
	return &Client{
		eventHubNamespaceClient: namespaceClient,
		resourceGroupClient:     resourceGroupClient,
	}
}

func (nc *Client) ListResourceGroup(ctx context.Context, filter string, top *int32) (resources.GroupListResultPage, error) {
	return nc.resourceGroupClient.List(ctx, filter, top)
}

func (nc *Client) ListEHNamespaceByResourceGroup(ctx context.Context, resourceGroupName string) (eventhub.EHNamespaceListResultPage, error) {
	return nc.eventHubNamespaceClient.ListByResourceGroup(ctx, resourceGroupName)
}
