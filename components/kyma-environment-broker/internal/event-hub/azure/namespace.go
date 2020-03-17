package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type EventhubsInterface interface {
	GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	CreateResourceGroup(ctx context.Context, config *Config, name string) (resources.Group, error)
	CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string) (*eventhub.EHNamespace, error)
}

var _ EventhubsInterface = (*NamespaceClient)(nil)

type NamespaceClient struct {
	// the actual azure client
	eventhubNamespaceClient eventhub.NamespacesClient
}

func NewNamespaceClient(client eventhub.NamespacesClient) *NamespaceClient {
	return &NamespaceClient{
		eventhubNamespaceClient: client,
	}
}
func (nc *NamespaceClient) GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	return nc.eventhubNamespaceClient.ListKeys(ctx, resourceGroupName, namespaceName, authorizationRuleName)
}

func (nc *NamespaceClient) CreateResourceGroup(ctx context.Context, config *Config, name string) (resources.Group, error) {
	client, err := getGroupsClient(config)
	if err != nil {
		return resources.Group{}, err
	}
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := config.GetLocation()
	return client.CreateOrUpdate(ctx, name, resources.Group{Location: &locationCopy})
}

func (nc *NamespaceClient) CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string) (*eventhub.EHNamespace, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := azureCfg.GetLocation()
	parameters := eventhub.EHNamespace{Location: &locationCopy}
	ehNamespace, err := nc.createOrUpdate(ctx, groupName, namespace, parameters)
	return &ehNamespace, err
}

func (nc *NamespaceClient) createOrUpdate(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
	future, err := nc.eventhubNamespaceClient.CreateOrUpdate(ctx, resourceGroupName, namespaceName, parameters)
	if err != nil {
		return eventhub.EHNamespace{}, err
	}

	err = future.WaitForCompletionRef(ctx, nc.eventhubNamespaceClient.Client)
	if err != nil {
		return eventhub.EHNamespace{}, err
	}

	result, err = future.Result(nc.eventhubNamespaceClient)
	return result, err
}
