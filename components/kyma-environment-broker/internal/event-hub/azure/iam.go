package azure

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

var (
	armAuthorizer autorest.Authorizer
)

func GetResourceManagementAuthorizer(config *Config) (autorest.Authorizer, error) {
	if armAuthorizer != nil {
		return armAuthorizer, nil
	}

	environment, err := config.Environment()
	if err != nil {
		return nil, err
	}
	a, err := getAuthorizerForResource(config, environment.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}

	// cache
	armAuthorizer = a
	return armAuthorizer, err
}

func getAuthorizerForResource(config *Config, resource string) (autorest.Authorizer, error) {

	environment, err := config.Environment()
	if err != nil {
		return nil, err
	}
	oauthConfig, err := adal.NewOAuthConfig(environment.ActiveDirectoryEndpoint, config.tenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(*oauthConfig, config.clientID, config.clientSecret, resource)
	if err != nil {
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), err
}
