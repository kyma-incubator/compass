package e2e

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

type hydraToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func TestCompassAuth(t *testing.T) {
	domain := os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)
	tenant := os.Getenv("DEFAULT_TENANT")
	require.NotEmpty(t, tenant)
	ctx := context.Background()

	t.Log("Get Dex id_token")
	config, err := idtokenprovider.LoadConfig()
	require.NoError(t, err)

	dexToken, err := idtokenprovider.Authenticate(config.IdProviderConfig)
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Create Integration System with Dex id token")
	intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")

	t.Log("Generate Client Credentials for Integration System")
	intSysAuth := generateClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	hydraToken, err := fetchHydraAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL)
	require.NoError(t, err)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(hydraToken.AccessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

	t.Log("Create an application as Integration System")
	appInput := graphql.ApplicationRegisterInput{
		Name:                "app-created-by-integration-system",
		ProviderName:        "compass",
		IntegrationSystemID: &intSys.ID,
	}
	appByIntSys := registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenant, appInput)
	require.NotEmpty(t, appByIntSys.ID)

	t.Log("Add API Spec to Application")
	apiInput := graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://kyma-project.io",
	}
	addAPIWithinTenant(t, ctx, oauthGraphQLClient, tenant, apiInput, appByIntSys.ID)
	t.Log("Try removing Integration System")
	unregisterIntegrationSystemWithErr(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	t.Log("Check if SystemAuths are still present in the db")
	auths := getSystemAuthsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
	require.NotEmpty(t, auths)
	require.NotNil(t, auths[0])
	assert.Equal(t, intSysAuth, *auths[0])

	t.Log("Remove application to check if the oAuth token is still valid")
	deleteApplication(t, ctx, oauthGraphQLClient, tenant, appByIntSys.ID)

	t.Log("Remove Integration System")
	unregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

	t.Log("Check if token granted for Integration System is invalid")
	appByIntSys = registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenant, appInput)
	require.Empty(t, appByIntSys.ID)

	t.Log("Check if token can not be fetched with old client credentials")
	_, err = fetchHydraAccessToken(t, encodedCredentials, intSysOauthCredentialData.URL)
	require.Error(t, err)
	assert.Equal(t, "response status code is 401", err.Error())
}

func fetchHydraAccessToken(t *testing.T, encodedCredentials string, tokenURL string) (*hydraToken, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "application:write application:read")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	require.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer httpRequestBodyCloser(t, resp)

	token, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	hydraToken := hydraToken{}
	err = json.Unmarshal(token, &hydraToken)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status code is %d", resp.StatusCode))
	}
	return &hydraToken, nil
}

func httpRequestBodyCloser(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
