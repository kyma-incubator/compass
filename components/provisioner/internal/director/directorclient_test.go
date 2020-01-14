package director

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	oauthmocks "github.com/kyma-incubator/compass/components/provisioner/internal/oauth/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	gcli "github.com/machinebox/graphql"

	"fmt"
	"testing"
	"time"
)

const (
	runtimeTestingID   = "test-runtime-ID-12345"
	runtimeTestingName = "Runtime Test name"
	validTokenValue    = "12345"
	tenantValue        = "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"
	oneTimeToken       = "54321"
	connectorURL       = "https://kyma.cx/connector/graphql"

	expectedRegisterRuntimeQuery = `mutation {
	result: registerRuntime(in: {
		name: "Runtime Test name",
		description: "runtime description",
	}) { id } }`

	expectedDeleteRuntimeQuery = `mutation {
	result: unregisterRuntime(id: "test-runtime-ID-12345") {
		id
}}`

	expectedOneTimeTokenQuery = `mutation {
	result: requestOneTimeTokenForRuntime(id: test-runtime-ID-12345)}`
)

var (
	futureExpirationTime = time.Now().Add(time.Duration(60) * time.Minute).Unix()
	passedExpirationTime = time.Now().Add(time.Duration(60) * time.Minute * -1).Unix()
)

func TestDirectorClient_RuntimeRegistering(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedRegisterRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	inputDescription := "runtime description"

	runtimeInput := &gqlschema.RuntimeInput{
		Name:        runtimeTestingName,
		Description: &inputDescription,
	}

	t.Run("Should register runtime and return new runtime ID when the Director access token is valid", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
			Description: &responseDescription,
		}

		expectedID := runtimeTestingID

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput, tenantValue)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedID, receivedRuntimeID)
	})

	t.Run("Should not register runtime and return an error when the Director access token is empty", func(t *testing.T) {
		// given
		token := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput, tenantValue)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should not register runtime and return an error when the Director access token is expired", func(t *testing.T) {
		// given
		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput, tenantValue)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should not register Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token{}, errors.New("Failed token error"))

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput, tenantValue)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should return error when the result of the call to Director service is nil", func(t *testing.T) {
		// given
		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}, expectedRequest)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput, tenantValue)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should return error when Director fails to register Runtime ", func(t *testing.T) {
		// given
		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}, expectedRequest)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput, tenantValue)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})
}

func TestDirectorClient_RuntimeUnregistering(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedDeleteRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("Should unregister runtime of given ID and return no error when the Director access token is valid", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.NoError(t, err)
	})

	t.Run("Should not unregister runtime and return an error when the Director access token is empty", func(t *testing.T) {
		// given
		emptyToken := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(emptyToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister register runtime and return an error when the Director access token is expired", func(t *testing.T) {
		// given
		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token{}, errors.New("Failed token error"))

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when the result of the call to Director service is nil", func(t *testing.T) {
		// given
		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		// given
		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}, expectedRequest)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when Director fails to delete Runtime", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}, expectedRequest)

		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})

	// unusual and strange case
	t.Run("Should return error when Director returns bad ID after Deleting", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID:          "BadId",
			Name:        runtimeTestingName,
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		validToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(validToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})
}

func TestDirectorClient_GetConnectionToken(t *testing.T) {
	expectedRequest := gcli.NewRequest(expectedOneTimeTokenQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("Should return OneTimeToken when Oauth Token is valid", func(t *testing.T) {
		//given
		expectedResponse := &graphql.OneTimeToken{
			Token:        oneTimeToken,
			ConnectorURL: connectorURL,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*OneTimeTokenResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedOneTimeToken, err := configClient.GetConnectionToken(runtimeTestingID, tenantValue)

		//then
		require.NoError(t, err)
		require.NotEmpty(t, receivedOneTimeToken)
		assert.Equal(t, oneTimeToken, receivedOneTimeToken.Token)
		assert.Equal(t, connectorURL, receivedOneTimeToken.ConnectorURL)
	})

	t.Run("Should return error when Oauth Token is empty", func(t *testing.T) {
		//given
		token := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		receivedOneTimeToken, err := configClient.GetConnectionToken(runtimeTestingID, tenantValue)

		//then
		require.Error(t, err)
		require.Empty(t, receivedOneTimeToken)
	})

	t.Run("Should return error when Oauth Token is expired", func(t *testing.T) {
		//given
		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		receivedOneTimeToken, err := configClient.GetConnectionToken(runtimeTestingID, tenantValue)

		//then
		require.Error(t, err)
		require.Empty(t, receivedOneTimeToken)
	})

	t.Run("Should return error when Director call returns nil reponse", func(t *testing.T) {
		//given
		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*OneTimeTokenResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
		}, expectedRequest)

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedOneTimeToken, err := configClient.GetConnectionToken(runtimeTestingID, tenantValue)

		//then
		require.Error(t, err)
		require.Empty(t, receivedOneTimeToken)
	})
}
