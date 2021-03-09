package tests

import (
	"context"
	"crypto/tls"
	"fmt"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/ord-service/pkg"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"strings"
	"testing"
	"time"
)

const (
	acceptHeader = "Accept"
	tenantHeader = "Tenant"
)

type TestData struct {
	tenant    string
	appInput  directorSchema.ApplicationRegisterInput
}

func TestORDAggregator(t *testing.T) {
	appInput := createApp()

	ctx := context.Background()

	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	_, err = pkg.RegisterApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NoError(t, err)

	//defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

	t.Log("Create integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, "test-int-system")
	require.NotEmpty(t, intSys)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)

	intSystemCredentials := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)
	defer pkg.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSystemCredentials.ID)

	unsecuredHttpClient := http.DefaultClient
	unsecuredHttpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	oauthCredentialData, ok := intSystemCredentials.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	conf := &clientcredentials.Config{
		ClientID:     oauthCredentialData.ClientID,
		ClientSecret: oauthCredentialData.ClientSecret,
		TokenURL:     oauthCredentialData.URL,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, unsecuredHttpClient)
	httpClient := conf.Client(ctx)
	httpClient.Timeout = 10 * time.Second

	testData := TestData{
		tenant:    testConfig.DefaultTenant,
		appInput:  appInput,
	}

	t.Run(fmt.Sprintf("Verifying ORD Document for tenant %s to be valid", testData.tenant), func(t *testing.T) {
		err = VerifyORDDocument(testConfig.DefaultCheckInterval, testConfig.DefaultTestTimeout, func() bool {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Bundles...will try again")
				return false
			}

			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, "SAP LoB System 3 Cloud Bundle 1", gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, "SAP LoB System 3 Cloud, our next generation cloud ERP suite designed for in-memory computing, acts as a digital core, connecting your enterprise with people, business networks, the Internet of Things, Big Data, and more.", gjson.Get(respBody, "value.0.description").String())

			return true
		})
		require.NoError(t, err)
	})
}

func createApp() directorSchema.ApplicationRegisterInput {
	return directorSchema.ApplicationRegisterInput{
		Name: "test-app-3",
		ProviderName: ptr.String("testy"),
		Description: ptr.String("my application"),
		Webhooks: []*directorSchema.WebhookInput{
			{
				Type: directorSchema.WebhookTypeOpenResourceDiscovery,
				URL:  &testConfig.ExternalServicesMockBaseURL,
			},
		},
	}
}

func makeRequestWithHeaders(t *testing.T, httpClient *http.Client, url string, headers map[string][]string) string {
	return makeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK)
}

func makeRequestWithHeadersAndStatusExpect(t *testing.T, httpClient *http.Client, url string, headers map[string][]string, expectedHTTPStatus int) string {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	for key, values := range headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := httpClient.Do(request)

	require.NoError(t, err)
	require.Equal(t, expectedHTTPStatus, response.StatusCode)

	parsedURL, err := urlpkg.Parse(url)
	require.NoError(t, err)

	if !strings.Contains(parsedURL.Path, "/specification") {
		formatParam := parsedURL.Query().Get("$format")
		acceptHeader, acceptHeaderProvided := headers[acceptHeader]

		contentType := response.Header.Get("Content-Type")
		if formatParam != "" {
			require.Contains(t, contentType, formatParam)
		} else if acceptHeaderProvided && acceptHeader[0] != "*/*" {
			require.Contains(t, contentType, acceptHeader[0])
		} else {
			require.Contains(t, contentType, "xml")
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}

func VerifyORDDocument(interval time.Duration, timeout time.Duration, conditionalFunc func() bool) error {
	done := time.After(timeout)

	for {
		if conditionalFunc() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for entities to be present in DB")
		default:
			time.Sleep(interval)
		}
	}
}
