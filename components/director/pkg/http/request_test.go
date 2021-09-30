package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/stretchr/testify/require"
)

const testURL = "http://localhost:8000"

var expectedResp = &http.Response{
	StatusCode: http.StatusOK,
	Body:       nil,
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestRequestWithoutCredentials_Success(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		return expectedResp
	})

	resp, err := httputil.RequestWithoutCredentials(client, testURL)
	require.NoError(t, err)
	require.Equal(t, resp, expectedResp)
}

func TestRequestWithoutCredentials_FailedRequest(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		return nil
	})

	_, err := httputil.RequestWithoutCredentials(client, testURL)
	require.Error(t, err)
}

func TestRequestWithCredentials_SuccessWithBasicAuth(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		return expectedResp
	})

	resp, err := httputil.RequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "user",
				Password: "pass",
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

	_, err := httputil.RequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "user",
				Password: "pass",
			},
		},
	})
	require.Error(t, err)
}

func TestRequestWithCredentials_SuccessWithOAuth(t *testing.T) {
	client := newTestClient(func(req *http.Request) *http.Response {
		if strings.Contains(req.URL.String(), testURL) {
			return expectedResp
		}

		data, err := json.Marshal(struct {
			AccessToken string `json:"access_token"`
		}{
			AccessToken: "test-tkn",
		})
		require.NoError(t, err)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
		}
	})

	resp, err := httputil.RequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Oauth: &model.OAuthCredentialData{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				URL:          "url",
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

	_, err := httputil.RequestWithCredentials(context.Background(), client, testURL, &model.Auth{
		Credential: model.CredentialData{
			Oauth: &model.OAuthCredentialData{},
		},
	})
	require.Error(t, err)
}
