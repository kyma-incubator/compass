package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/sirupsen/logrus"
)

const (
	FutureOperationSucceeded string = "Succeeded"
	FutureOperationDeleting  string = "Deleting"
)

type Interface interface {
	GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	CreateResourceGroup(ctx context.Context, config *Config, name string, tags Tags) (resources.Group, error)
	CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags Tags) (*eventhub.EHNamespace, error)
	GetResourceGroup(ctx context.Context, tags Tags) (resources.Group, error)
	DeleteResourceGroup(ctx context.Context, tags Tags) (resources.GroupsDeleteFuture, error)
}

var _ Interface = (*Client)(nil)

type Client struct {
	// the actual azure client
	eventhubNamespaceClient eventhub.NamespacesClient
	resourcegroupClient     resources.GroupsClient
	logger                  logrus.FieldLogger
}

type ResourceGroupDoesNotExist struct {
	errorMessage string
}

func NewResourceGroupDoesNotExist(errorMessage string) ResourceGroupDoesNotExist {
	return ResourceGroupDoesNotExist{errorMessage}
}

func (e ResourceGroupDoesNotExist) Error() string {
	return e.errorMessage
}

func NewAzureClient(namespaceClient eventhub.NamespacesClient, resourcegroupClient resources.GroupsClient, logger logrus.FieldLogger) *Client {
	return &Client{
		eventhubNamespaceClient: namespaceClient,
		resourcegroupClient:     resourcegroupClient,
		logger:                  logger,
	}
}

func (nc *Client) GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	return nc.eventhubNamespaceClient.ListKeys(ctx, resourceGroupName, namespaceName, authorizationRuleName)
}

func (nc *Client) CreateResourceGroup(ctx context.Context, config *Config, name string, tags Tags) (resources.Group, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := config.GetLocation()
	return nc.resourcegroupClient.CreateOrUpdate(ctx, name, resources.Group{Location: &locationCopy, Tags: tags})
}

func (nc *Client) CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags Tags) (*eventhub.EHNamespace, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := azureCfg.GetLocation()
	parameters := eventhub.EHNamespace{Location: &locationCopy, Tags: tags}
	ehNamespace, err := nc.createNamespaceAndWait(ctx, groupName, namespace, parameters)
	return &ehNamespace, err
}

func (nc *Client) DeleteResourceGroup(ctx context.Context, tags Tags) (resources.GroupsDeleteFuture, error) {
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
	return future, err
}

// GetResourceGroup gets the resource group by tags.
// If more than one resource group is found, it is treated as an error.
func (nc *Client) GetResourceGroup(ctx context.Context, tags Tags) (resources.Group, error) {
	if tags[TagInstanceID] == nil {
		return resources.Group{}, fmt.Errorf("serviceInstance is nil")
	}

	serviceInstanceID := *tags[TagInstanceID]
	filter := fmt.Sprintf("tagName eq 'InstanceID' and tagValue eq '%s'", serviceInstanceID)

	// we only expect one ResourceGroup, so not using pagination here should be fine
	resourceGroupIterator, err := nc.resourcegroupClient.List(ctx, filter, nil)
	if err != nil {
		return resources.Group{}, err
	}
	// values() gives us all values because `top` is set to nil in the List call - so there is only one page
	resourceGroups := resourceGroupIterator.Values()

	if len(resourceGroups) > 1 {
		return resources.Group{}, fmt.Errorf("only one resource group expected with the given instance id: %s, found: %d", serviceInstanceID, len(resourceGroups))
	} else if len(resourceGroups) == 1 {
		// we found it
		return resourceGroups[0], nil
	}
	return resources.Group{}, NewResourceGroupDoesNotExist(fmt.Sprintf("no resource group found for service instance id: %s", serviceInstanceID))
}

func (nc *Client) createNamespaceAndWait(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
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
