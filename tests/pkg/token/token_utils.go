package token

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/sjson"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

type HydraToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type OauthConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Data         url.Values
}

const (
	RuntimeScopes                                    = "webhook:write runtime:read runtime:write application:read runtime.auths:read bundle.instance_auths:read"
	ApplicationScopes                                = "webhook:write application:read application:write application.auths:read application.webhooks:read bundle.instance_auths:read document.fetch_request:read event_spec.fetch_request:read api_spec.fetch_request:read fetch-request.auth:read"
	IntegrationSystemScopes                          = "application.local_tenant_id:write webhook:write application:read application:write application_template:read application_template:write runtime:read runtime:write integration_system:read label_definition:read label_definition:write automatic_scenario_assignment:read automatic_scenario_assignment:write integration_system.auths:read application_template.webhooks:read internal_visibility:read application.auths:read formation:write"
	IntegrationSystemScopesWithoutInternalVisibility = "application.local_tenant_id:write webhook:write application:read application:write application_template:read application_template:write runtime:read runtime:write integration_system:read label_definition:read label_definition:write automatic_scenario_assignment:read automatic_scenario_assignment:write integration_system.auths:read application_template.webhooks:read"

	contentTypeHeader                = "Content-Type"
	contentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"

	grantTypeFieldName   = "grant_type"
	passwordGrantType    = "password"
	credentialsGrantType = "client_credentials"
	claimsKey            = "claims_key"

	clientIDKey     = "client_id"
	clientSecretKey = "client_secret"
	scopeKey        = "scope"
	userNameKey     = "username"
	passwordKey     = "password"
)

func FetchHydraAccessToken(t *testing.T, encodedCredentials string, tokenURL string, scopes string) (*HydraToken, error) {
	form := url.Values{}
	form.Set(grantTypeFieldName, credentialsGrantType)
	form.Set(scopeKey, scopes)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	require.NoError(t, err)

	req.Header.Add(contentTypeHeader, contentTypeApplicationURLEncoded)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer HttpRequestBodyCloser(t, resp)

	token, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	hydraToken := HydraToken{}
	err = json.Unmarshal(token, &hydraToken)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status code is %d", resp.StatusCode))
	}
	return &hydraToken, nil
}

func FetchHydraAccessTokenBench(b *testing.B, encodedCredentials string, tokenURL string, scopes string) (*HydraToken, error) {
	form := url.Values{}
	form.Set(grantTypeFieldName, credentialsGrantType)
	form.Set(scopeKey, scopes)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	require.NoError(b, err)

	req.Header.Add(contentTypeHeader, contentTypeApplicationURLEncoded)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	require.NoError(b, err)
	defer HttpRequestBodyCloserBench(b, resp)

	token, err := ioutil.ReadAll(resp.Body)
	require.NoError(b, err)

	hydraToken := HydraToken{}
	err = json.Unmarshal(token, &hydraToken)
	require.NoError(b, err)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status code is %d", resp.StatusCode))
	}
	return &hydraToken, nil
}

