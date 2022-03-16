package http_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/pkg/errors"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/stretchr/testify/require"
)

const testURL = "http://localhost:8000"

var (
	expectedResp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       nil,
	}
	testErr = errors.New("test error")
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	if resp := f(req); resp == nil {
		return nil, testErr
	}
	return f(req), nil
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestRequestWithoutCredentials_Success(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		require.Empty(t, req.Header.Get("Authorization"))
		return expectedResp
	})

	resp, err := httputil.GetRequestWithoutCredentials(client, testURL)
	require.NoError(t, err)
	require.Equal(t, resp, expectedResp)
}

func TestRequestWithoutCredentials_FailedRequest(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		return nil
	})

	_, err := httputil.GetRequestWithoutCredentials(client, testURL)
	require.ErrorIs(t, err, testErr)
}

func TestRequestWithCredentials_SuccessWithBasicAuth(t *testing.T) {
	user, pass := "username", "password"
	client := newTestClient(func(req *http.Request) *http.Response {
		username, password, exists := req.BasicAuth()
		require.True(t, exists)
		require.Equal(t, username, user)
		require.Equal(t, password, pass)
		return expectedResp
	})

	resp, err := httputil.GetRequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: user,
				Password: pass,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, resp, expectedResp)
}

func TestRequestWithCredentials_FailedWithBasicAuth(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		return nil
	})

	_, err := httputil.GetRequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "user",
				Password: "pass",
			},
		},
	})
	require.ErrorIs(t, err, testErr)
}

func TestRequestWithCredentials_SuccessWithOAuth(t *testing.T) {
	clientID := "client-id"
	clientSecret := "client-secret"
	tokenURL := "http://oauth-server.com/token"
	testTkn := "test-tkn"

	client := newTestClient(func(req *http.Request) *http.Response {
		authorizationHeader := req.Header.Get("Authorization")
		if strings.Contains(req.URL.String(), testURL) {
			require.Equal(t, authorizationHeader, "Bearer "+testTkn)

			return expectedResp
		}

		auth := strings.TrimLeft(authorizationHeader, "Basic ")
		data, err := base64.StdEncoding.DecodeString(auth)
		require.NoError(t, err)
		clientCreds := strings.Split(string(data), ":")
		require.Equal(t, clientID, clientCreds[0])
		require.Equal(t, clientSecret, clientCreds[1])
		require.Equal(t, tokenURL, req.URL.String())

		data, err = json.Marshal(struct {
			AccessToken string `json:"access_token"`
		}{
			AccessToken: testTkn,
		})
		require.NoError(t, err)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
		}
	})

	resp, err := httputil.GetRequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Oauth: &model.OAuthCredentialData{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				URL:          tokenURL,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, resp, expectedResp)
}

func TestRequestWithCredentials_FailedWithOAuthDueToInvalidCredentials(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
		}
	})

	_, err := httputil.GetRequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Oauth: &model.OAuthCredentialData{},
		},
	})
	require.Error(t, err)
}
