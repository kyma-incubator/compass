package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
	WaitForCredentials() error
}

type oauthClient struct {
	httpClient      *http.Client
	k8sClient       client.Client
	secretName      string
	secretNamespace string
}

func NewOauthClient(httpClient *http.Client, k8sClient client.Client, secretName, secretNamespace string) Client {
	return &oauthClient{
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		secretName:      secretName,
		secretNamespace: secretNamespace,
	}
}

func (c *oauthClient) GetAuthorizationToken() (Token, error) {
	credentials, err := c.getCredentials()
	if err != nil {
		return Token{}, errors.Wrap(err, "while get credentials from secret")
	}

	return c.getAuthorizationToken(credentials)
}

func (c *oauthClient) WaitForCredentials() error {
	err := wait.Poll(time.Second, time.Minute*3, func() (bool, error) {
		secret := &v1.Secret{}
		err := c.k8sClient.Get(context.Background(), client.ObjectKey{
			Namespace: c.secretNamespace,
			Name:      c.secretName,
		}, secret)
		// it fails on connection-refused error on first call and it restarts our application.
		if err != nil {
			log.Warnf("secret %s not found with error: %v", c.secretName, err)
			return false, nil
		}
		return true, nil
	})

	return errors.Wrapf(err, "while waiting for secret %s", c.secretName)
}

func (c *oauthClient) getCredentials() (credentials, error) {
	secret := &v1.Secret{}
	err := c.k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: c.secretNamespace,
		Name:      c.secretName,
	}, secret)
	if err != nil {
		return credentials{}, err
	}

	return credentials{
		clientID:       string(secret.Data[clientIDKey]),
		clientSecret:   string(secret.Data[clientSecretKey]),
		tokensEndpoint: string(secret.Data[tokensEndpointKey]),
	}, nil
}

func (c *oauthClient) getAuthorizationToken(credentials credentials) (Token, error) {
	log.Infof("Getting authorization token for credentials to access Director from endpoint: %s", credentials.tokensEndpoint)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	form.Add(scopeFieldName, scopes)

	request, err := http.NewRequest(http.MethodPost, credentials.tokensEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, errors.Wrap(err, "while creating authorisation token request")
	}

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Token{}, errors.Wrap(err, "while send request to token endpoint")
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Warnf("Cannot close connection body inside oauth client")
		}
	}()

	if response.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("while calling to token endpoint: unexpected status code, %d, %s", response.StatusCode, response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Token{}, errors.Wrapf(err, "while reading token response body from %q", credentials.tokensEndpoint)
	}

	tokenResponse := Token{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return Token{}, errors.Wrap(err, "while unmarshalling token response body")
	}

	if tokenResponse.AccessToken == "" {
		return Token{}, errors.New("while fetching token: access token from oauth client is empty")
	}

	log.Errorf("Successfully unmarshal response oauth token for accessing Director")
	tokenResponse.Expiration += time.Now().Unix()

	return tokenResponse, nil
}
