package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

func PersistResourceGroup(ctx context.Context, config *Config, name string) (resources.Group, error) {
	client := getGroupsClientOrDie(config)
	// we need to use a copy of the location, because the following azure call will modify it
	locationCopy := config.GetLocation()
	return client.CreateOrUpdate(ctx, name, resources.Group{Location: &locationCopy})
}

func getGroupsClientOrDie(config *Config) resources.GroupsClient {
	client := resources.NewGroupsClient(config.subscriptionID)

	authorizer, err := GetResourceManagementAuthorizer(config)
	if err != nil {
		log.Fatalf("Failed to initialize authorizer with error: %v", err)
	}
	client.Authorizer = authorizer

	if err = client.AddToUserAgent(config.userAgent); err != nil {
		log.Fatalf("Failed to add use agent [%s] with error: %v", config.userAgent, err)
	}

	return client
}
