package oathkeeper

import (
	"context"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

var ctx = context.TODO()

func TestAuthFlow(t *testing.T) {
	t.Run("IsCertFlow returns true when AuthFlow equals to Certificate", func(t *testing.T) {
		authFlow := CertificateFlow

		require.True(t, authFlow.IsCertFlow())
		require.False(t, authFlow.IsJWTFlow())
		require.False(t, authFlow.IsOAuth2Flow())
	})

	t.Run("IsOAuth2Flow returns true when AuthFlow equals to OAuth2", func(t *testing.T) {
		authFlow := OAuth2Flow

		require.True(t, authFlow.IsOAuth2Flow())
		require.False(t, authFlow.IsCertFlow())
		require.False(t, authFlow.IsJWTFlow())
	})

	t.Run("IsJWTFlow returns true when AuthFlow equals to JWT", func(t *testing.T) {
		authFlow := JWTAuthFlow

		require.True(t, authFlow.IsJWTFlow())
		require.False(t, authFlow.IsOAuth2Flow())
		require.False(t, authFlow.IsCertFlow())
	})
}

func TestReqData_GetTenantID(t *testing.T) {
	t.Run("returns tenant ID when it is specified in the Header map", func(t *testing.T) {
		expectedTenant := "f640a8e6-2ce4-450c-bd1c-cba9397f9d79"
		reqData := ReqData{
			Header: http.Header{
				textproto.CanonicalMIMEHeaderKey(ExternalTenantKey): []string{expectedTenant},
			},
		}

		tenant, err := reqData.GetExternalTenantID()

		require.NoError(t, err)
		require.Equal(t, expectedTenant, tenant)
	})

	t.Run("returns tenant ID when it is specified in the Header map of Body", func(t *testing.T) {
		expectedTenant := "f640a8e6-2ce4-450c-bd1c-cba9397f9d79"
		reqData := ReqData{
			Body: ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(ExternalTenantKey): []string{expectedTenant},
				},
			},
		}

		tenant, err := reqData.GetExternalTenantID()

		require.NoError(t, err)
		require.Equal(t, expectedTenant, tenant)
	})

	t.Run("returns tenant ID when it is specified in the Extra map", func(t *testing.T) {
		expectedTenant := "f640a8e6-2ce4-450c-bd1c-cba9397f9d79"
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ExternalTenantKey: expectedTenant,
				},
			},
		}

		tenant, err := reqData.GetExternalTenantID()

		require.NoError(t, err)
		require.Equal(t, expectedTenant, tenant)
	})

	t.Run("returns error when tenant ID is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ExternalTenantKey: []byte{1, 2, 3},
				},
			},
		}

		_, err := reqData.GetExternalTenantID()

		require.EqualError(t, err, "while parsing the value for key=tenant: Internal Server Error: unable to cast the value to a string type")
	})

	t.Run("returns error when tenant ID is not specified", func(t *testing.T) {
		reqData := ReqData{}

		_, err := reqData.GetExternalTenantID()

		require.Error(t, err)
		require.EqualError(t, err, "the key does not exist in the source object [key=tenant]")
	})
}

func TestReqData_GetScopes(t *testing.T) {
	t.Run("returns scopes string when it is specified in the Extra map", func(t *testing.T) {
		expectedScopes := "applications:write runtimes:write"
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ScopesKey: expectedScopes,
				},
			},
		}

		scopes, err := reqData.GetScopes()

		require.NoError(t, err)
		require.Equal(t, expectedScopes, scopes)
	})

	t.Run("returns error when scopes value is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ScopesKey: []byte{1, 2, 3},
				},
			},
		}

		_, err := reqData.GetScopes()

		require.EqualError(t, err, "while parsing the value for scope: Internal Server Error: unable to cast the value to a string type")
	})

	t.Run("returns error when scopes value is not specified", func(t *testing.T) {
		reqData := ReqData{}

		_, err := reqData.GetScopes()

		require.Error(t, err)
		require.EqualError(t, err, "the key does not exist in the source object [key=scope]")
	})
}

