package oauth20_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20/automock"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	testErr := errors.New("test err")

	testCases := []struct {
		Name                       string
		ExpectedResult             *model.OAuthCredentialDataInput
		ExpectedError              error
		ClientDetailsCfgProviderFn func() *automock.ClientDetailsConfigProvider
		UIDServiceFn               func() *automock.UIDService
		HydraClient                func() *automock.OryHydraService
		ObjectType                 pkgmodel.SystemAuthReferenceObjectType
	}{
		{
			Name:           "Success",
			ExpectedError:  nil,
			ExpectedResult: successResult,
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(clientID).Once()
				return uidSvc
			},
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(grantTypes, nil).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("CreateOAuth2Client", mock.Anything).Return(&admin.CreateOAuth2ClientCreated{Payload: &models.OAuth2Client{ClientSecret: clientSecret}}, nil).Once()
				return hydra
			},
			ObjectType: objType,
		},
		{
			Name:          "Error when client registration in hydra fails",
			ExpectedError: testErr,
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(clientID).Once()
				return uidSvc
			},
			ClientDetailsCfgProviderFn: func() *automock.ClientDetailsConfigProvider {
				clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
				clientDetailsCfgProvider.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil).Once()
				clientDetailsCfgProvider.On("GetRequiredGrantTypes", "clientCredentialsRegistrationGrantTypes").Return(grantTypes, nil).Once()
				return clientDetailsCfgProvider
			},
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("CreateOAuth2Client", mock.Anything).Return(&admin.CreateOAuth2ClientCreated{}, testErr).Once()
				return hydra
			},
			ObjectType: objType,
		},
		{
			Name:          "Error when cannot get client credentials scopes",
			ExpectedError: testErr,
			UIDServiceFn: func() *automock.UIDService {
				return &automock.UIDService{}
			},
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
			UIDServiceFn: func() *automock.UIDService {
				return &automock.UIDService{}
			},
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
			uidService := testCase.UIDServiceFn()
			defer uidService.AssertExpectations(t)
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(clientDetailsCfgProvider, uidService, publicEndpoint, hydraService)

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
				hydra.On("UpdateOAuth2Client", mock.Anything).Return(nil, nil).Once()
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
				hydra.On("UpdateOAuth2Client", mock.Anything).Return(nil, testErr).Once()
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
			uidService := &automock.UIDService{}
			defer uidService.AssertExpectations(t)
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(clientDetailsCfgProvider, uidService, publicEndpoint, hydraService)

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
				hydra.On("DeleteOAuth2Client", mock.Anything).Return(nil, nil).Once()
				return hydra
			},
		},
		{
			Name:          "Fails when hydra cannot delete client",
			ExpectedError: testErr,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("DeleteOAuth2Client", mock.Anything).Return(nil, testErr).Once()
				return hydra
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(nil, nil, publicEndpoint, hydraService)

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
				hydra.On("DeleteOAuth2Client", mock.Anything).Return(nil, nil)
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
				hydra.On("DeleteOAuth2Client", admin.NewDeleteOAuth2ClientParams().WithID(clientID)).Return(nil, testErr)
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

			svc := oauth20.NewService(nil, nil, publicEndpoint, hydraService)

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
				hydra.On("ListOAuth2Clients", mock.Anything).Return(&admin.ListOAuth2ClientsOK{Payload: []*models.OAuth2Client{{ClientSecret: clientSecret}}}, nil).Once()
				return hydra
			},
		},
		{
			Name:          "Fails when hydra cannot list clients",
			ExpectedError: testErr,
			HydraClient: func() *automock.OryHydraService {
				hydra := &automock.OryHydraService{}
				hydra.On("ListOAuth2Clients", mock.Anything).Return(nil, testErr).Once()
				return hydra
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			clientDetailsCfgProvider := &automock.ClientDetailsConfigProvider{}
			defer clientDetailsCfgProvider.AssertExpectations(t)
			uidService := &automock.UIDService{}
			defer uidService.AssertExpectations(t)
			hydraService := testCase.HydraClient()
			defer hydraService.AssertExpectations(t)

			svc := oauth20.NewService(clientDetailsCfgProvider, uidService, publicEndpoint, hydraService)

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
