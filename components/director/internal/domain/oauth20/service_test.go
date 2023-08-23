package oauth20_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	hydraClient "github.com/ory/hydra-client-go/v2"
)

const (
	publicEndpoint = "accessTokenURL"
	clientID       = "clientid"
	clientSecret   = "secret"
	objType        = pkgmodel.IntegrationSystemReference
)

var (
	scopes     = []string{"foo", "bar", "baz"}
	grantTypes = []string{"client_credentials"}
)

func TestService_CreateClient(t *testing.T) {
	// GIVEN
	successResult := &model.OAuthCredentialDataInput{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          publicEndpoint,
	}

	scopesCreate := strings.Join(scopes, " ")
	clientToCreate := hydraClient.OAuth2Client{
		GrantTypes: grantTypes,
		Scope:      &scopesCreate,
	}
	createRequest := hydraClient.OAuth2ApiCreateOAuth2ClientRequest{}.OAuth2Client(clientToCreate)

	testErr := errors.New("test err")

	testCases := []struct {
		Name                       string
		ExpectedResult             *model.OAuthCredentialDataInput
		ExpectedError              error
		ClientDetailsCfgProviderFn func() *automock.ClientDetailsConfigProvider
		HydraClient                func() *automock.OryHydraService
		ObjectType                 pkgmodel.SystemAuthReferenceObjectType
	}{
		{
			Name:           "Success",
			ExpectedError:  nil,
			ExpectedResult: successResult,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(grantTypes, nil).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("CreateOAuth2Client", mock.Anything).Return(createRequest).Once()
				hydra.On("CreateOAuth2ClientExecute", createRequest).Return(&hydraClient.OAuth2Client{ClientSecret: &successResult.ClientSecret, ClientId: &successResult.ClientID}, nil, nil).Once()
				return hydra
			},
			ObjectType: objType,
		},
		{
			Name:          "Error when client registration in hydra fails",
			ExpectedError: testErr,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(grantTypes, nil).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("CreateOAuth2Client", mock.Anything).Return(createRequest).Once()
				hydra.On("CreateOAuth2ClientExecute", createRequest).Return(nil, nil, testErr).Once()
				return hydra
			},
			ObjectType: objType,
		},
		{
			Name:          "Error when cannot get client credentials scopes",
			ExpectedError: testErr,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(nil, testErr).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				return &automock.OryHydraService{}
			},
			ObjectType: pkgmodel.ApplicationReference,
		},
		{
			Name:          "Error when cannot get client grant types",
			ExpectedError: testErr,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(nil, testErr).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				return &automock.OryHydraService{}
			},
			ObjectType: pkgmodel.ApplicationReference,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			clientDetailsCfgProvider := testCase.ClientDetailsCfgProviderFn()
			defer clientDetailsCfgProvider.AssertExpectations(t)
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(clientDetailsCfgProvider, publicEndpoint, hydraService)

			// WHEN
			oauthData, err := svc.CreateClientCredentials(ctx, testCase.ObjectType)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, oauthData)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}

func TestService_UpdateClient(t *testing.T) {
	// GIVEN
	clientUpdate := clientID
	scopesCreate := strings.Join(scopes, " ")
	clientToUpgrade := hydraClient.OAuth2Client{
		ClientId:   &clientUpdate,
		GrantTypes: grantTypes,
		Scope:      &scopesCreate,
	}
	setRequest := hydraClient.OAuth2ApiSetOAuth2ClientRequest{}.OAuth2Client(clientToUpgrade)

	testErr := errors.New("test err")
	testCases := []struct {
		Name                       string
		ExpectedError              error
		ClientDetailsCfgProviderFn func() *automock.ClientDetailsConfigProvider
		HydraClient                func() *automock.OryHydraService
		ObjectType                 pkgmodel.SystemAuthReferenceObjectType
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(grantTypes, nil).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("SetOAuth2Client", mock.Anything, clientID).Return(setRequest).Once()
				hydra.On("SetOAuth2ClientExecute", setRequest).Return(nil, nil, nil).Once()
				return hydra
			},
			ObjectType: objType,
		},
		{
			Name:          "Error when client update in hydra fails",
			ExpectedError: testErr,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(grantTypes, nil).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("SetOAuth2Client", mock.Anything, clientID).Return(setRequest).Once()
				hydra.On("SetOAuth2ClientExecute", setRequest).Return(nil, nil, testErr).Once()
				return hydra
			},
			ObjectType: objType,
		},
		{
			Name:          "Error when cannot get client credentials scopes",
			ExpectedError: testErr,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(nil, testErr).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				return &automock.OryHydraService{}
			},
			ObjectType: pkgmodel.ApplicationReference,
		},
		{
			Name:          "Error when cannot get client grant types",
			ExpectedError: testErr,
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.application").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(nil, testErr).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				return &automock.OryHydraService{}
			},
			ObjectType: pkgmodel.ApplicationReference,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			clientDetailsCfgProvider := testCase.ClientDetailsCfgProviderFn()
			defer clientDetailsCfgProvider.AssertExpectations(t)
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(clientDetailsCfgProvider, publicEndpoint, hydraService)

			// WHEN
			err := svc.UpdateClient(ctx, clientID, testCase.ObjectType)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}

