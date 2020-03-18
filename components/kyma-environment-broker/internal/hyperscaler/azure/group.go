package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

// getGroupsClient gets a client for handling of Azure ResourceGroups
func getGroupsClient(config *Config) (resources.GroupsClient, error) {
	client := resources.NewGroupsClient(config.subscriptionID)

	authorizer, err := GetResourceManagementAuthorizer(config)
	if err != nil {
		return resources.GroupsClient{}, errors.Wrap(err, "while initializing authorizer")
	}
	client.Authorizer = authorizer

	if err = client.AddToUserAgent(config.userAgent); err != nil {
		return resources.GroupsClient{}, fmt.Errorf("failed to add use agent [%s] with error: %v", config.userAgent, err)
	}

	return client, nil
}
