package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	namespace  = "compass-system"
	secretName = "compass-provisioner-registration"
)

func TestOauthClient_GetAuthorizationToken(t *testing.T) {
	t.Run("Should return oauth token", func(t *testing.T) {
		//given
		credentials := credentials{
			clientID:     "12345",
			clientSecret: "some dark and scary secret",
		}

		token := Token{
			AccessToken: "12345",
			Expiration:  1234,
		}

		hydraUrl := "http://hydra:4445"

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

		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets, credentials)
		defer deleteSecret(t, secrets)

		oauthClient := NewOauthClient(hydraUrl, client, secrets, secretName)

		//when
		responseToken, err := oauthClient.GetAuthorizationToken()
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

func createFakeCredentialsSecret(t *testing.T, secrets core.SecretInterface, credentials credentials) {
	encodedClientID := base64.StdEncoding.EncodeToString([]byte(credentials.clientID))
	encodedClientSecret := base64.StdEncoding.EncodeToString([]byte(credentials.clientSecret))

	secret := &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			clientIDKey: []byte(encodedClientID),
			clientSecretKey: []byte(encodedClientSecret),
		},
	}

	_, err := secrets.Create(secret)

	require.NoError(t, err)
}

func deleteSecret(t *testing.T, secrets core.SecretInterface) {
	err := secrets.Delete(secretName, &meta.DeleteOptions{})
	require.NoError(t, err)
}