func GetAccessToken(t *testing.T, oauthCredentialData *graphql.OAuthCredentialData, scopes string) string {
	oauthCredentials := fmt.Sprintf("%s:%s", oauthCredentialData.ClientID, oauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	hydraToken, err := FetchHydraAccessToken(t, encodedCredentials, oauthCredentialData.URL, scopes)
	require.NoError(t, err)
	return hydraToken.AccessToken
}

func GetAccessTokenBench(b *testing.B, oauthCredentialData *graphql.OAuthCredentialData, scopes string) string {
	oauthCredentials := fmt.Sprintf("%s:%s", oauthCredentialData.ClientID, oauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	hydraToken, err := FetchHydraAccessTokenBench(b, encodedCredentials, oauthCredentialData.URL, scopes)
	require.NoError(b, err)
	return hydraToken.AccessToken
}

func HttpRequestBodyCloser(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}

func HttpRequestBodyCloserBench(b *testing.B, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(b, err)
}

func GetClientCredentialsToken(t *testing.T, ctx context.Context, tokenURL, clientID, clientSecret, staticMappingClaimsKey string) string {
	log.C(ctx).Info("Issuing client_credentials token...")
	data := url.Values{}
	data.Add(grantTypeFieldName, credentialsGrantType)
	data.Add(clientIDKey, clientID)
	data.Add(clientSecretKey, clientSecret)
	data.Add(claimsKey, staticMappingClaimsKey)

	token := GetToken(t, ctx, tokenURL, clientID, clientSecret, data)
	log.C(ctx).Info("Successfully issued client_credentials token")

	return token
}

func GetClientCredentialsTokenWithClient(t *testing.T, ctx context.Context, client *http.Client, tokenURL, clientID, clientSecret, staticMappingClaimsKey string) string {
	log.C(ctx).Info("Issuing client_credentials token...")
	data := url.Values{}
	data.Add(grantTypeFieldName, credentialsGrantType)
	data.Add(clientIDKey, clientID)
	data.Add(clientSecretKey, clientSecret)
	data.Add(claimsKey, staticMappingClaimsKey)

	oauthConfig := OauthConfig{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Data:         data,
	}

	token := GetTokenWithClient(t, ctx, client, oauthConfig)
	log.C(ctx).Info("Successfully issued client_credentials token")

	return token
}

func GetUserToken(t *testing.T, ctx context.Context, tokenURL, clientID, clientSecret, username, password, staticMappingClaimsKey string) string {
	log.C(ctx).Info("Issuing user token...")
	data := url.Values{}
	data.Add(grantTypeFieldName, passwordGrantType)
	data.Add(clientIDKey, clientID)
	data.Add(claimsKey, staticMappingClaimsKey)
	data.Add(userNameKey, username)
	data.Add(passwordKey, password)

	token := GetToken(t, ctx, tokenURL, clientID, clientSecret, data)
	log.C(ctx).Info("Successfully issued user token")

	return token
}

func GetToken(t *testing.T, ctx context.Context, tokenURL, clientID, clientSecret string, data url.Values) string {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	oauthConfig := OauthConfig{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Data:         data,
	}

	return GetTokenWithClient(t, ctx, httpClient, oauthConfig)
}

func GetTokenWithClient(t *testing.T, ctx context.Context, client *http.Client, oauthConfig OauthConfig) string {
	req, err := http.NewRequest(http.MethodPost, oauthConfig.TokenURL, bytes.NewBuffer([]byte(oauthConfig.Data.Encode())))
	require.NoError(t, err)

	if oauthConfig.Data.Get(userNameKey) != "" && oauthConfig.Data.Get(passwordKey) != "" {
		req.SetBasicAuth(oauthConfig.ClientID, oauthConfig.ClientSecret)
	}
	req.Header.Add(contentTypeHeader, contentTypeApplicationURLEncoded)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
		}
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("failed to get token: unexpected status code: expected: %d, actual: %d", http.StatusOK, resp.StatusCode))
	require.NotEmpty(t, resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	tkn := gjson.GetBytes(body, "access_token")
	require.True(t, tkn.Exists())
	require.NotEmpty(t, tkn)

	return tkn.String()
}

func FlattenTokenClaims(stdT *testing.T, consumerToken string) string {
	consumerTokenPayload := getTokenPayload(stdT, consumerToken)

	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(consumerTokenPayload), &jsonMap)
	require.NoError(stdT, err)

	fm := flatten(jsonMap)
	claims := "{}"
	for k, v := range fm {
		claims, err = sjson.Set(claims, k, v)
		require.NoError(stdT, err)
	}

	return claims
}

func getTokenPayload(t *testing.T, token string) string {
	// JWT format: <header>.<payload>.<signature>
	tokenParts := strings.Split(token, ".")
	require.Equal(t, 3, len(tokenParts), "invalid token format")
	payload := tokenParts[1]

	b, err := base64.RawURLEncoding.DecodeString(payload)
	require.NoError(t, err)

	return string(b)
}

func flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := flatten(child)
			for nk, nv := range nm {
				o[nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}
