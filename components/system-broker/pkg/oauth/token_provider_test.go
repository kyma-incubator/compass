package oauth_test

import (
	"bytes"
	"context"
	"encoding/json"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func TestTokenProviderFromValue_GetAuthorizationToken(t *testing.T) {
	tokenValue := "token"

	testProvider := oauth.NewTokenProviderFromValue(tokenValue)
	token, err := testProvider.GetAuthorizationToken(context.TODO())

	require.NoError(t, err)
	require.Equal(t, tokenValue, token.AccessToken)
}

func TestTokenProviderFromSecretTestSuite(t *testing.T) {
	suite.Run(t, new(TokenProviderFromSecretTestSuite))
}

type TokenProviderFromSecretTestSuite struct {
	suite.Suite

	clientID      string
	clientSecret  string
	tokenEndpoint string
	accessToken   string

	config    *oauth.Config
	k8sClient client.Client
}

func (suite *TokenProviderFromSecretTestSuite) SetupTest() {
	suite.clientID = "admin_client_id"
	suite.clientSecret = "admin_client_secret"
	suite.tokenEndpoint = "http://localhost:8080/oauth/token"
	suite.accessToken = "access_token"

	suite.config = oauth.DefaultConfig()
	suite.k8sClient = fake.NewFakeClientWithScheme(scheme.Scheme, &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.config.SecretName,
			Namespace: suite.config.SecretNamespace,
		},
		Data: map[string][]byte{
			"client_id":       []byte(suite.clientID),
			"client_secret":   []byte(suite.clientSecret),
			"tokens_endpoint": []byte(suite.tokenEndpoint),
		},
		Type: "Opaque",
	})
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken() {
	httpClient := &MockClient{
		DoFunc: suite.validResponseDoFunc(),
	}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, httpClient, suite.k8sClient)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.accessToken, token.AccessToken)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken_ShouldReturnErrorIfEmptyBody() {
	httpClient := &MockClient{}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, httpClient, suite.k8sClient)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken_ShouldReturnErrorIfReturnedTokenIsEmpty() {
	suite.accessToken = ""
	httpClient := &MockClient{
		DoFunc: suite.validResponseDoFunc(),
	}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, httpClient, suite.k8sClient)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken_ShouldReturnErrorWhenSecretNotFound() {
	suite.k8sClient = fake.NewFakeClientWithScheme(scheme.Scheme)
	httpClient := &MockClient{}
	suite.config.WaitSecretTimeout = 5 * time.Second

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, httpClient, suite.k8sClient)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *TokenProviderFromSecretTestSuite) validResponseDoFunc() func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		suite.Require().Equal(suite.tokenEndpoint, req.URL.String())

		username, password, ok := req.BasicAuth()
		suite.Require().True(ok, "Expected client credentials from secret to be provided as basic header")
		suite.Require().Equal(suite.clientID, username)
		suite.Require().Equal(suite.clientSecret, password)

		token := httputils.Token{
			AccessToken: suite.accessToken,
			Expiration:  9999,
		}
		jsonBytes, err := json.Marshal(token)
		suite.Require().NoError(err)

		return &http.Response{
			Body: ioutil.NopCloser(bytes.NewBuffer(jsonBytes)),
		}, nil
	}
}

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{
		Body: ioutil.NopCloser(&bytes.Buffer{}),
	}, nil
}
