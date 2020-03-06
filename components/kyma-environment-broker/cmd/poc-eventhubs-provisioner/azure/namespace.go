package azure

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/go-autorest/autorest/to"
)

func PersistEventHubsNamespace(ctx context.Context, config *Config, groupName, namespace string) (*eventhub.EHNamespace, error) {
	nsClient := getNamespacesClientOrDie(config)
	future, err := nsClient.CreateOrUpdate(ctx, groupName, namespace, eventhub.EHNamespace{Location: to.StringPtr(config.location)})
	if err != nil {
		return nil, err
	}

	err = future.WaitForCompletionRef(ctx, nsClient.Client)
	if err != nil {
		return nil, err
	}

	result, err := future.Result(nsClient)
	return &result, err
}

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

func ListReadyNamespaces(ctx context.Context, config *Config) (eventhub.EHNamespace, error) {
	nsClient := getNamespacesClientOrDie(config)
	ehNamespaceIterator, err := nsClient.ListComplete(ctx)
	if err != nil {
		return eventhub.EHNamespace{}, err
	}
	for ehNamespaceIterator.NotDone() {
		ehNamespace := ehNamespaceIterator.Value()
		if val, ok := ehNamespace.Tags["ready"]; ok {
			isReady, err := strconv.ParseBool(*val)
			if err == nil && isReady {
				return ehNamespace, nil
			}
		}
		log.Printf("EHNamespace: %+v", ehNamespace)
		if err := ehNamespaceIterator.NextWithContext(ctx); err != nil {
			return eventhub.EHNamespace{}, err
		}
	}
	return eventhub.EHNamespace{}, fmt.Errorf("no ready EHNamespace found")
}
