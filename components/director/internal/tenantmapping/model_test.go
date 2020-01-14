package tenantmapping

import (
	"net/http"
	"net/textproto"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/stretchr/testify/require"
)

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

		authID, authFlow, err := reqData.GetAuthID()

		require.NoError(t, err)
		require.Equal(t, JWTAuthFlow, authFlow)
		require.Equal(t, authID, username)
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

		authID, authFlow, err := reqData.GetAuthID()

		require.NoError(t, err)
		require.Equal(t, OAuth2Flow, authFlow)
		require.Equal(t, clientID, authID)
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

		authID, authFlow, err := reqData.GetAuthID()

		require.NoError(t, err)
		require.Equal(t, CertificateFlow, authFlow)
		require.Equal(t, clientID, authID)
	})

	t.Run("returns error when no ID string is specified in one of known fields", func(t *testing.T) {
		reqData := ReqData{}

		_, _, err := reqData.GetAuthID()

		require.EqualError(t, err, "unable to find valid auth ID")
	})

	t.Run("returns error when client_id is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					ClientIDKey: []byte{1, 2, 3},
				},
			},
		}

		_, _, err := reqData.GetAuthID()

		require.EqualError(t, err, "while parsing the value for client_id: unable to cast the value to a string type")
	})

	t.Run("returns error when a name is specified in Extra map in a non-string format", func(t *testing.T) {
		reqData := ReqData{
			Body: ReqBody{
				Extra: map[string]interface{}{
					UsernameKey: []byte{1, 2, 3},
				},
			},
		}

		_, _, err := reqData.GetAuthID()

		require.EqualError(t, err, "while parsing the value for name: unable to cast the value to a string type")
	})
}

func TestReqData_GetExternalTenantID(t *testing.T) {
	t.Run("returns tenant ID when it is specified in the Header map", func(t *testing.T) {
		expectedTenant := "f640a8e6-2ce4-450c-bd1c-cba9397f9d79"
		reqData := ReqData{
			Header: http.Header{
				textproto.CanonicalMIMEHeaderKey(TenantKey): []string{expectedTenant},
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
					textproto.CanonicalMIMEHeaderKey(TenantKey): []string{expectedTenant},
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
					TenantKey: expectedTenant,
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
					TenantKey: []byte{1, 2, 3},
				},
			},
		}

		_, err := reqData.GetExternalTenantID()

		require.EqualError(t, err, "while parsing the value for tenant: unable to cast the value to a string type")
	})

	t.Run("returns error when tenant ID is not specified", func(t *testing.T) {
		reqData := ReqData{}

		_, err := reqData.GetExternalTenantID()

		require.Error(t, err)
		require.Implements(t, (*apperrors.KeyDoesNotExist)(nil), err)
		require.EqualError(t, err, "the key (tenant) does not exist in source object")
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

		require.EqualError(t, err, "while parsing the value for scope: unable to cast the value to a string type")
	})

	t.Run("returns error when scopes value is not specified", func(t *testing.T) {
		reqData := ReqData{}

		_, err := reqData.GetScopes()

		require.Error(t, err)
		require.Implements(t, (*apperrors.KeyDoesNotExist)(nil), err)
		require.EqualError(t, err, "the key (scope) does not exist in source object")
	})
}
