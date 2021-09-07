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
	"github.com/kyma-incubator/compass/components/director/pkg/header"
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
		legacyTokenURL      = connectorURL + "?token=" + tokenValue
		rawEncodedToken     = "eyJ0b2tlbiI6ImFiYyIsImNvbm5lY3RvclVSTCI6ImNvbm5lY3Rvci51cmwifQ=="
		appID               = "4c86b315-c027-467f-a6fc-b184ca0a80f1"
		runtimeID           = "31a607c7-695f-4a31-b2d1-777939f84aac"
		integrationSystemID = "123607c7-695f-4a31-b2d1-777939f84123"

		suggestedTokenHeaderKey = "suggest_token"
	)

	fakeToken := &model.OneTimeToken{
		Used:         false,
		UsedAt:       time.Time{},
		Token:        tokenValue,
		ConnectorURL: connectorURL,
	}

	headers := http.Header{}
	headers.Add(suggestedTokenHeaderKey, "true")
	contextWithEnabledSuggestion := context.WithValue(context.TODO(), header.ContextKey, headers)

	ottConfig := onetimetoken.Config{
		ConnectorURL:          connectorURL,
		LegacyConnectorURL:    connectorURL,
		SuggestTokenHeaderKey: suggestedTokenHeaderKey,
	}

	testCases := []struct {
		description               string
		objectID                  string
		ctx                       context.Context
		connectorURL              string
		shouldHaveError           bool
		errorMsg                  string
		tokenType                 model.SystemAuthReferenceObjectType
		expectedToken             string
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
			ctx:         context.TODO(),
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", context.TODO(), model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", context.TODO(), appID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}}, nil)
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
			expectedToken:             tokenValue,
		},
		{
			description: "Generate Application token, no int system, with suggestion enabled, should succeed",
			ctx:         contextWithEnabledSuggestion,
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", contextWithEnabledSuggestion, model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", contextWithEnabledSuggestion, appID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}}, nil)
				appSvc.On("ListLabels", contextWithEnabledSuggestion, appID).Return(map[string]*model.Label{}, nil)
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
			expectedToken:             rawEncodedToken,
		},
		{
			description: "Generate Application token for legacy application, with suggestion enabled, should succeed",
			ctx:         contextWithEnabledSuggestion,
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", contextWithEnabledSuggestion, model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", contextWithEnabledSuggestion, appID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: appID}}, nil)
				appSvc.On("ListLabels", contextWithEnabledSuggestion, appID).Return(map[string]*model.Label{"legacy": {Value: true}}, nil)
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
			expectedToken:             legacyTokenURL,
		},
		{
			description: "Generate Application token should fail when no such app found",
			ctx:         context.TODO(),
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
			ctx:         context.TODO(),
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
			ctx:         context.TODO(),
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
				mockHTTPClient := &automock.HTTPDoer{}
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(respBody),
				}
				mockHTTPClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
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
				return mockHTTPClient
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				return &automock.TokenGenerator{}
			},
			shouldHaveError: false,
			objectID:        appID,
			tokenType:       model.ApplicationReference,
			expectedToken:   tokenValue,
			connectorURL:    connectorURL,
			intSystemToAdapterMapping: map[string]string{
				integrationSystemID: "https://my-integration-service.url",
			},
			timeService: directorTime.NewService(),
		},
		{
			description: "Generate Application token, with int system, and token suggestion enabled, should succeed",
			ctx:         contextWithEnabledSuggestion,
			systemAuthSvc: func() onetimetoken.SystemAuthService {
				systemAuthSvc := &automock.SystemAuthService{}
				systemAuthSvc.On("Create", contextWithEnabledSuggestion, model.ApplicationReference, appID, mock.MatchedBy(func(authInput *model.AuthInput) bool {
					return authInput.OneTimeToken.Token == tokenValue
				})).Return("", nil)
				return systemAuthSvc
			},
			appSvc: func() onetimetoken.ApplicationService {
				app := &model.Application{BaseEntity: &model.BaseEntity{ID: appID}}
				app.IntegrationSystemID = str.Ptr(integrationSystemID)
				app.Tenant = "test-tenant"
				appSvc := &automock.ApplicationService{}
				appSvc.On("Get", contextWithEnabledSuggestion, appID).Return(app, nil)
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
				tenantSvc.On("GetExternalTenant", contextWithEnabledSuggestion, "test-tenant").Return("external-tenant", nil)
				return tenantSvc
			},
			httpClient: func() onetimetoken.HTTPDoer {
				respBody := new(bytes.Buffer)
				respBody.WriteString(fmt.Sprintf(`{"token":"%s"}`, tokenValue))
				mockHTTPClient := &automock.HTTPDoer{}
				response := &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(respBody),
				}
				mockHTTPClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
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
				return mockHTTPClient
			},
			tokenGenerator: func() onetimetoken.TokenGenerator {
				return &automock.TokenGenerator{}
			},
			shouldHaveError: false,
			objectID:        appID,
			tokenType:       model.ApplicationReference,
			expectedToken:   tokenValue,
			connectorURL:    connectorURL,
			intSystemToAdapterMapping: map[string]string{
				integrationSystemID: "https://my-integration-service.url",
			},
			timeService: directorTime.NewService(),
		},
		{
			description: "Generate Application token, with int system, but no adapters defined should succeed",
			ctx:         context.TODO(),
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
			expectedToken:             tokenValue,
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: map[string]string{},
			timeService:               directorTime.NewService(),
		},
		{
			description: "Generate Application token, with int system, should fail when int system fails",
			ctx:         context.TODO(),
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
				mockHTTPClient := &automock.HTTPDoer{}
				response := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(&bytes.Buffer{}),
				}
				mockHTTPClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
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
				return mockHTTPClient
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
			ctx:         context.TODO(),
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
			ctx:         context.TODO(),
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
			ctx:         context.TODO(),
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
			expectedToken:             tokenValue,
			connectorURL:              connectorURL,
			intSystemToAdapterMapping: nil,
			timeService:               directorTime.NewService(),
		},
		{
			description: "Generate Runtime token should fail on token generating error",
			ctx:         context.TODO(),
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
			ctx:         context.TODO(),
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
		t.Run(test.description, func(t *testing.T) {

			// GIVEN
			systemAuthSvc := test.systemAuthSvc()
			appSvc := test.appSvc()
			appConverter := test.appConverter()
			tenantSvc := test.tenantSvc()
			httpClient := test.httpClient()
			tokenGenerator := test.tokenGenerator()
			timeService := test.timeService

			tokenSvc := onetimetoken.NewTokenService(systemAuthSvc, appSvc, appConverter, tenantSvc, httpClient, tokenGenerator, ottConfig, test.intSystemToAdapterMapping, timeService)

			//WHEN
			token, err := tokenSvc.GenerateOneTimeToken(test.ctx, test.objectID, test.tokenType)

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
				assert.Equal(t, test.expectedToken, token.Token)
				assert.Equal(t, fakeToken.UsedAt, token.UsedAt)
				assert.Equal(t, fakeToken.Used, token.Used)
				if test.intSystemToAdapterMapping == nil {
					assert.Equal(t, fakeToken.ConnectorURL, token.ConnectorURL)
				}
			}
			mock.AssertExpectationsForObjects(t, systemAuthSvc, appSvc, appConverter, tenantSvc, httpClient, tokenGenerator)
		})
	}
}