func TestService_DeleteClientCredentials(t *testing.T) {
	// GIVEN
	id := "foo"
	oauth2DeleteRequest := hydraClient.OAuth2ApiDeleteOAuth2ClientRequest{}
	testErr := errors.New("test err")
	testCases := []struct {
		Name          string
		ExpectedError error
		HydraClient   func() *automock.OryHydraService
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("DeleteOAuth2Client", mock.Anything, id).Return(oauth2DeleteRequest).Once()
				hydra.On("DeleteOAuth2ClientExecute", oauth2DeleteRequest).Return(nil, nil).Once()
				return hydra
			},
		},
		{
			Name:          "Fails when hydra cannot delete client",
			ExpectedError: testErr,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("DeleteOAuth2Client", mock.Anything, id).Return(oauth2DeleteRequest).Once()
				hydra.On("DeleteOAuth2ClientExecute", oauth2DeleteRequest).Return(nil, testErr).Once()
				return hydra
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(nil, publicEndpoint, hydraService)

			// WHEN
			err := svc.DeleteClientCredentials(ctx, id)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}

func TestService_DeleteMultipleClientCredentials(t *testing.T) {
	// GIVEN
	testErr := errors.New("test err")
	oauth2DeleteRequest := hydraClient.OAuth2ApiDeleteOAuth2ClientRequest{}
	testCases := []struct {
		Name          string
		ExpectedError error
		HydraClient   func() *automock.OryHydraService
		Auths         []pkgmodel.SystemAuth
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("DeleteOAuth2Client", mock.Anything, mock.Anything).Return(oauth2DeleteRequest).Once()
				hydra.On("DeleteOAuth2ClientExecute", oauth2DeleteRequest).Return(nil, nil).Once()
				return hydra
			},
			Auths: []pkgmodel.SystemAuth{
				{
					Value: &model.Auth{
						Credential: model.CredentialData{
							Oauth: &model.OAuthCredentialData{
								ClientID: clientID,
							},
						},
					},
				},
			},
		},
		{
			Name:          "Will not delete auth when value is nil",
			ExpectedError: nil,
			HydraClient: func() *automock.OryHydraService {
				return &automock.OryHydraService{}
			},
			Auths: []pkgmodel.SystemAuth{
				{
					Value: nil,
				},
			},
		},
		{
			Name:          "Will not delete auth when Oauth is nil",
			ExpectedError: nil,
			HydraClient: func() *automock.OryHydraService {
				return &automock.OryHydraService{}
			},
			Auths: []pkgmodel.SystemAuth{
				{
					Value: &model.Auth{
						Credential: model.CredentialData{
							Oauth: nil,
						},
					},
				},
			},
		},
		{
			Name:          "Fails when hydra cannot delete client",
			ExpectedError: testErr,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("DeleteOAuth2Client", mock.Anything, clientID).Return(oauth2DeleteRequest).Once()
				hydra.On("DeleteOAuth2ClientExecute", oauth2DeleteRequest).Return(nil, testErr).Once()
				return hydra
			},
			Auths: []pkgmodel.SystemAuth{
				{
					Value: &model.Auth{
						Credential: model.CredentialData{
							Oauth: &model.OAuthCredentialData{
								ClientID: clientID,
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(nil, publicEndpoint, hydraService)

			// WHEN
			err := svc.DeleteMultipleClientCredentials(ctx, testCase.Auths)

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}
func TestService_ListClients(t *testing.T) {
	// GIVEN
	testErr := errors.New("test err")
	oauth2ListRequest := hydraClient.OAuth2ApiListOAuth2ClientsRequest{}
	testCases := []struct {
		Name          string
		ExpectedError error
		HydraClient   func() *automock.OryHydraService
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				clientSecret := "secret"
				hydra.On("ListOAuth2Clients", mock.Anything).Return(oauth2ListRequest).Once()
				hydra.On("ListOAuth2ClientsExecute", oauth2ListRequest).Return([]hydraClient.OAuth2Client{{ClientSecret: &clientSecret}}, nil, nil).Once()
				return hydra
			},
		},
		{
			Name:          "Fails when hydra cannot list clients",
			ExpectedError: testErr,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("ListOAuth2Clients", mock.Anything).Return(oauth2ListRequest).Once()
				hydra.On("ListOAuth2ClientsExecute", oauth2ListRequest).Return(nil, nil, testErr).Once()
				return hydra
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
			hydraService := testCase.HydraClient()

			svc := oauth20.NewService(clientDetailsCfgProvider, publicEndpoint, hydraService)

			// WHEN
			clients, err := svc.ListClients()

			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Len(t, clients, 1)
			} else {
				require.Error(t, err)
				require.Len(t, clients, 0)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}
}
