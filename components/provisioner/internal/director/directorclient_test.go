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
	result: requestOneTimeTokenForRuntime(id: "test-runtime-ID-12345") {
		token connectorURL
}}`

	expectedGetRuntimeQuery = `query {
    result: runtime(id: "test-runtime-ID-12345") {
         id name description labels
}}`

	expectedUpdateMutation = `mutation {
    result: updateRuntime(id: "test-runtime-ID-12345" in: {
		name: "Runtime Test name",
		labels: {label1:"something",label2:"something2",},
		statusCondition: CONNECTED,
	}) {
		id
}}`
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

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

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

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

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

		gqlClient := gql.NewQueryAssertClient(t, true, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

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

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

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
		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when Director fails to delete Runtime", func(t *testing.T) {
		// given
		gqlClient := gql.NewQueryAssertClient(t, true, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

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

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

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
		expectedResponse := &graphql.OneTimeTokenForRuntimeExt{
			OneTimeTokenForRuntime: graphql.OneTimeTokenForRuntime{
				TokenWithURL: graphql.TokenWithURL{
					Token:        oneTimeToken,
					ConnectorURL: connectorURL,
				},
			},
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*OneTimeTokenResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

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
		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*OneTimeTokenResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
		})

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

func TestDirectorClient_GetRuntime(t *testing.T) {
	expectedRequest := gcli.NewRequest(expectedGetRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("should return Runtime", func(t *testing.T) {
		//given
		expectedResponse := &graphql.RuntimeExt{
			Runtime: graphql.Runtime{
				ID:   runtimeTestingID,
				Name: runtimeTestingName,
			},
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		runtime, err := configClient.GetRuntime(runtimeTestingID, tenantValue)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedResponse.Name, runtime.Name)
		assert.Equal(t, expectedResponse.ID, runtime.ID)
	})

	t.Run("should return error when access token is empty", func(t *testing.T) {
		// given
		emptyToken := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(emptyToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		runtime, err := configClient.GetRuntime(runtimeTestingID, tenantValue)

		// then
		assert.Error(t, err)
		assert.Empty(t, runtime)
	})

	t.Run("should return error when Director returns nil response", func(t *testing.T) {
		//given
		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		runtime, err := configClient.GetRuntime(runtimeTestingID, tenantValue)

		//then
		require.Error(t, err)
		assert.Empty(t, runtime)
	})

	t.Run("should return error when Director fails to get Runtime", func(t *testing.T) {
		//given
		gqlClient := gql.NewQueryAssertClient(t, true, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		runtime, err := configClient.GetRuntime(runtimeTestingID, tenantValue)

		//then
		require.Error(t, err)
		assert.Empty(t, runtime)
	})
}

func TestDirectorClient_UpdateRuntime(t *testing.T) {
	expectedRequest := gcli.NewRequest(expectedUpdateMutation)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("should update Runtime", func(t *testing.T) {
		//given
		conditionConnectoed := gqlschema.RuntimeStatusConditionConnected
		runtimeInput := &gqlschema.RuntimeInput{
			Name: runtimeTestingName,
			Labels: &gqlschema.Labels{
				"label1": "something",
				"label2": "something2",
			},
			StatusCondition: &conditionConnectoed,
		}

		expectedResponse := &graphql.Runtime{
			ID:   runtimeTestingID,
			Name: runtimeTestingName,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*UpdateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		})

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		err := configClient.UpdateRuntime(runtimeTestingID, runtimeInput, tenantValue)

		//then
		require.NoError(t, err)
	})

	t.Run("should return error when access token is empty", func(t *testing.T) {
		// given
		emptyToken := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		conditionConnectoed := gqlschema.RuntimeStatusConditionConnected
		runtimeInput := &gqlschema.RuntimeInput{
			Name: runtimeTestingName,
			Labels: &gqlschema.Labels{
				"label1": "something",
				"label2": "something2",
			},
			StatusCondition: &conditionConnectoed,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(emptyToken, nil)

		configClient := NewDirectorClient(nil, mockedOAuthClient)

		// when
		err := configClient.UpdateRuntime(runtimeTestingID, runtimeInput, tenantValue)

		// then
		assert.Error(t, err)
	})

	t.Run("should return error when Director returns nil response", func(t *testing.T) {
		//given
		conditionConnectoed := gqlschema.RuntimeStatusConditionConnected
		runtimeInput := &gqlschema.RuntimeInput{
			Name: runtimeTestingName,
			Labels: &gqlschema.Labels{
				"label1": "something",
				"label2": "something2",
			},
			StatusCondition: &conditionConnectoed,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedRequest}, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*UpdateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		})

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		err := configClient.UpdateRuntime(runtimeTestingID, runtimeInput, tenantValue)

		//then
		require.Error(t, err)
	})
}

func TestDirectorClient_SetRuntimeStatusCondition(t *testing.T) {
	expectedFirstRequest := gcli.NewRequest(expectedGetRuntimeQuery)
	expectedFirstRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedFirstRequest.Header.Set(TenantHeader, tenantValue)

	statusCondition := gqlschema.RuntimeStatusConditionConnected

	expectedSecondRequest := gcli.NewRequest(expectedUpdateMutation)
	expectedSecondRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedSecondRequest.Header.Set(TenantHeader, tenantValue)

	t.Run("should set runtime status condition", func(t *testing.T) {
		expectedGetResponse := &graphql.RuntimeExt{
			Runtime: graphql.Runtime{
				ID:   runtimeTestingID,
				Name: runtimeTestingName,
			},
			Labels: graphql.Labels{
				"label1": "something",
				"label2": "something2",
			},
		}

		getFunction := func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedGetResponse
		}

		expectedUpdateResponse := &graphql.Runtime{
			ID:   runtimeTestingID,
			Name: runtimeTestingName,
		}

		updateFunction := func(t *testing.T, r interface{}) {
			cfg, ok := r.(*UpdateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedUpdateResponse
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedFirstRequest, expectedSecondRequest}, getFunction, updateFunction)

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		err := configClient.SetRuntimeStatusCondition(runtimeTestingID, statusCondition, tenantValue)

		//then
		require.NoError(t, err)
	})

	t.Run("should return error when GetRuntime returns error", func(t *testing.T) {
		getFunction := func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}

		expectedUpdateResponse := &graphql.Runtime{
			ID:   runtimeTestingID,
			Name: runtimeTestingName,
		}

		updateFunction := func(t *testing.T, r interface{}) {
			cfg, ok := r.(*UpdateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedUpdateResponse
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedFirstRequest, expectedSecondRequest}, getFunction, updateFunction)

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		err := configClient.SetRuntimeStatusCondition(runtimeTestingID, statusCondition, tenantValue)

		//then
		require.Error(t, err)
	})

	t.Run("should return error when UpdateRuntime returns error", func(t *testing.T) {
		expectedGetResponse := &graphql.RuntimeExt{
			Runtime: graphql.Runtime{
				ID:   runtimeTestingID,
				Name: runtimeTestingName,
			},
			Labels: graphql.Labels{
				"label1": "something",
				"label2": "something2",
			},
		}

		getFunction := func(t *testing.T, r interface{}) {
			cfg, ok := r.(*GetRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedGetResponse
		}

		updateFunction := func(t *testing.T, r interface{}) {
			cfg, ok := r.(*UpdateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}

		gqlClient := gql.NewQueryAssertClient(t, false, []*gcli.Request{expectedFirstRequest, expectedSecondRequest}, getFunction, updateFunction)

		token := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		//when
		err := configClient.SetRuntimeStatusCondition(runtimeTestingID, statusCondition, tenantValue)

		//then
		require.Error(t, err)
	})
}
