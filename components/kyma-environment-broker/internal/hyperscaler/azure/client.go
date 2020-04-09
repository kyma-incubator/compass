package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type AzureInterface interface {
	GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	CreateResourceGroup(ctx context.Context, config *Config, name string, tags Tags) (resources.Group, error)
	CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags Tags) (*eventhub.EHNamespace, error)
	DeleteResourceGroup(ctx context.Context, tags Tags) error
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

func (nc *AzureClient) CreateResourceGroup(ctx context.Context, config *Config, name string, tags Tags) (resources.Group, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := config.GetLocation()
	return nc.resourcegroupClient.CreateOrUpdate(ctx, name, resources.Group{Location: &locationCopy, Tags: tags})
}

func (nc *AzureClient) CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags Tags) (*eventhub.EHNamespace, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := azureCfg.GetLocation()
	parameters := eventhub.EHNamespace{Location: &locationCopy, Tags: tags}
	ehNamespace, err := nc.createAndWaitNamespace(ctx, groupName, namespace, parameters)
	return &ehNamespace, err
}

// TODO(nachtmaar): can we have map[string]string here instead and do we nillness checking before ?
func (nc *AzureClient) DeleteResourceGroup(ctx context.Context, tags Tags) error {
	resourceGroup, err := nc.getResourceGroup(ctx, *tags[TagInstanceID])
	if err != nil {
		return err
	}
	// TODO: use logger
	fmt.Printf("found resource group with name: %s\n", *resourceGroup.Name)
	return nc.deleteAndWaitResourceGroup(ctx, *resourceGroup.Name)
}

func (nc *AzureClient) getResourceGroup(ctx context.Context, serviceInstanceID string) (resources.Group, error) {
	filter := fmt.Sprintf("tagName eq 'InstanceID' and tagValue eq '%s'", serviceInstanceID)
	resourceGroupIterator, err := nc.resourcegroupClient.ListComplete(ctx, filter, nil)
	if err != nil {
		return resources.Group{}, err
	}
	// TODO(nachtmaar): return error if there are more than one entry
	for resourceGroupIterator.NotDone() {
		resourceGroup := resourceGroupIterator.Value()
		// there should only be one resource group with the given service instance id
		return resourceGroup, nil
	}
	return resources.Group{}, fmt.Errorf("no resource group found for service instance id: %s", serviceInstanceID)
}

// TODO(nachtmaar): can we have a shared method to wait for azure futures ?
func (nc *AzureClient) deleteAndWaitResourceGroup(ctx context.Context, resourceGroupName string) error {

	fmt.Printf("deleting resource group: %s", resourceGroupName)
	future, err := nc.resourcegroupClient.Delete(ctx, resourceGroupName)
		if err != nil {
		return err
	}

	// TODO: instead of blocking just re-enqueue the task
	fmt.Printf("waiting for deletion of resource group: %s", resourceGroupName)
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
