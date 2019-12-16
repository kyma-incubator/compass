package director

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	"github.com/stretchr/testify/assert"

	oauthmocks "github.com/kyma-incubator/compass/components/provisioner/internal/oauth/mocks"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"

	"fmt"
	"testing"
	"time"
)

const (

	runtimeID = "test-runtime-ID-12345"

	expectedCreateRuntimeQuery =
	`mutation {
	result: createRuntime(in: {
		name: "Runtime Test name",
		description: "runtime description",
	})`

	validTokenValue = "12345"

	expectedDeleteRuntimeQuery = `mutation {
	result: deleteRuntime(id: test-runtime-ID-12345)`

)

var (
	futureExpirationTime = time.Now().Add(time.Duration(60) * time.Minute).Unix()
	passedExpirationTime = time.Now().Add(time.Duration(60) * time.Minute * -1).Unix()
)

func TestDirectorClient_RuntimeRegistering(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedCreateRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))

	inputDescription := "runtime description"

	runtimeInput := &gqlschema.RuntimeInput{
		Name :  "Runtime Test name",
		Description : &inputDescription,
	}

	t.Run("Should register runtime and return new runtime ID when the Director access token is valid", func(t *testing.T) {
		// given

		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		expectedID := "test-runtime-ID-12345"

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
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.NoError(t, err)
		assert.Equal(t, expectedID, receivedRuntimeID)
	})

	t.Run("Should not register runtime and return an error when the Director access token is empty", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		token := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(token, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should not register runtime and return an error when the Director access token is expired", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})


	t.Run("Should not register Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token {}, errors.New("Failed token error"))

		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

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

		// given
		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = nil
		}, expectedRequest)


		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should return error when Director fails to register Runtime ", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
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
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})
}


func TestDirectorClient_UnregisterRuntime(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedDeleteRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))

	t.Run("Should unregister runtime of given ID and return no error when the Director access token is valid", func(t *testing.T) {
		// given

		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: runtimeID,
			Name : "runtime name",
			Description: &responseDescription,
		}

		expectedID := "test-runtime-ID-12345"

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
		err := configClient.DeleteRuntime(expectedID)

		// then
		assert.NoError(t, err)
	})

	t.Run("Should not unregister runtime and return an error when the Director access token is empty", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: runtimeID,
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		emptyToken := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(emptyToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister register runtime and return an error when the Director access token is expired", func(t *testing.T) {

		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: runtimeID,
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeID)

		// then
		assert.Error(t, err)
	})


	t.Run("Should not unregister Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token {}, errors.New("Failed token error"))

		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*DeleteRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient)

		// when
		err := configClient.DeleteRuntime(runtimeID)

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
		err := configClient.DeleteRuntime(runtimeID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when Director fails to delete Runtime ", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID: "test-runtime-ID-12345",
			Name : "runtime name",
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
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
		err := configClient.DeleteRuntime(runtimeID)

		// then
		assert.Error(t, err)
	})
}

func TestConfigClient_SetURLsLabels(t *testing.T) {

	//runtimeURLsConfig := RuntimeURLsConfig{
	//	EventsURL:  "https://gateway.kyma.local",
	//	ConsoleURL: "https://console.kyma.local",
	//}
	//
	//expectedSetEventsURLRequest := gcli.NewRequest(expectedSetEventsURLLabelQuery)
	//expectedSetEventsURLRequest.Header.Set(TenantHeader, tenant)
	//expectedSetConsoleURLRequest := gcli.NewRequest(expectedSetConsoleURLLabelQuery)
	//expectedSetConsoleURLRequest.Header.Set(TenantHeader, tenant)
	//
	//newSetExpectedLabelFunc := func(expectedResponses []*graphql.Label) func(t *testing.T, r interface{}) {
	//	var responseIndex = 0
	//
	//	return func(t *testing.T, r interface{}) {
	//		cfg, ok := r.(*SetRuntimeLabelResponse)
	//		require.True(t, ok)
	//		assert.Empty(t, cfg.Result)
	//		cfg.Result = expectedResponses[responseIndex]
	//		responseIndex++
	//	}
	//}

	t.Run("should set URLs as labels", func(t *testing.T) {

		//expectedResponses := []*graphql.Label{
		//	{
		//		Key:   eventsURLLabelKey,
		//		Value: runtimeURLsConfig.EventsURL,
		//	},
		//	{
		//		Key:   consoleURLLabelKey,
		//		Value: runtimeURLsConfig.ConsoleURL,
		//	},
		//}
		//
		//gqlClient := gql.NewQueryAssertClient(t, false, newSetExpectedLabelFunc(expectedResponses), expectedSetEventsURLRequest, expectedSetConsoleURLRequest)
		//
		//configClient := NewConfigurationClient(gqlClient, runtimeConfig)
		//
		//// when
		//labels, err := configClient.SetURLsLabels(runtimeURLsConfig)
		//
		//// then
		//require.NoError(t, err)
		//assert.Equal(t, runtimeURLsConfig.EventsURL, labels[eventsURLLabelKey])
		//assert.Equal(t, runtimeURLsConfig.ConsoleURL, labels[consoleURLLabelKey])
	})

	t.Run("should return error if setting Console URL as label returned nil response", func(t *testing.T) {
		//expectedResponses := []*graphql.Label{
		//	{
		//		Key:   eventsURLLabelKey,
		//		Value: runtimeURLsConfig.EventsURL,
		//	},
		//	nil,
		//}
		//
		//gqlClient := gql.NewQueryAssertClient(t, false, newSetExpectedLabelFunc(expectedResponses), expectedSetEventsURLRequest, expectedSetConsoleURLRequest)
		//
		//configClient := NewConfigurationClient(gqlClient, runtimeConfig)
		//
		//// when
		//labels, err := configClient.SetURLsLabels(runtimeURLsConfig)
		//
		//// then
		//require.Error(t, err)
		//assert.Nil(t, labels)
	})

	t.Run("should return error if setting Console URL as label returned nil response", func(t *testing.T) {
	/*	expectedResponses := []*graphql.Label{nil, nil}
		//
		gqlClient := gql.NewQueryAssertClient(t, false, newSetExpectedLabelFunc(expectedResponses), expectedSetEventsURLRequest, expectedSetConsoleURLRequest)
		//
		directorClient := NewDirectorClient(gqlClient)
		//
		//// when
		runtimeID, err := directorClient.CreateRuntime(input)
		//
		//// then
		require.Error(t, err)
		assert.Nil(t, labels)*/
	})

	t.Run("should return error if failed to set labels", func(t *testing.T) {
	//	gqlClient := gql.NewQueryAssertClient(t, true, func(t *testing.T, r interface{}) {
	//		cfg, ok := r.(*SetRuntimeLabelResponse)
	//		require.True(t, ok)
	//		assert.Empty(t, cfg.Result)
	//	}, expectedSetEventsURLRequest, expectedSetConsoleURLRequest)
	//
	//	configClient := NewConfigurationClient(gqlClient, runtimeConfig)
	//
	//	// when
	//	labels, err := configClient.SetURLsLabels(runtimeURLsConfig)
	//
	//	// then
	//	require.Error(t, err)
	//	assert.Nil(t, labels)
	})

}
