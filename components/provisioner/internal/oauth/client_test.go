package oauth

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestOauthClient_GetAuthorizationToken(t *testing.T) {
	t.Run("Should return oauth token", func(t *testing.T) {
		//given
		credentials := Credentials{
			ClientID:     "12345",
			ClientSecret: "some dark and scary secret",
		}

		token := TokenResponse{
			AccessToken: "12345",
			Expiration: 1234,
		}

		hydraUrl := "http://hydra:4445"

		client := NewTestClient(func(req *http.Request) *http.Response {
			username, secret, ok := req.BasicAuth()

			if ok && username == credentials.ClientID && secret == credentials.ClientSecret {
				jsonToken, err := json.Marshal(&token)

				require.NoError(t, err)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(jsonToken)),
				}
			}
			return &http.Response{
				StatusCode: http.StatusForbidden,
			}
		})

		oauthClient := NewOauthClient(hydraUrl, client)

		//when
		responseToken, err := oauthClient.GetAuthorizationToken(credentials)
		require.NoError(t, err)

		//then
		assert.Equal(t, token.AccessToken, responseToken.AccessToken)
		assert.Equal(t, token.Expiration, responseToken.Expiration)
	})
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