func TestReqData_GetUserScopes(t *testing.T) {
	const scopePrefix = "test-compass@b12345."

	t.Run("returns scopes string array when it is specified in the Extra map", func(t *testing.T) {
		scopes := []interface{}{scopePrefix + "applications:write", scopePrefix + "runtimes:write"}
		expectedScopes := []interface{}{"applications:write", "runtimes:write"}
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ScopesKey: scopes,
				},
			},
		}

		actualScopes, err := reqData.GetUserScopes(scopePrefix)

		require.NoError(t, err)
		require.ElementsMatch(t, expectedScopes, actualScopes)
	})

	t.Run("returns empty scopes string array when it is not specified in the Extra map", func(t *testing.T) {
		reqData := ReqData{}

		actualScopes, err := reqData.GetUserScopes(scopePrefix)

		require.NoError(t, err)
		require.Empty(t, actualScopes)
	})

	t.Run("returns empty scopes string array when it is specified in the Extra map but is not array", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ScopesKey: 1,
				},
			},
		}

		actualScopes, err := reqData.GetUserScopes(scopePrefix)

		require.NoError(t, err)
		require.Empty(t, actualScopes)
	})

	t.Run("returns error when scopes are specified in the Extra map but some elements/scopes are not strings", func(t *testing.T) {
		scopes := []interface{}{"applications:write", "runtimes:write", 24}
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ScopesKey: scopes,
				},
			},
		}

		_, err := reqData.GetUserScopes(scopePrefix)

		require.EqualError(t, err, "while parsing the value for scope: Internal Server Error: unable to cast the value to a string type")
	})
}

func TestReqData_GetUserGroups(t *testing.T) {
	t.Run("returns groups when it is specified in the Extra map", func(t *testing.T) {
		expectedGroups := []string{
			"developers",
			"admin",
			"tenantID=123",
		}

		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					GroupsKey: []interface{}{
						"developers",
						"admin",
						"tenantID=123",
					},
				},
			},
		}

		groups := reqData.GetUserGroups()

		require.Equal(t, expectedGroups, groups)
	})

	t.Run("returns empty array when groups value is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					GroupsKey: []byte{1, 2, 3},
				},
			},
		}

		groups := reqData.GetUserGroups()

		require.Equal(t, []string{}, groups)
	})
}

func TestReqData_ExtractCoordinates(t *testing.T) {
	coordinates := authenticator.Coordinates{
		Name:  "test",
		Index: 1,
	}

	tests := []struct {
		name                 string
		body                 ReqBody
		expectedCoordinates  authenticator.Coordinates
		expectedSuccess      bool
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name: "success",
			body: ReqBody{
				Extra: map[string]interface{}{
					authenticator.CoordinatesKey: coordinates,
				},
			},
			expectedCoordinates: coordinates,
			expectedSuccess:     true,
			expectError:         false,
		},
		{
			name:                "fail when key does not exist",
			body:                ReqBody{},
			expectedCoordinates: authenticator.Coordinates{},
			expectedSuccess:     false,
			expectError:         false,
		},
		{
			name: "fail when error occurs while marshaling",
			body: ReqBody{
				Extra: map[string]interface{}{
					authenticator.CoordinatesKey: "fail",
				},
			},
			expectedCoordinates:  authenticator.Coordinates{},
			expectedSuccess:      true,
			expectError:          true,
			expectedErrorMessage: "while unmarshaling authenticator coordinates",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewReqData(context.TODO(), tt.body, nil)
			coords, success, err := d.ExtractCoordinates()

			if tt.expectError {
				assert.Contains(t, err.Error(), tt.expectedErrorMessage)
			}

			assert.Equal(t, tt.expectedCoordinates, coords)
			assert.Equal(t, tt.expectedSuccess, success)
		})
	}
}
