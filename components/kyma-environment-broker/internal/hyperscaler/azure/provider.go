package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type HyperscalerProvider interface {
	GetClient(config *Config) (AzureInterface, error)
}

var _ HyperscalerProvider = (*azureProvider)(nil)

type azureProvider struct{}

func NewAzureProvider() HyperscalerProvider {
	return &azureProvider{}
}

// GetClient gets a client for interacting with Azure
func (ac *azureProvider) GetClient(config *Config) (AzureInterface, error) {

	environment, err := config.Environment()
	if err != nil {
		return nil, err
	}

	authorizer, err := ac.getResourceManagementAuthorizer(config, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authorizer with error: %v", err)
	}

	// create namespace client
	nsClient, err := ac.getNamespaceClient(config, authorizer)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace client with error: %v", err)
	}

	// create resource group client
	resourcegroupClient, err := ac.getGroupsClient(config, authorizer)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource group client with error: %v", err)
	}

	// create azure client
	return NewAzureClient(nsClient, resourcegroupClient), nil
}

// getGroupsClient gets a client for handling of Azure Namespaces
func (ac *azureProvider) getNamespaceClient(config *Config, authorizer autorest.Authorizer) (eventhub.NamespacesClient, error) {
	nsClient := eventhub.NewNamespacesClient(config.subscriptionID)
	nsClient.Authorizer = authorizer

	if err := nsClient.AddToUserAgent(config.userAgent); err != nil {
		return eventhub.NamespacesClient{}, fmt.Errorf("failed to add use agent [%s] with error: %v", config.userAgent, err)
	}
	return nsClient, nil
}

// getGroupsClient gets a client for handling of Azure ResourceGroups
func (ac *azureProvider) getGroupsClient(config *Config, authorizer autorest.Authorizer) (resources.GroupsClient, error) {
	client := resources.NewGroupsClient(config.subscriptionID)
	client.Authorizer = authorizer

	if err := client.AddToUserAgent(config.userAgent); err != nil {
		return resources.GroupsClient{}, fmt.Errorf("failed to add use agent [%s] with error: %v", config.userAgent, err)
	}

	return client, nil
}

func (ac *azureProvider) getResourceManagementAuthorizer(config *Config, environment *azure.Environment) (autorest.Authorizer, error) {
	armAuthorizer, err := ac.getAuthorizerForResource(config, environment)
	if err != nil {
		return nil, err
	}

	return armAuthorizer, err
}

func (ac *azureProvider) getEnvironment(config *Config) (*azure.Environment, error) {
	environment, err := config.Environment()
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (ac *azureProvider) getAuthorizerForResource(config *Config, environment *azure.Environment) (autorest.Authorizer, error) {

	oauthConfig, err := adal.NewOAuthConfig(environment.ActiveDirectoryEndpoint, config.tenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig, config.clientID, config.clientSecret, environment.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), err
}
