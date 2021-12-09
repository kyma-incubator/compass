package oauth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	contentTypeHeader                = "Content-Type"
	contentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"

	grantTypeFieldName   = "grant_type"
	credentialsGrantType = "client_credentials"

	scopeFieldName = "scope"
	scopes         = "application:read application:write runtime:read runtime:write"

	clientIDKey       = "client_id"
	clientSecretKey   = "client_secret"
	tokensEndpointKey = "tokens_endpoint"
)

type tokenAuthorizationProviderFromSecret struct {
	httpClient        httputils.Client
	k8sClient         client.Client
	waitSecretTimeout time.Duration
	secretName        string
	secretNamespace   string

	token        httputils.Token
	tokenTimeout time.Duration
	lock         sync.RWMutex
}

type Credentials struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
}

func NewAuthorizationProviderFromSecret(config *Config, httpClient httputils.Client, tokenTimeout time.Duration, k8sClientConstructor func(time.Duration) (client.Client, error)) (*tokenAuthorizationProviderFromSecret, error) {
	k8sClient, err := k8sClientConstructor(config.WaitKubeMapperTimeout)
	if err != nil {
		return nil, err
	}

	return &tokenAuthorizationProviderFromSecret{
		httpClient:        httpClient,
		k8sClient:         k8sClient,
		waitSecretTimeout: config.WaitSecretTimeout,
		secretName:        config.SecretName,
		secretNamespace:   config.SecretNamespace,

		token:        httputils.Token{},
		tokenTimeout: tokenTimeout,
		lock:         sync.RWMutex{},
	}, nil
}

func (c *tokenAuthorizationProviderFromSecret) Name() string {
	return "TokenAuthorizationProviderFromSecret"
}

func (c *tokenAuthorizationProviderFromSecret) Matches(ctx context.Context) bool {
	if _, err := getBearerAuthorizationValue(ctx); err != nil {
		log.C(ctx).WithError(err).Warn("while obtaining bearer token")
		return true
	}

	return false
}

func (c *tokenAuthorizationProviderFromSecret) GetAuthorization(ctx context.Context) (string, error) {
	c.lock.RLock()
	isValidToken := !c.token.EmptyOrExpired(c.tokenTimeout)
	c.lock.RUnlock()
	if isValidToken {
		return "Bearer " + c.token.AccessToken, nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.token.EmptyOrExpired(c.tokenTimeout) {
		return "Bearer " + c.token.AccessToken, nil
	}

	log.C(ctx).Debug("Token is invalid, getting a new one...")

	credentials, err := c.extractOAuthClientFromSecret(ctx)
	if err != nil {
		return "", errors.Wrap(err, "while get credentials from secret")
	}

	token, err := GetAuthorizationToken(ctx, c.httpClient, credentials, scopes, nil)
	if err != nil {
		return "", err
	}

	c.token = token
	return "Bearer " + c.token.AccessToken, err
}

func (c *tokenAuthorizationProviderFromSecret) extractOAuthClientFromSecret(ctx context.Context) (Credentials, error) {
	secret := &v1.Secret{}
	err := wait.Poll(time.Second*2, c.waitSecretTimeout, func() (bool, error) {
		err := c.k8sClient.Get(ctx, client.ObjectKey{
			Namespace: c.secretNamespace,
			Name:      c.secretName,
		}, secret)
		// it fails on connection-refused error on first call and it restarts our application.
		if err != nil {
			log.C(ctx).Warnf("secret %s not found", c.secretName)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return Credentials{}, err
	}

	return Credentials{
		ClientID:     string(secret.Data[clientIDKey]),
		ClientSecret: string(secret.Data[clientSecretKey]),
		TokenURL:     string(secret.Data[tokensEndpointKey]),
	}, nil
}

func GetAuthorizationToken(ctx context.Context, httpClient httputils.Client, credentials Credentials, scopes string, additionalHeaders map[string]string) (httputils.Token, error) {
	log.C(ctx).Infof("Getting authorization token from endpoint: %s", credentials.TokenURL)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	if scopes != "" {
		form.Add(scopeFieldName, scopes)
	}
	body := strings.NewReader(form.Encode())
	request, err := http.NewRequest(http.MethodPost, credentials.TokenURL, body)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "Failed to create authorisation token request")
	}

	request.SetBasicAuth(credentials.ClientID, credentials.ClientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)
	if additionalHeaders != nil {
		for headerName, headerValue := range additionalHeaders {
			request.Header.Set(headerName, headerValue)
		}
	}

	log.C(ctx).Infof("Sending request: %+v", request)

	response, err := httpClient.Do(request)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while send request to token endpoint")
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.C(ctx).Warn("Cannot close connection body inside oauth client")
		}
	}()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return httputils.Token{}, errors.Wrapf(err, "while reading token response body from %q", credentials.TokenURL)
	}

	tokenResponse := httputils.Token{}
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while unmarshalling token response body")
	}

	if tokenResponse.AccessToken == "" {
		return httputils.Token{}, errors.New("while fetching token: access token from oauth client is empty")
	}

	log.C(ctx).Info("Successfully unmarshal response oauth token for accessing Director")
	tokenResponse.Expiration += time.Now().Unix()

	return tokenResponse, nil
}

func PrepareK8sClient(duration time.Duration) (client.Client, error) {
	k8sCfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(k8sCfg)
	if err != nil {
		err = wait.Poll(time.Second, duration, func() (bool, error) {
			mapper, err = apiutil.NewDiscoveryRESTMapper(k8sCfg)
			if err != nil {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "while waiting for client mapper")
		}
	}

	cli, err := client.New(k8sCfg, client.Options{Mapper: mapper})
	if err != nil {
		return nil, errors.Wrap(err, "while creating a client")
	}

	return cli, nil
}
