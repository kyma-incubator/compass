package securehttp_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"

	"golang.org/x/oauth2"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/securehttp"
)

const (
	testClientID        = "client-id"
	testClientIDKey     = "client_id"
	testGrantType       = "client_credentials"
	testGrantTypeKey    = "grant_type"
	testClientSecret    = "client-secret"
	testUser            = "user"
	testPassword        = "pass"
	testToken           = "bGFzbWFnaS1qYXNtYWdpLWtyaXphLWU="
	basicPrefix         = "Basic "
	oauthPrefix         = "Bearer "
	authorizationHeader = "Authorization"
	testScopes          = "scopes"
	testScopesKey       = "scope"
)

func TestCaller_Call(t *testing.T) {
	oauthServerExpectingCredentialsFromHeader := httptest.NewServer(getTestOauthServer(t, requireClientCredentialsFromHeader))
	oauthCredentials := &auth.OAuthCredentials{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		TokenURL:     oauthServerExpectingCredentialsFromHeader.URL,
		Scopes:       testScopes,
	}

	oauthServerExpectingCredentialsFromBody := httptest.NewServer(getTestOauthServer(t, requireClientCredentialsFromBody))
	oauthMtlsCredentials := &auth.OAuthMtlsCredentials{
		ClientID: testClientID,
		TokenURL: oauthServerExpectingCredentialsFromBody.URL,
		Scopes:   testScopes,
	}

	basicCredentials := &auth.BasicCredentials{
		Username: testUser,
		Password: testPassword,
	}

	testCases := []struct {
		Name        string
		Server      *httptest.Server
		Config      securehttp.CallerConfig
		ExpectedErr error
	}{
		{
			Name: "Success for oauth credentials with secret",
			Config: securehttp.CallerConfig{
				Credentials:       oauthCredentials,
				ClientTimeout:     time.Second,
				SkipSSLValidation: true,
			},
			ExpectedErr: nil,
			Server: httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					token := req.Header.Get(authorizationHeader)
					token = strings.TrimPrefix(token, oauthPrefix)
					require.Equal(t, testToken, token)
				}),
			),
		},
		{
			Name: "Success for oauth Mtls",
			Config: securehttp.CallerConfig{
				Credentials:       oauthMtlsCredentials,
				ClientTimeout:     time.Second,
				SkipSSLValidation: true,
			},
			ExpectedErr: nil,
			Server: httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					token := req.Header.Get(authorizationHeader)
					token = strings.TrimPrefix(token, oauthPrefix)
					require.Equal(t, testToken, token)
				}),
			),
		},
		{
			Name: "Success for basic credentials",
			Config: securehttp.CallerConfig{
				Credentials:   basicCredentials,
				ClientTimeout: time.Second,
			},
			ExpectedErr: nil,
			Server: httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					credentials := getBase64EncodedCredentials(t, req, basicPrefix)
					require.Equal(t, testUser+":"+testPassword, credentials)
				}),
			),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			caller, err := securehttp.NewCaller(testCase.Config)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodGet, testCase.Server.URL, nil)
			require.NoError(t, err)

			_, err = caller.Call(request)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func getTestOauthServer(t *testing.T, assertCredentials func(t *testing.T, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		assertCredentials(t, req)

		err := json.NewEncoder(w).Encode(oauth2.Token{
			AccessToken: testToken,
			Expiry:      time.Now().Add(time.Minute),
		})
		require.NoError(t, err)
	}
}

func requireClientCredentialsFromHeader(t *testing.T, req *http.Request) {
	credsStr := getBase64EncodedCredentials(t, req, basicPrefix)
	require.Equal(t, testClientID+":"+testClientSecret, credsStr)
}

func requireClientCredentialsFromBody(t *testing.T, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, testClientIDKey+"="+testClientID+"&"+testGrantTypeKey+"="+testGrantType+"&"+testScopesKey+"="+testScopes, string(requestBody))
}

func getBase64EncodedCredentials(t *testing.T, r *http.Request, prefix string) string {
	creds := r.Header.Get(authorizationHeader)
	creds = strings.TrimPrefix(creds, prefix)

	credsDecoded, err := base64.StdEncoding.DecodeString(creds)
	require.NoError(t, err)
	return string(credsDecoded)
}
