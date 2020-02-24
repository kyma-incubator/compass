package director

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"

	machineGraphql "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_setToken(t *testing.T) {
	for name, tc := range map[string]struct {
		token       string
		expire      int64
		expectedErr error
	}{
		"OauthClient returns token": {
			token:       "abcd1234",
			expire:      10,
			expectedErr: nil,
		},
		"OauthClient returns error": {
			token:       "",
			expire:      0,
			expectedErr: fmt.Errorf("some token error"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Given
			oc := &mocks.OauthClient{}
			qc := &mocks.GraphQLClient{}

			oc.On("GetAuthorizationToken").Return(oauth.Token{tc.token, tc.expire}, tc.expectedErr)
			client := NewDirectorClient(oc, qc)

			// When
			tokenErr := client.setToken()

			// Then
			if tc.expectedErr == nil {
				assert.NoError(t, tokenErr)
				assert.Equal(t, tc.token, client.token.AccessToken)
				assert.Equal(t, tc.expire, client.token.Expiration)
			} else {
				assert.Error(t, tokenErr)
				assert.Equal(t, tc.token, client.token.AccessToken)
				assert.Equal(t, int64(0), client.token.Expiration)
			}
		})
	}
}

func TestClient_GetConsoleURL(t *testing.T) {
	var (
		runtimeID   = "620f2303-f084-4956-8594-b351fbff124d"
		accountID   = "32f2e45c-74dc-4bb8-b03f-7cb6a44c1fd9"
		expectedURL = "http://example.com"
		token       = oauth.Token{
			AccessToken: "1234xyza",
			Expiration:  time.Now().Unix() + 999,
		}
		oc = &mocks.OauthClient{}
	)

	t.Run("url returned successfully", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Run(func(args mock.Arguments) {
			arg, ok := args.Get(2).(*successResponse)
			if !ok {
				return
			}
			arg.Result = graphql.RuntimeExt{
				Runtime: graphql.Runtime{
					Status: &graphql.RuntimeStatus{
						Condition: graphql.RuntimeStatusConditionReady,
					},
				},
				Labels: map[string]interface{}{
					consoleURLLabelKey: expectedURL,
				},
			}
		}).Return(nil)

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.NoError(t, tokenErr)
		assert.False(t, IsTemporaryError(tokenErr))
		assert.Equal(t, expectedURL, URL)
	})

	t.Run("response from director is empty", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Return(nil)

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.Error(t, tokenErr)
		assert.True(t, IsTemporaryError(tokenErr))
		assert.Equal(t, "", URL)
	})

	t.Run("response from director is in failed state", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Run(func(args mock.Arguments) {
			arg, ok := args.Get(2).(*successResponse)
			if !ok {
				return
			}
			arg.Result = graphql.RuntimeExt{
				Runtime: graphql.Runtime{
					Status: &graphql.RuntimeStatus{
						Condition: graphql.RuntimeStatusConditionFailed,
					},
				},
				Labels: map[string]interface{}{
					consoleURLLabelKey: "",
				},
			}
		}).Return(nil)

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.Error(t, tokenErr)
		assert.False(t, IsTemporaryError(tokenErr))
		assert.Equal(t, "", URL)
	})

	t.Run("response from director has no proper labels", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Run(func(args mock.Arguments) {
			arg, ok := args.Get(2).(*successResponse)
			if !ok {
				return
			}
			arg.Result = graphql.RuntimeExt{
				Runtime: graphql.Runtime{
					Status: &graphql.RuntimeStatus{
						Condition: graphql.RuntimeStatusConditionReady,
					},
				},
				Labels: map[string]interface{}{
					"wrongURLLabel": expectedURL,
				},
			}
		}).Return(nil)

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.Error(t, tokenErr)
		assert.True(t, IsTemporaryError(tokenErr))
		assert.Equal(t, "", URL)
	})

	t.Run("response from director has label with wrong type", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Run(func(args mock.Arguments) {
			arg, ok := args.Get(2).(*successResponse)
			if !ok {
				return
			}
			arg.Result = graphql.RuntimeExt{
				Runtime: graphql.Runtime{
					Status: &graphql.RuntimeStatus{
						Condition: graphql.RuntimeStatusConditionReady,
					},
				},
				Labels: map[string]interface{}{
					consoleURLLabelKey: 42,
				},
			}
		}).Return(nil)

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.Error(t, tokenErr)
		assert.False(t, IsTemporaryError(tokenErr))
		assert.Equal(t, "", URL)
	})

	t.Run("response from director has wrong URL value", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Run(func(args mock.Arguments) {
			arg, ok := args.Get(2).(*successResponse)
			if !ok {
				return
			}
			arg.Result = graphql.RuntimeExt{
				Runtime: graphql.Runtime{
					Status: &graphql.RuntimeStatus{
						Condition: graphql.RuntimeStatusConditionReady,
					},
				},
				Labels: map[string]interface{}{
					consoleURLLabelKey: "wrong-URL",
				},
			}
		}).Return(nil)

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.Error(t, tokenErr)
		assert.False(t, IsTemporaryError(tokenErr))
		assert.Equal(t, "", URL)
	})

	t.Run("client graphQL returns error", func(t *testing.T) {
		// Given
		qc := &mocks.GraphQLClient{}
		oc = &mocks.OauthClient{}

		client := NewDirectorClient(oc, qc)
		client.token = token

		// #create request
		request := createGraphQLRequest(client, accountID, runtimeID)

		// #mock on GetAuthorizationToken for OauthClient
		oc.On("GetAuthorizationToken").Return(token, nil)

		// #mock on Run method for grapQL client
		qc.On("Run", context.Background(), request, mock.AnythingOfType("*director.successResponse")).Times(3).Return(fmt.Errorf("director error"))

		// When
		URL, tokenErr := client.GetConsoleURL(accountID, runtimeID)

		// Then
		assert.Error(t, tokenErr)
		assert.True(t, IsTemporaryError(tokenErr))
		assert.Equal(t, "", URL)
	})
}

func createGraphQLRequest(client *Client, accountID, runtimeID string) *machineGraphql.Request {
	query := client.queryProvider.Runtime(runtimeID)
	request := machineGraphql.NewRequest(query)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.token.AccessToken))
	request.Header.Add(accountIDKey, accountID)

	return request
}
