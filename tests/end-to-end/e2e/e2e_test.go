package e2e

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/common"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

func init() {
	var err error
	tc, err = NewTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

type hydraToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

var domain string

func TestCompassAuth(t *testing.T) {
	domain = os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)
	ctx := context.Background()

	t.Log("Get Dex id token")
	config, err := idtokenprovider.LoadConfig()
	require.NoError(t, err)

	dexToken, err := idtokenprovider.Authenticate(config.IdProviderConfig)
	require.NoError(t, err)
	t.Log("token:", dexToken)
	t.Log("Create Integration System with Dex id token")
	tc.cli = common.NewAuthorizedGraphQLClient(dexToken)
	intSys := createIntegrationSystem(t, ctx, "integration-system")

	t.Log("Generate Client Credentials for Integration System")
	intSysAuth := generateClientCredentialsForIntegrationSystem(t, ctx, intSys.ID)
	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	hydraToken := fetchHydraAccessToken(t, domain, encodedCredentials, http.StatusOK)

	t.Log("Create an application as Integration System")
	tc.cli = common.NewAuthorizedGraphQLClientWithCustomURL(hydraToken.AccessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))
	appInput := graphql.ApplicationCreateInput{
		Name: "app-created-by-integration-system",
	}
	appByIntSys := createApplicationFromInputWithinTenant(t, ctx, appInput, *tc)

	t.Log("Add API Spec to Application")
	apiInput := graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "new-api-url",
	}
	addApi(t, ctx, apiInput, appByIntSys.ID)

	t.Log("Remove application using Dex id token")
	tc.cli = common.NewAuthorizedGraphQLClient(dexToken)
	deleteApplication(t, ctx, appByIntSys.ID)

	t.Log("Remove Integration System")
	deleteIntegrationSystem(t, ctx, intSys.ID)

	t.Log("Check if token granted for Integration System is invalid")
	tc.cli = common.NewAuthorizedGraphQLClient(hydraToken.AccessToken)
	appInput = graphql.ApplicationCreateInput{
		Name: "app-which-should-be-not-created",
	}
	appInputGQL, err := tc.Graphqlizer.ApplicationCreateInputToGQL(appInput)
	require.NoError(t, err)

	createRequest := fixCreateApplicationRequest(appInputGQL)
	createRequest.Header = http.Header{"Tenant": {"3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"}}
	app := graphql.ApplicationExt{}
	m := resultMapperFor(&app)

	err = tc.withRetryOnTemporaryConnectionProblems(func() error { return tc.cli.Run(ctx, createRequest, &m) })
	require.NoError(t, err)
	require.Empty(t, app.ID)

	t.Log("Check if token can not be fetched with old client credentials")
	fetchHydraAccessToken(t, domain, encodedCredentials, http.StatusUnauthorized)
}

func fetchHydraAccessToken(t *testing.T, domain string, encodedCredentials string, expectedStatusCode int) *hydraToken {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "application:write application:read")

	req, err := http.NewRequest("POST", fmt.Sprintf("https://oauth2.%s/oauth2/token", domain), strings.NewReader(form.Encode()))
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
	require.Equal(t, expectedStatusCode, resp.StatusCode)

	return &hydraToken
}

func httpRequestBodyCloser(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
