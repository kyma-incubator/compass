package oauth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	namespace  = "compass-system"
	secretName = "compass-broker-registration"
)

func TestOauthClient_GetAuthorizationToken(t *testing.T) {
	t.Run("Should return oauth token", func(t *testing.T) {
		//given
		credentials := credentials{
			clientID:       "12345",
			clientSecret:   "some dark and scary secret",
			tokensEndpoint: "http://hydra:4445",
		}

		token := Token{
			AccessToken: "12345",
			Expiration:  1234,
		}

		client := NewTestClient(func(req *http.Request) *http.Response {
			username, secret, ok := req.BasicAuth()

			if ok && username == credentials.clientID && secret == credentials.clientSecret {
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

		sch := runtime.NewScheme()
		require.NoError(t, v1.AddToScheme(sch))
		cli := fake.NewFakeClientWithScheme(sch, fixCredentialsSecret(credentials))
		oauthClient := NewOauthClient(client, cli, secretName, namespace)

		//when
		responseToken, err := oauthClient.GetAuthorizationToken()
		require.NoError(t, err)
		token.Expiration += time.Now().Unix()

		//then
		assert.Equal(t, token.AccessToken, responseToken.AccessToken)
		assert.Equal(t, token.Expiration, responseToken.Expiration)
	})
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func fixCredentialsSecret(credentials credentials) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			clientIDKey:       []byte(credentials.clientID),
			clientSecretKey:   []byte(credentials.clientSecret),
			tokensEndpointKey: []byte(credentials.tokensEndpoint),
		},
	}
}
