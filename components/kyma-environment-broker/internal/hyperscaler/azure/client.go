package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type AzureInterface interface {
	GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	CreateResourceGroup(ctx context.Context, config *Config, name string, tags map[string]*string) (resources.Group, error)
	DeleteResourceGroup(ctx context.Context) error
	CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags map[string]*string) (*eventhub.EHNamespace, error)
}

var _ AzureInterface = (*AzureClient)(nil)

type AzureClient struct {
	// the actual azure client
	eventhubNamespaceClient eventhub.NamespacesClient
	resourcegroupClient     resources.GroupsClient
}

func NewAzureClient(namespaceClient eventhub.NamespacesClient, resourcegroupClient resources.GroupsClient) *AzureClient {
	return &AzureClient{
		eventhubNamespaceClient: namespaceClient,
		resourcegroupClient:     resourcegroupClient,
	}
}
func (nc *AzureClient) GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	return nc.eventhubNamespaceClient.ListKeys(ctx, resourceGroupName, namespaceName, authorizationRuleName)
}

func (nc *AzureClient) CreateResourceGroup(ctx context.Context, config *Config, name string, tags map[string]*string) (resources.Group, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := config.GetLocation()
	return nc.resourcegroupClient.CreateOrUpdate(ctx, name, resources.Group{Location: &locationCopy, Tags: tags})
}

// TODO(nachtmaar): can we have map[string]string here instead and do we nillness checking before ?
func (nc *AzureClient) DeleteResourceGroup(ctx context.Context) error {
	// we need to use a copy of the location, because the following azure call will modify it
	resourceGroupName := nc.getResourceGroupName(ctx)
	return nc.deleteAndWaitResourceGroup(ctx, resourceGroupName)
}

// TODO(nachtmaar): implement me
func (nc *AzureClient) getResourceGroupName(ctx context.Context) string {
	panic("todo")
}

func (nc *AzureClient) CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags map[string]*string) (*eventhub.EHNamespace, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := azureCfg.GetLocation()
	parameters := eventhub.EHNamespace{Location: &locationCopy, Tags: tags}
	ehNamespace, err := nc.createAndWaitNamespace(ctx, groupName, namespace, parameters)
	return &ehNamespace, err
}

// TODO(nachtmaar): can we have a shared method to wait for azure futures ?
func (nc *AzureClient) deleteAndWaitResourceGroup(ctx context.Context, resourceGroupName string) error {
	future, err := nc.resourcegroupClient.Delete(ctx, resourceGroupName)
	if err != nil {
		return err
	}

	err = future.WaitForCompletionRef(ctx, nc.resourcegroupClient.Client)
	return err
}

func (nc *AzureClient) createAndWaitNamespace(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
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
