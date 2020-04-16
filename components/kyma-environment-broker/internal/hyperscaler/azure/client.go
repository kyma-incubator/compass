package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	AzureFutureOperationSucceeded string = "Succeeded"
	AzureFutureOperationDeleting  string = "Deleting"
)

type AzureInterface interface {
	GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	CreateResourceGroup(ctx context.Context, config *Config, name string, tags Tags) (resources.Group, error)
	CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags Tags) (*eventhub.EHNamespace, error)
	GetResourceGroup(ctx context.Context, tags Tags) (resources.Group, error)
	DeleteResourceGroup(ctx context.Context, tags Tags) (resources.GroupsDeleteFuture, error)
}

var _ AzureInterface = (*AzureClient)(nil)

type AzureClient struct {
	// the actual azure client
	eventhubNamespaceClient eventhub.NamespacesClient
	resourcegroupClient     resources.GroupsClient
	logger                  logrus.FieldLogger
}

type ResourceGroupDoesNotExist struct {
	S string
}

func NewResourceGroupDoesNotExist(errorMessage string) ResourceGroupDoesNotExist {
	return ResourceGroupDoesNotExist{errorMessage}
}

func (e ResourceGroupDoesNotExist) Error() string {
	return e.S
}

func NewAzureClient(namespaceClient eventhub.NamespacesClient, resourcegroupClient resources.GroupsClient, logger logrus.FieldLogger) *AzureClient {
	return &AzureClient{
		eventhubNamespaceClient: namespaceClient,
		resourcegroupClient:     resourcegroupClient,
		logger:                  logger,
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

func (nc *AzureClient) DeleteResourceGroup(ctx context.Context, tags Tags) (resources.GroupsDeleteFuture, error) {
	// get name of resource group
	resourceGroup, err := nc.GetResourceGroup(ctx, tags)
	if err != nil {
		return resources.GroupsDeleteFuture{}, err
	}

	// trigger async deletion of the resource group
	if resourceGroup.Name == nil {
		return resources.GroupsDeleteFuture{}, fmt.Errorf("resource group name is nil")
	}
	nc.logger.Infof("deleting resource group: %s", *resourceGroup.Name)
	future, err := nc.resourcegroupClient.Delete(ctx, *resourceGroup.Name)
	if err != nil {
		return resources.GroupsDeleteFuture{}, err
	}
	return future, err
}

func (nc *AzureClient) GetResourceGroup(ctx context.Context, tags Tags) (resources.Group, error) {
	if tags[TagInstanceID] == nil {
		return resources.Group{}, fmt.Errorf("serviceInstance is nil")
	}

	serviceInstanceID := *tags[TagInstanceID]
	filter := fmt.Sprintf("tagName eq 'InstanceID' and tagValue eq '%s'", serviceInstanceID)

	// we only expect one ResourceGroup, so not using pagination here should be fine
	resourceGroupIterator, err := nc.resourcegroupClient.ListComplete(ctx, filter, nil)
	if err != nil {
		return resources.Group{}, err
	}
	for resourceGroupIterator.NotDone() {
		resourceGroup := resourceGroupIterator.Value()

		// there should only be one resource group with the given service instance id
		if err := resourceGroupIterator.NextWithContext(ctx); err != nil {
			return resources.Group{}, errors.Wrapf(err, "while getting resource group for service instance: %s", serviceInstanceID)
		}
		return resourceGroup, nil
	}
	return resources.Group{}, NewResourceGroupDoesNotExist(fmt.Sprintf("no resource group found for service instance id: %s", serviceInstanceID))
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
