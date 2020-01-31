package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	v1 "github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
}

type oauthClient struct {
	httpClient    *http.Client
	secretsClient v1.SecretInterface
	secretName    string
}

func NewOauthClient(client *http.Client, secrets v1.SecretInterface, secretName string) Client {
	return &oauthClient{
		httpClient:    client,
		secretsClient: secrets,
		secretName:    secretName,
	}
}

func (c *oauthClient) GetAuthorizationToken() (Token, error) {
	credentials, err := c.getCredentials()

	if err != nil {
		return Token{}, err
	}

	return c.getAuthorizationToken(credentials)
}

func (c *oauthClient) getCredentials() (credentials, error) {
	secret, err := c.secretsClient.Get(c.secretName, metav1.GetOptions{})

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
	log.Infof("Getting authorisation token for credentials to access Director from endpoint: %s", credentials.tokensEndpoint)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	form.Add(scopeFieldName, scopes)

	request, err := http.NewRequest(http.MethodPost, credentials.tokensEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		log.Errorf("Failed to create authorisation token request")
		return Token{}, err
	}

	now := time.Now().Unix()

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Token{}, err
	}
	defer util.Close(response.Body)

	if response.StatusCode != http.StatusOK {
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			dump = []byte("failed to dump response body")
		}
		return Token{}, fmt.Errorf("Get token call returned unexpected status: %s. Response dump: %s", response.Status, string(dump))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Token{}, fmt.Errorf("Failed to read token response body from '%s': %s", credentials.tokensEndpoint, err.Error())
	}

	tokenResponse := Token{}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token response body: %s", err.Error())
	}

	log.Infof("Successfully unmarshal response oauth token for accessing Director")

	tokenResponse.Expiration += now

	return tokenResponse, nil
}
