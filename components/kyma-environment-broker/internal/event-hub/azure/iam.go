package azure

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

var (
	armAuthorizer autorest.Authorizer
)

type OAuthGrantType int

const (
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	OAuthGrantTypeDeviceFlow
)

func grantType() OAuthGrantType {
	return OAuthGrantTypeServicePrincipal
}

func GetResourceManagementAuthorizer(config *Config) (autorest.Authorizer, error) {
	if armAuthorizer != nil {
		return armAuthorizer, nil
	}

	var a autorest.Authorizer
	var err error

	a, err = getAuthorizerForResource(config, grantType(), config.Environment().ResourceManagerEndpoint)

	if err == nil {
		// cache
		armAuthorizer = a
	} else {
		// clear cache
		armAuthorizer = nil
	}
	return armAuthorizer, err
}

func getAuthorizerForResource(config *Config, grantType OAuthGrantType, resource string) (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error

	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		oauthConfig, err := adal.NewOAuthConfig(config.Environment().ActiveDirectoryEndpoint, config.tenantID)
		if err != nil {
			return nil, err
		}

		token, err := adal.NewServicePrincipalToken(*oauthConfig, config.clientID, config.clientSecret, resource)
		if err != nil {
			return nil, err
		}
		a = autorest.NewBearerAuthorizer(token)

	// TODO(nachtmaar): delete me
	case OAuthGrantTypeDeviceFlow:
		deviceConfig := auth.NewDeviceFlowConfig(config.clientID, config.tenantID)
		deviceConfig.Resource = resource
		a, err = deviceConfig.Authorizer()
		if err != nil {
			return nil, err
		}

	default:
		return a, fmt.Errorf("invalid grant type specified")
	}

	return a, err
}
