package oathkeeper

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

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

func TestReqData_GetAuthID(t *testing.T) {
	t.Run("returns ID string and JWTAuthFlow when a name is specified in the Extra map of request body", func(t *testing.T) {
		username := "some-username"
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					UsernameKey: username,
				},
			},
		}

		authDetails, err := reqData.GetAuthID(ctx)

		require.NoError(t, err)
		require.Equal(t, JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
	})

	t.Run("returns ID string and OAuth2Flow when a client_id is specified in the Extra map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ClientIDKey: clientID,
				},
			},
		}

		authDetails, err := reqData.GetAuthID(ctx)

		require.NoError(t, err)
		require.Equal(t, OAuth2Flow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		reqData := ReqData{
			Body: ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(ClientIDCertKey): []string{clientID},
				},
			},
		}

		authDetails, err := reqData.GetAuthID(ctx)

		require.NoError(t, err)
		require.Equal(t, CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns error when no ID string is specified in one of known fields", func(t *testing.T) {
		reqData := ReqData{}

		_, err := reqData.GetAuthID(ctx)

		require.EqualError(t, err, apperrors.NewInternalError("unable to find valid auth ID").Error())
	})

	t.Run("returns error when client_id is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ClientIDKey: []byte{1, 2, 3},
				},
			},
		}

		_, err := reqData.GetAuthID(ctx)

		require.EqualError(t, err, "while parsing the value for client_id: Internal Server Error: unable to cast the value to a string type")
	})

	t.Run("returns error when a name is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					UsernameKey: []byte{1, 2, 3},
				},
			},
		}

		_, err := reqData.GetAuthID(ctx)

		require.EqualError(t, err, "while parsing the value for name: Internal Server Error: unable to cast the value to a string type")
	})
}

func TestReqData_GetAuthIDWithAuthenticators(t *testing.T) {
	var (
		uniqueAttributeKey   string
		uniqueAttributeValue string
		identityAttributeKey string
		username             string
		authenticatorName    string
		scopePrefix          string
		domainURL            string
		reqData              ReqData
		authn                []authenticator.Config
	)
	setup := func() {
		uniqueAttributeKey = "uniqueAttribute"
		uniqueAttributeValue = "uniqueAttributeValue"
		identityAttributeKey = "identity"
		authenticatorName = "auth1"
		scopePrefix = "prefix"
		domainURL = "domain.com"
		username = "some-username"
		reqData = ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					authenticator.CoordinatesKey: authenticator.Coordinates{
						Name:  authenticatorName,
						Index: 0,
					},
					uniqueAttributeKey:   uniqueAttributeValue,
					identityAttributeKey: username,
				},
			},
		}

		authn = []authenticator.Config{
			{
				Name: authenticatorName,
				TrustedIssuers: []authenticator.TrustedIssuer{
					{
						DomainURL:   domainURL,
						ScopePrefix: scopePrefix,
					},
				},
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
			},
		}
	}

	t.Run("returns ID string and JWTAuthFlow when authenticator identity is specified in the Extra map of request body", func(t *testing.T) {
		setup()
		authDetails, err := reqData.GetAuthIDWithAuthenticators(ctx, authn)

		require.NoError(t, err)
		require.Equal(t, JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
		require.Equal(t, scopePrefix, authDetails.ScopePrefix)
	})

	t.Run("returns ID string and JWTAuthFlow when authenticator identity and also default username attribute is specified in the Extra map of request body", func(t *testing.T) {
		setup()
		identityUsername := "some-identity"
		reqData.Body.Extra[UsernameKey] = username
		reqData.Body.Extra[identityAttributeKey] = identityUsername

		authDetails, err := reqData.GetAuthIDWithAuthenticators(ctx, authn)

		require.NoError(t, err)
		require.Equal(t, JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, identityUsername, authDetails.AuthID)
		require.Equal(t, scopePrefix, authDetails.ScopePrefix)
	})

	t.Run("returns ID string and JWTAuthFlow when username attribute is specified in the Extra map of request body and no authenticators match", func(t *testing.T) {
		setup()
		reqData.Body.Extra[authenticator.CoordinatesKey] = authenticator.Coordinates{
			Name: "unknown",
		}
		reqData.Body.Extra[UsernameKey] = username

		authn := []authenticator.Config{
			{
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
			},
		}

		authDetails, err := reqData.GetAuthIDWithAuthenticators(ctx, authn)

		require.NoError(t, err)
		require.Equal(t, JWTAuthFlow, authDetails.AuthFlow)
		require.Equal(t, username, authDetails.AuthID)
	})

	t.Run("returns error during JWTAuthFlow when authenticator unique attribute is present but identity attribute is not specified in the Extra map of request body", func(t *testing.T) {
		setup()
		delete(reqData.Body.Extra, identityAttributeKey)

		_, err := reqData.GetAuthIDWithAuthenticators(ctx, authn)

		require.EqualError(t, err, apperrors.NewInvalidDataError(fmt.Sprintf("missing identity attribute from %q authenticator token", authn[0].Name)).Error())
	})

	t.Run("returns error when no ID string is specified in one of known fields", func(t *testing.T) {
		setup()
		identityAttributeKey = "identity"
		reqData = ReqData{}

		authn = []authenticator.Config{
			{
				Attributes: authenticator.Attributes{
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
			},
		}

		_, err := reqData.GetAuthIDWithAuthenticators(ctx, authn)

		require.EqualError(t, err, apperrors.NewInternalError("unable to find valid auth ID").Error())
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
