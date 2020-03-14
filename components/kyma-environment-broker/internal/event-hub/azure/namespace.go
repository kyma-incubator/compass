package azure

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

const EHNamespaceTagInUse = "in_use"

type NamespaceClientInterface interface {
	ListKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error)
	Update(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error)
	ListComplete(ctx context.Context) (result eventhub.EHNamespaceListResultIterator, err error)
	CreateOrUpdate(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error)
	PersistResourceGroup(ctx context.Context, config *Config, name string) (resources.Group, error)
	PersistEventHubsNamespace(ctx context.Context, azureCfg *Config, namespaceClient NamespaceClientInterface, groupName, namespace string) (*eventhub.EHNamespace, error)
}

type NamespaceClient struct {
	namespaceClient eventhub.NamespacesClient
}

func NewNamespaceClient(client eventhub.NamespacesClient) NamespaceClientInterface {
	return &NamespaceClient{
		namespaceClient: client,
	}
}
func (nc *NamespaceClient) ListKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	return nc.namespaceClient.ListKeys(ctx, resourceGroupName, namespaceName, authorizationRuleName)
}

func (nc *NamespaceClient) Update(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
	return nc.namespaceClient.Update(ctx, resourceGroupName, namespaceName, parameters)
}

func (nc *NamespaceClient) ListComplete(ctx context.Context) (result eventhub.EHNamespaceListResultIterator, err error) {
	return nc.namespaceClient.ListComplete(ctx)
}

func (nc *NamespaceClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
	future, err := nc.namespaceClient.CreateOrUpdate(ctx, resourceGroupName, namespaceName, parameters)
	if err != nil {
		return eventhub.EHNamespace{}, err
	}

	err = future.WaitForCompletionRef(ctx, nc.namespaceClient.Client)
	if err != nil {
		return eventhub.EHNamespace{}, err
	}

	result, err = future.Result(nc.namespaceClient)
	return result, err
}

func (nc *NamespaceClient) PersistResourceGroup(ctx context.Context, config *Config, name string) (resources.Group, error) {
	return PersistResourceGroup(ctx, config, name)
}

func (nc *NamespaceClient) PersistEventHubsNamespace(ctx context.Context, azureCfg *Config, namespaceClient NamespaceClientInterface, groupName, namespace string) (*eventhub.EHNamespace, error) {
	return PersistEventHubsNamespace(ctx, azureCfg, namespaceClient, groupName, namespace)
}

func GetNamespacesClientOrDie(config *Config) NamespaceClientInterface {
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

// MarkNamespaceAsUsed sets a tag to indicate that the Namespace is used
func MarkNamespaceAsUsed(ctx context.Context, namespaceClient NamespaceClientInterface, resourceGroupName string, namespace eventhub.EHNamespace) (eventhub.EHNamespace, error) {
	log.Printf("Marking namespace to be unused, namespace: %+v, resourceGroup: %s", namespace, resourceGroupName)
	trueVal := strconv.FormatBool(true)
	namespace.Tags[EHNamespaceTagInUse] = &trueVal
	updatedEHNamespace, err := namespaceClient.Update(ctx, resourceGroupName, *namespace.Name, namespace)
	return updatedEHNamespace, err
}

// GetResourceGroup extract the ResouceGroup from a given EventHub Namespace
func GetResourceGroup(namespace eventhub.EHNamespace) string {
	// id has the following format "/subscriptions/<subscription>/resourceGroups/<resource-group>/providers/Microsoft.EventHub/namespaces/<namespace-name>"
	// the code extract <resource-group> from the string
	return strings.Split(strings.Split(*namespace.ID, "resourceGroups/")[1], "/")[0]
}

func GetFirstUnusedNamespaces(ctx context.Context, namespaceClient NamespaceClientInterface) (eventhub.EHNamespace, error) {
	// TODO(nachtmaar): optimize ?
	ehNamespaceIterator, err := namespaceClient.ListComplete(ctx)
	if err != nil {
		return eventhub.EHNamespace{}, err
	}
	for ehNamespaceIterator.NotDone() {
		ehNamespace := ehNamespaceIterator.Value()
		if val, ok := ehNamespace.Tags[EHNamespaceTagInUse]; ok {
			inUse, err := strconv.ParseBool(*val)
			if err == nil && !inUse {
				return ehNamespace, nil
			}
		}
		if err := ehNamespaceIterator.NextWithContext(ctx); err != nil {
			return eventhub.EHNamespace{}, err
		}
	}
	return eventhub.EHNamespace{}, fmt.Errorf("no ready EHNamespace found")
}

func PersistEventHubsNamespace(ctx context.Context, azureCfg *Config, namespaceClient NamespaceClientInterface, groupName, namespace string) (*eventhub.EHNamespace, error) {
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := azureCfg.GetLocation()
	parameters := eventhub.EHNamespace{Location: &locationCopy}
	ehNamespace, err := namespaceClient.CreateOrUpdate(ctx, groupName, namespace, parameters)
	return &ehNamespace, err
}
