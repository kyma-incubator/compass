package onetimetoken_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pairing"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	directorTime "github.com/kyma-incubator/compass/components/director/pkg/time"
	timeMocks "github.com/kyma-incubator/compass/components/director/pkg/time/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateOneTimeToken(t *testing.T) {
	const (
		tokenValue          = "abc"
		connectorURL        = "connector.url"
		appID               = "4c86b315-c027-467f-a6fc-b184ca0a80f1"
		runtimeID           = "31a607c7-695f-4a31-b2d1-777939f84aac"
		integrationSystemID = "123607c7-695f-4a31-b2d1-777939f84123"
	)

	fakeToken := &model.OneTimeToken{
		Used:         false,
		UsedAt:       time.Time{},
		Token:        tokenValue,
		ConnectorURL: connectorURL,
	}

	testCases := []struct {
		description               string
		objectID                  string
		connectorURL              string
		shouldHaveError           bool
		errorMsg                  string
		tokenType                 model.SystemAuthReferenceObjectType
		intSystemToAdapterMapping map[string]string
		systemAuthSvc             func() onetimetoken.SystemAuthService
		appSvc                    func() onetimetoken.ApplicationService
		appConverter              func() onetimetoken.ApplicationConverter
		tenantSvc                 func() onetimetoken.ExternalTenantsService
		httpClient                func() onetimetoken.HTTPDoer
		tokenGenerator            func() onetimetoken.TokenGenerator
		timeService               directorTime.Service
	}{
		{
			description: "Generate Application token, no int system, should succeed",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(&model.Application{}, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return(tokenValue, nil)
				return tokenGenerator
			},
			shouldHaveError:           false,
			objectID:                  appID,
			tokenType:                 model.ApplicationReference,
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
			timeService:               directorTime.NewService(),
		},
		{
			description: "Generate Application token should fail when no such app found",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				return &automock.SystemAuthService{}
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(nil, errors.New("not found"))
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				return &automock.TokenGenerator{}
			},
			shouldHaveError:           true,
			objectID:                  appID,
			tokenType:                 model.ApplicationReference,
			errorMsg:                  "not found",
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
		},
		{
			description: "Generate Application token should fail on db error",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", errors.New("db error"))
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(&model.Application{}, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return(tokenValue, nil)
				return tokenGenerator
			},
			shouldHaveError:           true,
			objectID:                  appID,
			tokenType:                 model.ApplicationReference,
			errorMsg:                  "db error",
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
			timeService:               directorTime.NewService(),
		},
		{
			description: "Generate Application token, with int system, should succeed",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				app := &model.Application{}
				app.IntegrationSystemID = str.Ptr(integrationSystemID)
				app.Tenant = "test-tenant"
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(app, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				mockAppConverter := &automock.ApplicationConverter{}
				givenGraphQLApp := graphql.Application{
					IntegrationSystemID: str.Ptr(integrationSystemID),
					BaseEntity: &graphql.BaseEntity{
						ID: appID,
					},
				}
				mockAppConverter.On("ToGraphQL", mock.Anything).Return(&givenGraphQLApp)
				return mockAppConverter
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				tenantSvc := &automock.ExternalTenantsService{}
				tenantSvc.On("GetExternalTenant", context.TODO(), "test-tenant").Return("external-tenant", nil)
				return tenantSvc
			},
			httpClient: func() onetimetoken.HTTPDoer {
				respBody := new(bytes.Buffer)
				respBody.WriteString(fmt.Sprintf(`{"token":"%s"}`, tokenValue))
				mockHttpClient := &automock.HTTPDoer{}
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(respBody),
				}
				mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					b, err := req.GetBody()
					if err != nil {
						return false
					}
					appData := pairing.RequestData{}
					err = json.NewDecoder(b).Decode(&appData)
					if err != nil {
						return false
					}
					tenantMatches := appData.Tenant == "external-tenant"
					clientUserMatches := appData.ClientUser == ""
					appIDMatches := appData.Application.ID == appID
					urlMatches := req.URL.String() == "https://my-integration-service.url"

					return urlMatches && appIDMatches && tenantMatches && clientUserMatches
				})).Return(response, nil)
				return mockHttpClient
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				return &automock.TokenGenerator{}
			},
			shouldHaveError: false,
			objectID:        appID,
			tokenType:       model.ApplicationReference,
			connectorURL:    connectorURL,
			intSystemToAdapterMapping: map[string]string{
				integrationSystemID: "https://my-integration-service.url",
			},
			timeService: directorTime.NewService(),
		},
		{
			description: "Generate Application token, with int system, but no adapters defined should succeed",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				app := &model.Application{}
				app.IntegrationSystemID = str.Ptr(integrationSystemID)
				app.Tenant = "test-tenant"
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(app, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return(tokenValue, nil)
				return tokenGenerator
			},
			shouldHaveError:           false,
			objectID:                  appID,
			tokenType:                 model.ApplicationReference,
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: map[string]string{},
			timeService:               directorTime.NewService(),
		},
		{
			description: "Generate Application token, with int system, should fail when int system fails",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				return &automock.SystemAuthService{}
			},
			appSvc: func() onetimetoken.ApplicationService {
				app := &model.Application{}
				app.IntegrationSystemID = str.Ptr(integrationSystemID)
				app.Tenant = "test-tenant"
				app.BaseEntity = &model.BaseEntity{
					ID: appID,
				}
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(app, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				mockAppConverter := &automock.ApplicationConverter{}
				givenGraphQLApp := graphql.Application{
					IntegrationSystemID: str.Ptr(integrationSystemID),
					BaseEntity: &graphql.BaseEntity{
						ID: appID,
					},
				}
				mockAppConverter.On("ToGraphQL", mock.Anything).Return(&givenGraphQLApp)
				return mockAppConverter
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				tenantSvc := &automock.ExternalTenantsService{}
				tenantSvc.On("GetExternalTenant", context.TODO(), "test-tenant").Return("external-tenant", nil)
				return tenantSvc
			},
			httpClient: func() onetimetoken.HTTPDoer {
				mockHttpClient := &automock.HTTPDoer{}
				response := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(&bytes.Buffer{}),
				}
				mockHttpClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					b, err := req.GetBody()
					if err != nil {
						return false
					}
					appData := pairing.RequestData{}
					err = json.NewDecoder(b).Decode(&appData)
					if err != nil {
						return false
					}
					tenantMatches := appData.Tenant == "external-tenant"
					clientUserMatches := appData.ClientUser == ""
					appIDMatches := appData.Application.ID == appID
					urlMatches := req.URL.String() == "https://my-integration-service.url"

					return urlMatches && appIDMatches && tenantMatches && clientUserMatches
				})).Return(response, nil).Times(3)
				return mockHttpClient
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				return &automock.TokenGenerator{}
			},
			shouldHaveError: true,
			objectID:        appID,
			tokenType:       model.ApplicationReference,
			errorMsg:        "wrong status code",
			connectorURL:    connectorURL,
			intSystemToAdapterMapping: map[string]string{
				integrationSystemID: "https://my-integration-service.url",
			},
		},
		{
			description: "Generate Application token, with int system, should fail when no external tenant found",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				return &automock.SystemAuthService{}
			},
			appSvc: func() onetimetoken.ApplicationService {
				app := &model.Application{}
				app.IntegrationSystemID = str.Ptr(integrationSystemID)
				app.Tenant = "test-tenant"
				app.BaseEntity = &model.BaseEntity{
					ID: appID,
				}
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(app, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				tenantSvc := &automock.ExternalTenantsService{}
				tenantSvc.On("GetExternalTenant", context.TODO(), "test-tenant").Return("", errors.New("some-error"))
				return tenantSvc
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				return &automock.TokenGenerator{}
			},
			shouldHaveError: true,
			objectID:        appID,
			tokenType:       model.ApplicationReference,
			errorMsg:        "some-error",
			connectorURL:    connectorURL,
			intSystemToAdapterMapping: map[string]string{
				integrationSystemID: "https://my-integration-service.url",
			},
		},
		{
			description: "Generate Application token, should fail on token generating error",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				return &automock.SystemAuthService{}
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(&model.Application{}, nil)
				return appSvc
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return("", errors.New("error generating token"))
				return tokenGenerator
			},
			shouldHaveError:           true,
			objectID:                  appID,
			tokenType:                 model.ApplicationReference,
			errorMsg:                  "error generating token",
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
		},
		{
			description: "Generate Runtime token should succeed",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.RuntimeReference, runtimeID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				return &automock.ApplicationService{}
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return(tokenValue, nil)
				return tokenGenerator
			},
			shouldHaveError:           false,
			objectID:                  runtimeID,
			tokenType:                 model.RuntimeReference,
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
			timeService:               directorTime.NewService(),
		},
		{
			description: "Generate Runtime token should fail on token generating error",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				return &automock.SystemAuthService{}
			},
			appSvc: func() onetimetoken.ApplicationService {
				return &automock.ApplicationService{}
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return("", errors.New("error generating token"))
				return tokenGenerator
			},
			shouldHaveError:           true,
			objectID:                  runtimeID,
			tokenType:                 model.RuntimeReference,
			errorMsg:                  "error generating token",
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
		},
		{
			description: "Generate Runtime token should fail on db error",
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.RuntimeReference, runtimeID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", errors.New("db error"))
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				return &automock.ApplicationService{}
			},
			appConverter: func() onetimetoken.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
			tenantSvc: func() onetimetoken.ExternalTenantsService {
				return &automock.ExternalTenantsService{}
			},
			httpClient: func() onetimetoken.HTTPDoer {
				return &automock.HTTPDoer{}
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				tokenGenerator := &automock.TokenGenerator{}
				tokenGenerator.On("NewToken").Return(tokenValue, nil)
				return tokenGenerator
			},
			shouldHaveError:           true,
			objectID:                  runtimeID,
			tokenType:                 model.RuntimeReference,
			errorMsg:                  "db error",
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
			timeService:               directorTime.NewService(),
		},
	}

	for _, test := range testCases {
		// GIVEN
		systemAuthSvc := test.systemAuthSvc()
		appSvc := test.appSvc()
		appConverter := test.appConverter()
		tenantSvc := test.tenantSvc()
		httpClient := test.httpClient()
		tokenGenerator := test.tokenGenerator()
		timeService := test.timeService

		tokenSvc := onetimetoken.NewTokenService(systemAuthSvc, appSvc, appConverter, tenantSvc, httpClient, tokenGenerator, test.connectorURL, test.intSystemToAdapterMapping, timeService)

		//WHEN
		token, err := tokenSvc.GenerateOneTimeToken(context.TODO(), test.objectID, test.tokenType)

		//THEN
		if test.shouldHaveError {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.errorMsg)
			assert.Empty(t, token)
		} else {
			assert.NoError(t, err)
			if test.tokenType == model.ApplicationReference {
				assert.Equal(t, tokens.ApplicationToken, token.Type)
			} else {
				assert.Equal(t, tokens.RuntimeToken, token.Type)
			}
			assert.Equal(t, fakeToken.Token, token.Token)
			assert.Equal(t, fakeToken.UsedAt, token.UsedAt)
			assert.Equal(t, fakeToken.Used, token.Used)
			if test.intSystemToAdapterMapping == nil {
				assert.Equal(t, fakeToken.ConnectorURL, token.ConnectorURL)
			}
		}
		mock.AssertExpectationsForObjects(t, systemAuthSvc, appSvc, appConverter, tenantSvc, httpClient, tokenGenerator)
	}
}

