package securehttp_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/securehttp"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	testClientID        = "client-id"
	testClientSecret    = "client-secret"
	testUser            = "user"
	testPassword        = "pass"
	testToken           = "bGFzbWFnaS1qYXNtYWdpLWtyaXphLWU="
	basicPrefix         = "Basic "
	oauthPrefix         = "Bearer "
	authorizationHeader = "Authorization"
)

func TestCaller_Call(t *testing.T) {
	oauthServer := httptest.NewServer(getTestOauthServer(t))
	oauthCredentials := &graphql.OAuthCredentialData{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		URL:          oauthServer.URL,
	}
	basicCredentials := &graphql.BasicCredentialData{
		Username: testUser,
		Password: testPassword,
	}

	testCases := []struct {
		Name        string
		Server      *httptest.Server
		Credentials graphql.CredentialData
		ExpectedErr error
	}{
		{
			Name:        "Success for oauth credentials",
			Credentials: oauthCredentials,
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
			Name:        "Success for basic credentials",
			Credentials: basicCredentials,
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
			caller := securehttp.NewCaller(testCase.Credentials, time.Second)
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

func getTestOauthServer(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		credsStr := getBase64EncodedCredentials(t, req, basicPrefix)
		require.Equal(t, testClientID+":"+testClientSecret, credsStr)

		err := json.NewEncoder(w).Encode(oauth2.Token{
			AccessToken: testToken,
			Expiry:      time.Now().Add(time.Minute),
		})
		require.NoError(t, err)
	}
}

func getBase64EncodedCredentials(t *testing.T, r *http.Request, prefix string) string {
	creds := r.Header.Get(authorizationHeader)
	creds = strings.TrimPrefix(creds, prefix)

	credsDecoded, err := base64.StdEncoding.DecodeString(creds)
	require.NoError(t, err)
	return string(credsDecoded)
}
