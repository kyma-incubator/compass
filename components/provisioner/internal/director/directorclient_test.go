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

	expectedRegisterRuntimeQuery = `mutation {
	result: registerRuntime(in: {
		name: "Runtime Test name",
		description: "runtime description",
	}) { id } }`

	expectedDeleteRuntimeQuery = `mutation {
	result: deleteRuntime(id: test-runtime-ID-12345)}`
)

var (
	futureExpirationTime = time.Now().Add(time.Duration(60) * time.Minute).Unix()
	passedExpirationTime = time.Now().Add(time.Duration(60) * time.Minute * -1).Unix()
)

func TestDirectorClient_RuntimeRegistering(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedRegisterRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantKey, tenantValue)

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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

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
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

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
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})

	t.Run("Should not register Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token{}, errors.New("Failed token error"))

		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
			Description: &responseDescription,
		}

		gqlClient := gql.NewQueryAssertClient(t, false, func(t *testing.T, r interface{}) {
			cfg, ok := r.(*CreateRuntimeResponse)
			require.True(t, ok)
			assert.Empty(t, cfg.Result)
			cfg.Result = expectedResponse
		}, expectedRequest)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

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
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		receivedRuntimeID, err := configClient.CreateRuntime(runtimeInput)

		// then
		assert.Error(t, err)
		assert.Empty(t, receivedRuntimeID)
	})
}

func TestDirectorClient_RuntimeUnregistering(t *testing.T) {

	expectedRequest := gcli.NewRequest(expectedDeleteRuntimeQuery)
	expectedRequest.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", validTokenValue))
	expectedRequest.Header.Set(TenantKey, tenantValue)

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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

		// then
		assert.NoError(t, err)
	})

	t.Run("Should not unregister runtime and return an error when the Director access token is empty", func(t *testing.T) {
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

		emptyToken := oauth.Token{
			AccessToken: "",
			Expiration:  futureExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(emptyToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister register runtime and return an error when the Director access token is expired", func(t *testing.T) {

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

		expiredToken := oauth.Token{
			AccessToken: validTokenValue,
			Expiration:  passedExpirationTime,
		}

		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(expiredToken, nil)

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should not unregister Runtime and return error when the client fails to get an access token for Director", func(t *testing.T) {
		// given
		mockedOAuthClient := &oauthmocks.Client{}
		mockedOAuthClient.On("GetAuthorizationToken").Return(oauth.Token{}, errors.New("Failed token error"))

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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

		// then
		assert.Error(t, err)
	})

	t.Run("Should return error when Director fails to delete Runtime", func(t *testing.T) {
		// given
		responseDescription := "runtime description"
		expectedResponse := &graphql.Runtime{
			ID:          runtimeTestingID,
			Name:        runtimeTestingName,
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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

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

		configClient := NewDirectorClient(gqlClient, mockedOAuthClient, tenantValue)

		// when
		err := configClient.DeleteRuntime(runtimeTestingID)

		// then
		assert.Error(t, err)
	})
}
