package onetimetoken_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/client"

	"github.com/kyma-incubator/compass/components/director/pkg/pairing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken/automock"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var URL = `http://localhost:3001/graphql`

func TestTokenService_GetOneTimeTokenForRuntime(t *testing.T) {
	runtimeID := "98cb3b05-0f27-43ea-9249-605ac74a6cf0"
	authID := "90923fe8-91bd-4070-aa31-f2ebb07a0963"

	expectedRuntimeRequest := gcli.NewRequest(fmt.Sprintf(`
		mutation { generateRuntimeToken (authID:"%s")
		  {
			token
		  }
		}`, authID))

	t.Run("Success - token for runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		expectedToken := "token"

		expected := onetimetoken.ConnectorTokenModel{RuntimeToken: onetimetoken.ConnectorToken{Token: expectedToken}}
		cli.On("Run", ctx, expectedRuntimeRequest, &onetimetoken.ConnectorTokenModel{}).
			Run(generateFakeToken(t, expected)).Return(nil).Once()

		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.RuntimeReference, runtimeID, (*model.AuthInput)(nil)).
			Return(authID, nil)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, nil, nil, nil, nil, URL, nil)

		//WHEN
		authToken, err := svc.GenerateOneTimeToken(ctx, runtimeID, model.RuntimeReference)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedToken, authToken.Token)
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})

	t.Run("Error - generating token failed", func(t *testing.T) {
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRuntimeRequest, &onetimetoken.ConnectorTokenModel{}).
			Return(testErr).Once()
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.RuntimeReference, runtimeID, (*model.AuthInput)(nil)).
			Return(authID, nil)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, nil, nil, nil, nil, URL, nil)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, runtimeID, model.RuntimeReference)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})

	t.Run("Error - saving auth failed", func(t *testing.T) {
		ctx := context.TODO()
		testErr := errors.New("test error")
		cli := &automock.GraphQLClient{}
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.RuntimeReference, runtimeID, (*model.AuthInput)(nil)).
			Return("", testErr)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, nil, nil, nil, nil, URL, nil)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, runtimeID, model.RuntimeReference)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
		sysAuthSvc.AssertExpectations(t)
	})

}

