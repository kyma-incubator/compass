package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "while initializing authorizer")
	}

	// create namespace client
	nsClient, err := ac.getNamespaceClient(config, authorizer)
	if err != nil {
		return nil, errors.Wrap(err, "while creating namespace client")
	}

	// create resource group client
	resourcegroupClient, err := ac.getGroupsClient(config, authorizer)
	if err != nil {
		return nil, errors.Wrap(err, "while creating resource group client")
	}

	// create azure client
	return NewAzureClient(nsClient, resourcegroupClient), nil
}

// getGroupsClient gets a client for handling of Azure Namespaces
func (ac *azureProvider) getNamespaceClient(config *Config, authorizer autorest.Authorizer) (eventhub.NamespacesClient, error) {
	nsClient := eventhub.NewNamespacesClient(config.subscriptionID)
	nsClient.Authorizer = authorizer

	if err := nsClient.AddToUserAgent(config.userAgent); err != nil {
		return eventhub.NamespacesClient{}, errors.Wrapf(err, "while adding user agent [%s]", config.userAgent)
	}
	return nsClient, nil
}

// getGroupsClient gets a client for handling of Azure ResourceGroups
func (ac *azureProvider) getGroupsClient(config *Config, authorizer autorest.Authorizer) (resources.GroupsClient, error) {
	client := resources.NewGroupsClient(config.subscriptionID)
	client.Authorizer = authorizer

	if err := client.AddToUserAgent(config.userAgent); err != nil {
		return resources.GroupsClient{}, errors.Wrapf(err, "while adding user agent [%s]", config.userAgent)
	}

	return client, nil
}

func (ac *azureProvider) getResourceManagementAuthorizer(config *Config, environment *azure.Environment) (autorest.Authorizer, error) {
	armAuthorizer, err := ac.getAuthorizerForResource(config, environment)
	if err != nil {
		return nil, errors.Wrap(err, "while creating resource authorizer")
	}

	return armAuthorizer, err
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
