package azure

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
)

const EHNamespaceTagInUse = "in_use"

func GetEventHubsNamespaceAccessKeys(ctx context.Context, config *Config, resourceGroupName, namespaceName, authorizationRuleName string) (*eventhub.AccessKeys, error) {
	nsClient := getNamespacesClientOrDie(config)
	accessKeys, err := nsClient.ListKeys(ctx, resourceGroupName, namespaceName, authorizationRuleName)
	return &accessKeys, err
}

func getNamespacesClientOrDie(config *Config) eventhub.NamespacesClient {
	nsClient := eventhub.NewNamespacesClient(config.subscriptionID)

	authorizer, err := GetResourceManagementAuthorizer(config)
	if err != nil {
		log.Fatalf("Failed to initialize authorizer with error: %v", err)
	}
	nsClient.Authorizer = authorizer

	if err = nsClient.AddToUserAgent(config.userAgent); err != nil {
		log.Fatalf("Failed to add use agent [%s] with error: %v", config.userAgent, err)
	}

	return nsClient
}

// MarkNamespaceAsUsed sets a tag to indicate that the Namespace is used
func MarkNamespaceAsUsed(ctx context.Context, config *Config, resourceGroupName string, namespace eventhub.EHNamespace) (eventhub.EHNamespace, error) {
	log.Printf("Marking namespace to be unused, namespace: %+v, resourceGroup: %s", namespace, resourceGroupName)
	nsClient := getNamespacesClientOrDie(config)
	trueVal := strconv.FormatBool(true)
	namespace.Tags[EHNamespaceTagInUse] = &trueVal
	updatedEHNamespace, err := nsClient.Update(ctx, resourceGroupName, *namespace.Name, namespace)
	return updatedEHNamespace, err
}

// GetResourceGroup extract the ResouceGroup from a given EventHub Namespace
func GetResourceGroup(namespace eventhub.EHNamespace) string {
	// id has the following format "/subscriptions/<subscription>/resourceGroups/<resource-group>/providers/Microsoft.EventHub/namespaces/<namespace-name>"
	// the code extract <resource-group> from the string
	return strings.Split(strings.Split(*namespace.ID, "resourceGroups/")[1], "/")[0]
}

func GetFirstUnusedNamespaces(ctx context.Context, config *Config) (eventhub.EHNamespace, error) {
	nsClient := getNamespacesClientOrDie(config)
	// TODO(nachtmaar): optimize ?
	ehNamespaceIterator, err := nsClient.ListComplete(ctx)
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