func TestRegenerateOneTimeToken(t *testing.T) {
	const (
		systemAuthID = "sysAuthID"
		connectorURL = "http://connector.url"
		token        = "tokenValue"
	)

	t.Run("fails when systemAuth cannot be fetched", func(t *testing.T) {
		sysAuthSvc := &automock.SystemAuthService{}
		appSvc := &automock.ApplicationService{}
		appConverter := &automock.ApplicationConverter{}
		extTenantsSvc := &automock.ExternalTenantsService{}
		doer := &automock.HTTPDoer{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(nil, errors.New("error while fetching"))
		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)
		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err, "error while fetching")
	})

	t.Run("fails when new token cannot be generated", func(t *testing.T) {
		sysAuthSvc := &automock.SystemAuthService{}
		appSvc := &automock.ApplicationService{}
		appConverter := &automock.ApplicationConverter{}
		extTenantsSvc := &automock.ExternalTenantsService{}
		doer := &automock.HTTPDoer{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		tokenGenerator.On("NewToken").Return("", errors.New("error while token generating"))

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err, "while generating onetime token error while token generating")
	})

	t.Run("succeeds when systemAuth cannot be updated", func(t *testing.T) {
		sysAuthSvc := &automock.SystemAuthService{}
		appSvc := &automock.ApplicationService{}
		appConverter := &automock.ApplicationConverter{}
		extTenantsSvc := &automock.ExternalTenantsService{}
		doer := &automock.HTTPDoer{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(errors.New("error while updating"))
		tokenGenerator.On("NewToken").Return(token, nil)
		timeService.On("Now").Return(time.Now())
		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err)
	})

	t.Run("succeeds when no errors are thrown", func(t *testing.T) {
		sysAuthSvc := &automock.SystemAuthService{}
		appSvc := &automock.ApplicationService{}
		appConverter := &automock.ApplicationConverter{}
		extTenantsSvc := &automock.ExternalTenantsService{}
		doer := &automock.HTTPDoer{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(nil)
		tokenGenerator.On("NewToken").Return(token, nil)
		now := time.Now()
		timeService.On("Now").Return(now)
		expectedToken := &model.OneTimeToken{
			Token:        token,
			ConnectorURL: connectorURL,
			Type:         tokens.ApplicationToken,
			CreatedAt:    now,
			Used:         false,
			UsedAt:       time.Time{},
		}
		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		assert.Equal(t, expectedToken, &token)
		assert.Nil(t, err)
	})
}