func TestTokenService_GetOneTimeTokenForApp(t *testing.T) {
	authID := "77cabc16-9fb8-4338-b252-7b404f2e6487"
	expectedRequest := gcli.NewRequest(fmt.Sprintf(`
		mutation { generateApplicationToken (authID:"%s")
		  {
			token
		  }
		}`, authID))

	applicationID := "5b560bbe-c45b-49e7-847f-20d63b1ac91d"
	integrationSystemID := "fabd8d1e-7a13-485a-8176-e3ca4187bf2c"

	t.Run("Success - token for application not managed by integration system", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		expectedToken := onetimetoken.ConnectorTokenModel{AppToken: onetimetoken.ConnectorToken{Token: "token"}}
		mockGraphqlClient := &automock.GraphQLClient{}
		mockGraphqlClient.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).Run(generateFakeToken(t, expectedToken)).Return(nil).Once()

		mockSysAuthSvc := &automock.SystemAuthService{}
		mockSysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		mockAppService.On("Get", ctx, applicationID).Return(&model.Application{BaseEntity: &model.BaseEntity{ID: applicationID}}, nil)

		svc := onetimetoken.NewTokenService(mockGraphqlClient, mockSysAuthSvc, mockAppService, nil, nil, nil, URL, nil)
		defer mock.AssertExpectationsForObjects(t, mockGraphqlClient, mockAppService, mockSysAuthSvc)
		// WHEN
		actualToken, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "token", actualToken.Token)
	})

	t.Run("Success - token for application with integration system that registered pairing adapter", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		mockSysAuthSvc := &automock.SystemAuthService{}
		mockSysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		givenApplication := model.Application{IntegrationSystemID: &integrationSystemID, Tenant: "internal-tenant", BaseEntity: &model.BaseEntity{ID: applicationID}}
		mockAppService.On("Get", ctx, applicationID).Return(&givenApplication, nil)
		adaptersMapping := map[string]string{integrationSystemID: "https://my-integration-service.url"}

		mockAppConverter := &automock.ApplicationConverter{}
		givenGraphQLApp := graphql.Application{
			IntegrationSystemID: &integrationSystemID,
			BaseEntity: &graphql.BaseEntity{
				ID: givenApplication.ID,
			},
		}
		mockAppConverter.On("ToGraphQL", &givenApplication).Return(&givenGraphQLApp)

		respBody := new(bytes.Buffer)
		respBody.WriteString(`{"token":"external-token"}`)
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
			appIDMatches := appData.Application.ID == givenGraphQLApp.ID
			urlMatches := req.URL.String() == "https://my-integration-service.url"

			return urlMatches && appIDMatches && tenantMatches && clientUserMatches
		})).Return(response, nil)

		mockExtTenants := &automock.ExternalTenantsService{}
		mockExtTenants.On("GetExternalTenant", ctx, "internal-tenant").Return("external-tenant", nil)

		svc := onetimetoken.NewTokenService(nil, mockSysAuthSvc, mockAppService, mockAppConverter, mockExtTenants, mockHttpClient, URL, adaptersMapping)
		defer mock.AssertExpectationsForObjects(t, mockSysAuthSvc, mockAppService, mockHttpClient, mockExtTenants)
		// WHEN
		actualToken, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "external-token", actualToken.Token)
	})

	t.Run("Success - token for application with integration system that registered pairing adapter and client user is provided", func(t *testing.T) {
		const clientUser = "foo"

		// GIVEN
		ctx := context.TODO()
		ctx = client.SaveToContext(ctx, clientUser)

		mockSysAuthSvc := &automock.SystemAuthService{}
		mockSysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		givenApplication := model.Application{IntegrationSystemID: &integrationSystemID, Tenant: "internal-tenant", BaseEntity: &model.BaseEntity{ID: applicationID}}
		mockAppService.On("Get", ctx, applicationID).Return(&givenApplication, nil)
		adaptersMapping := map[string]string{integrationSystemID: "https://my-integration-service.url"}

		mockAppConverter := &automock.ApplicationConverter{}
		givenGraphQLApp := graphql.Application{
			IntegrationSystemID: &integrationSystemID,
			BaseEntity: &graphql.BaseEntity{
				ID: givenApplication.ID,
			},
		}
		mockAppConverter.On("ToGraphQL", &givenApplication).Return(&givenGraphQLApp)

		respBody := new(bytes.Buffer)
		respBody.WriteString(`{"token":"external-token"}`)
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
			clientUserMatches := appData.ClientUser == clientUser
			appIDMatches := appData.Application.ID == givenGraphQLApp.ID
			urlMatches := req.URL.String() == "https://my-integration-service.url"

			return urlMatches && appIDMatches && tenantMatches && clientUserMatches
		})).Return(response, nil)

		mockExtTenants := &automock.ExternalTenantsService{}
		mockExtTenants.On("GetExternalTenant", ctx, "internal-tenant").Return("external-tenant", nil)

		svc := onetimetoken.NewTokenService(nil, mockSysAuthSvc, mockAppService, mockAppConverter, mockExtTenants, mockHttpClient, URL, adaptersMapping)
		defer mock.AssertExpectationsForObjects(t, mockSysAuthSvc, mockAppService, mockHttpClient, mockExtTenants)
		// WHEN
		actualToken, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "external-token", actualToken.Token)
	})

	t.Run("Success - token for application managed by integration system but without registered pairing adapter", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		expectedToken := onetimetoken.ConnectorTokenModel{AppToken: onetimetoken.ConnectorToken{Token: "token"}}
		mockGraphqlClient := &automock.GraphQLClient{}
		mockGraphqlClient.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).Run(generateFakeToken(t, expectedToken)).Return(nil).Once()

		mockSysAuthSvc := &automock.SystemAuthService{}
		mockSysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		mockAppService.On("Get", ctx, applicationID).Return(&model.Application{IntegrationSystemID: &integrationSystemID, BaseEntity: &model.BaseEntity{ID: applicationID}}, nil)

		svc := onetimetoken.NewTokenService(mockGraphqlClient, mockSysAuthSvc, mockAppService, nil, nil, nil, URL, nil)
		defer mock.AssertExpectationsForObjects(t, mockGraphqlClient, mockAppService, mockSysAuthSvc)
		// WHEN
		actualToken, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "token", actualToken.Token)
	})

	t.Run("Error from pairing adapter", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		mockSysAuthSvc := &automock.SystemAuthService{}
		mockSysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		givenApplication := model.Application{IntegrationSystemID: &integrationSystemID, Tenant: "internal-tenant", BaseEntity: &model.BaseEntity{ID: applicationID}}
		mockAppService.On("Get", ctx, applicationID).Return(&givenApplication, nil)
		adaptersMapping := map[string]string{integrationSystemID: "https://my-integration-service.url"}

		mockAppConverter := &automock.ApplicationConverter{}
		givenGraphQLApp := graphql.Application{
			IntegrationSystemID: &integrationSystemID,
			BaseEntity: &graphql.BaseEntity{
				ID: givenApplication.ID,
			},
		}
		mockAppConverter.On("ToGraphQL", &givenApplication).Return(&givenGraphQLApp)

		mockHttpClient := &automock.HTTPDoer{}
		response := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(&bytes.Buffer{}),
		}
		mockHttpClient.On("Do", mock.Anything).Return(response, nil).Times(3)

		mockExtTenants := &automock.ExternalTenantsService{}
		mockExtTenants.On("GetExternalTenant", ctx, "internal-tenant").Return("external-tenant", nil)

		svc := onetimetoken.NewTokenService(nil, mockSysAuthSvc, mockAppService, mockAppConverter, mockExtTenants, mockHttpClient, URL, adaptersMapping)
		defer mock.AssertExpectationsForObjects(t, mockSysAuthSvc, mockAppService, mockHttpClient, mockExtTenants)
		// WHEN
		_, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		require.EqualError(t, err, "while calling adapter [https://my-integration-service.url] for application [5b560bbe-c45b-49e7-847f-20d63b1ac91d] with integration system [fabd8d1e-7a13-485a-8176-e3ca4187bf2c]: All attempts fail:\n#1: wrong status code, got [500], expected [200]\n#2: wrong status code, got [500], expected [200]\n#3: wrong status code, got [500], expected [200]")
	})

	t.Run("Error on getting information about application", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		mockAppService.On("Get", ctx, applicationID).Return(nil, errors.New("some error"))

		svc := onetimetoken.NewTokenService(nil, sysAuthSvc, mockAppService, nil, nil, nil, URL, nil)
		defer mock.AssertExpectationsForObjects(t, sysAuthSvc, mockAppService)
		// WHEN
		_, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		assert.EqualError(t, err, "while getting application [id: 5b560bbe-c45b-49e7-847f-20d63b1ac91d]: some error")

	})

	t.Run("Error on getting information about external tenant", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)

		mockAppService := &automock.ApplicationService{}
		givenApplication := model.Application{IntegrationSystemID: &integrationSystemID, Tenant: "internal-tenant", BaseEntity: &model.BaseEntity{ID: applicationID}}
		mockAppService.On("Get", ctx, applicationID).Return(&givenApplication, nil)
		adaptersMapping := map[string]string{integrationSystemID: "https://my-integration-service.url"}
		mockExtTenants := &automock.ExternalTenantsService{}
		mockExtTenants.On("GetExternalTenant", ctx, "internal-tenant").Return("", errors.New("some error"))

		svc := onetimetoken.NewTokenService(nil, sysAuthSvc, mockAppService, nil, mockExtTenants, nil, URL, adaptersMapping)
		defer mock.AssertExpectationsForObjects(t, sysAuthSvc, mockAppService, mockExtTenants)
		// WHEN
		_, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)
		// THEN
		assert.EqualError(t, err, "while getting external tenant for internal tenant [internal-tenant]: some error")
	})

	t.Run("Error - generating token failed", func(t *testing.T) {
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRequest, &onetimetoken.ConnectorTokenModel{}).
			Return(testErr).Once()
		sysAuthSvc := &automock.SystemAuthService{}
		sysAuthSvc.On("Create", ctx, model.ApplicationReference, applicationID, (*model.AuthInput)(nil)).
			Return(authID, nil)
		appSvc := &automock.ApplicationService{}
		defer mock.AssertExpectationsForObjects(t, cli, sysAuthSvc, appSvc)
		appSvc.On("Get", ctx, applicationID).Return(&model.Application{
			BaseEntity: &model.BaseEntity{ID: applicationID},
		}, nil)

		defer mock.AssertExpectationsForObjects(t, appSvc, cli, sysAuthSvc)
		svc := onetimetoken.NewTokenService(cli, sysAuthSvc, appSvc, nil, nil, nil, URL, nil)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, applicationID, model.ApplicationReference)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
}

func generateFakeToken(t *testing.T, generated onetimetoken.ConnectorTokenModel) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*onetimetoken.ConnectorTokenModel)
		require.True(t, ok)
		*arg = generated
	}
}
