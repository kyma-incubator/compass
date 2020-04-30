package azure

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

type AzureInterface interface {
	GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	CreateResourceGroup(ctx context.Context, config *Config, name string, tags Tags) (resources.Group, error)
	CreateNamespace(ctx context.Context, azureCfg *Config, groupName, namespace string, tags Tags) (*eventhub.EHNamespace, error)
	// CheckNamespaceAvailability check an Event Hubs namespace name availability
	CheckNamespaceAvailability(ctx context.Context, name string) (bool, error)
	CheckResourceGroupAvailability(ctx context.Context, name string) (bool, error)
	NamespaceExists(ctx context.Context, resourceGroupName string, namespaceName string, tags Tags) (bool, error)
	ResourceGroupExists(ctx context.Context, name string, tags Tags) (bool, error)
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
	ehNamespace, err := nc.createAndWait(ctx, groupName, namespace, parameters)
	return &ehNamespace, err
}

func (nc *AzureClient) CheckNamespaceAvailability(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, errors.New("Name cannot be empty")
	}
	res, err := nc.eventhubNamespaceClient.CheckNameAvailability(ctx, eventhub.CheckNameAvailabilityParameter{Name: &name})
	if err != nil {
		return false, errors.Wrapf(err, "Failed to check Event Hubs namespace availability with name: %s", name)
	}
	if res.NameAvailable == nil {
		return false, errors.Errorf("Failed to check Event Hubs namespace availability with name: %s. Received no response using Azure client.", name)
	}
	return *res.NameAvailable, nil
}

func (nc *AzureClient) CheckResourceGroupAvailability(ctx context.Context, name string) (bool, error) {
	var available bool
	if name == "" {
		return false, errors.New("Resource group name cannot be empty")
	}
	group, err := nc.resourcegroupClient.CheckExistence(ctx, name)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to check Azure resource group availability with name: %s. Error "+
			"while using Azure client.", name)
	}
	switch group.StatusCode {
	case 404:
		available = true
	case 204:
		available = false
	default:
		return false, errors.Errorf("Failed to check Azure resource group availability with name: %s."+
			" Unexpected API response code %d", name, group.StatusCode)
	}
	return available, nil
}

func (nc *AzureClient) NamespaceExists(ctx context.Context, resourceGroupName string, namespaceName string, tags Tags) (bool, error) {
	exists := false
	if namespaceName == "" {
		return false, errors.New("Namespace name cannot be empty")
	}
	if resourceGroupName == "" {
		return false, errors.New("Resource group name cannot be empty")
	}
	exists, err := nc.CheckResourceGroupAvailability(ctx, resourceGroupName)

	if err != nil {
		return false, errors.Wrapf(err, "Failed to check Azure Event Hubs namespace availability with name: %s.", namespaceName)
	}
	if !exists {
		return false, nil
	}

	res, err := nc.eventhubNamespaceClient.ListByResourceGroup(ctx, resourceGroupName)
	if err != nil {
		return false, errors.Errorf("Failed to check Event Hubs namespace availability with name: %s in resource"+
			"group %s. Error while using Azure client. %v", namespaceName, resourceGroupName, err)
	}
	if res.Response().StatusCode != 200 && res.Response().StatusCode != 201 {
		return false, errors.Errorf("Failed to check Event Hubs namespace availability with name: %s in resource"+
			"group %s. Unexpected API response code. Expected 2XX but received %d", namespaceName, resourceGroupName,
			res.Response().StatusCode)
	}
	for res.NotDone() {
		namespaces := res.Values()
		for _, ns := range namespaces {
			exists = matchWithTags(namespaceName, tags, *ns.Name, ns.Tags)
		}
		err := res.NextWithContext(ctx)
		if err != nil {
			return false, errors.Errorf("Failed to check Event Hubs namespace availability with name: %s in resource"+
				"group %s. Error while listing namespaces. %v", namespaceName, resourceGroupName, err)
		}
	}
	return exists, nil
}

func (nc *AzureClient) ResourceGroupExists(ctx context.Context, name string, tags Tags) (bool, error) {
	if name == "" {
		return false, errors.New("Resource group name cannot be empty")
	}
	// Will not return an error if resource group does not exist
	exists, err := nc.CheckResourceGroupAvailability(ctx, name)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to check Azure resource group availability with name: %s.", name)
	}
	if !exists {
		return false, nil
	}
	// Now we are safe to assume resource group exists, so any errors returned are not related to resource group not
	// existing
	group, err := nc.resourcegroupClient.Get(ctx, name)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to check Azure resource group availability with name: %s. Error "+
			"while using Azure client.", name)
	}
	if group.StatusCode != 200 {
		return false, errors.Errorf("Failed to check Azure resource group availability with name: %s."+
			" Unexpected API response code. Expected 2XX but received %d", name, group.StatusCode)
	}
	return matchWithTags(name, tags, *group.Name, group.Tags), nil
}

func matchWithTags(nameA string, tagsA Tags, nameB string, tagsB Tags) bool {
	for k, v := range tagsA {
		if t, ok := tagsB[k]; ok {
			if v != t {
				// same value
				return false
			}
			// values are different
			continue
		}
		// tag does not exist
		return false
	}
	return nameA == nameB
}

func (nc *AzureClient) createAndWait(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
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