func TestRegenerateOneTimeToken(t *testing.T) {
	const (
		systemAuthID       = "sysAuthID"
		connectorURL       = "http://connector.url"
		legacyConnectorURL = "http://connector.url"
		token              = "tokenValue"
	)

	ottConfig := onetimetoken.Config{
		ConnectorURL:       connectorURL,
		LegacyConnectorURL: connectorURL,
	}

	t.Run("fails when systemAuth cannot be fetched", func(t *testing.T) {
		// GIVEN
		sysAuthSvc := &automock.SystemAuthService{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(nil, errors.New("error while fetching"))
		defer sysAuthSvc.AssertExpectations(t)

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, &automock.ApplicationService{}, &automock.ApplicationConverter{}, &automock.ExternalTenantsService{},
			&automock.HTTPDoer{}, tokenGenerator, ottConfig, intSystemToAdapterMapping, timeService)

		// WHEN
		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		// THEN
		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error while fetching")
	})

	t.Run("fails when new token cannot be generated", func(t *testing.T) {
		// GIVEN
		sysAuthSvc := &automock.SystemAuthService{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		defer sysAuthSvc.AssertExpectations(t)
		tokenGenerator.On("NewToken").Return("", errors.New("error while token generating"))
		defer tokenGenerator.AssertExpectations(t)

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, &automock.ApplicationService{}, &automock.ApplicationConverter{}, &automock.ExternalTenantsService{},
			&automock.HTTPDoer{}, tokenGenerator, ottConfig, intSystemToAdapterMapping, timeService)

		// WHEN
		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		// THEN
		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while generating onetime token: error while token generating")
	})

	t.Run("fails when systemAuth cannot be updated", func(t *testing.T) {
		// GIVEN
		updateErrMsg := "error while updating"
		sysAuthSvc := &automock.SystemAuthService{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		timeService.On("Now").Return(time.Now())
		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(errors.New(updateErrMsg))
		defer sysAuthSvc.AssertExpectations(t)

		tokenGenerator.On("NewToken").Return(token, nil)
		defer tokenGenerator.AssertExpectations(t)

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, &automock.ApplicationService{}, &automock.ApplicationConverter{}, &automock.ExternalTenantsService{},
			&automock.HTTPDoer{}, tokenGenerator, ottConfig, intSystemToAdapterMapping, timeService)

		// WHEN
		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		// THEN
		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), updateErrMsg)
	})

	t.Run("succeeds when systemAuth has missing 'value' value", func(t *testing.T) {
		// GIVEN
		sysAuthSvc := &automock.SystemAuthService{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)
		now := time.Now()
		timeService.On("Now").Return(now)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(nil)
		defer sysAuthSvc.AssertExpectations(t)

		tokenGenerator.On("NewToken").Return(token, nil)
		defer tokenGenerator.AssertExpectations(t)

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, &automock.ApplicationService{}, &automock.ApplicationConverter{}, &automock.ExternalTenantsService{},
			&automock.HTTPDoer{}, tokenGenerator, ottConfig, intSystemToAdapterMapping, timeService)
		expectedToken := model.OneTimeToken{
			Token:        token,
			ConnectorURL: connectorURL,
			Type:         tokens.ApplicationToken,
			CreatedAt:    now,
			Used:         false,
			UsedAt:       time.Time{},
		}

		// WHEN
		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		// THEN
		assert.Equal(t, expectedToken, token)
		assert.NoError(t, err)
	})

	t.Run("succeeds when no errors are thrown", func(t *testing.T) {
		// GIVEN
		sysAuthSvc := &automock.SystemAuthService{}
		tokenGenerator := &automock.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)
		now := time.Now()
		timeService.On("Now").Return(now)

		sysAuthSvc.On("GetGlobal", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(nil)
		defer sysAuthSvc.AssertExpectations(t)

		tokenGenerator.On("NewToken").Return(token, nil)
		defer tokenGenerator.AssertExpectations(t)
		expectedToken := model.OneTimeToken{
			Token:        token,
			ConnectorURL: connectorURL,
			Type:         tokens.ApplicationToken,
			CreatedAt:    now,
			Used:         false,
			UsedAt:       time.Time{},
		}

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, &automock.ApplicationService{}, &automock.ApplicationConverter{}, &automock.ExternalTenantsService{},
			&automock.HTTPDoer{}, tokenGenerator, ottConfig, intSystemToAdapterMapping, timeService)

		// WHEN
		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		// THEN
		assert.Equal(t, expectedToken, token)
		assert.NoError(t, err)
	})
}
