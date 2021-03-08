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
	apisMap   map[string]directorSchema.APIDefinitionInput
	eventsMap map[string]directorSchema.EventDefinitionInput
}

func TestORDAggregator(t *testing.T) {
	appInput := createApp("tenant1")

	apisMap := make(map[string]directorSchema.APIDefinitionInput, 0)
	for _, apiDefinition := range appInput.Bundles[0].APIDefinitions {
		apisMap[apiDefinition.Name] = *apiDefinition
	}

	eventsMap := make(map[string]directorSchema.EventDefinitionInput, 0)
	for _, eventDefinition := range appInput.Bundles[0].EventDefinitions {
		eventsMap[eventDefinition.Name] = *eventDefinition
	}

	ctx := context.Background()

	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	app, err := pkg.RegisterApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NoError(t, err)

	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

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

	for _, resource := range []string{"vendors", "tombstones", "products"} { // This tests assert integrity between ORD Service JPA model and our Database model
		t.Run(fmt.Sprintf("Requesting %s returns empty", resource), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/%s?$format=json", testConfig.ORDServiceURL, resource), map[string][]string{tenantHeader: {testConfig.DefaultTenant}})
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})
	}

	testData := TestData{
		tenant:    testConfig.DefaultTenant,
		appInput:  appInput,
		apisMap:   apisMap,
		eventsMap: eventsMap,
	}

	t.Run(fmt.Sprintf("Requesting Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
		err = WaitForFunction(testConfig.DefaultCheckInterval, testConfig.DefaultTestTimeout, func() bool {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				//t.Log("Missing Bundles...will try again")
				return false
			}

			require.Equal(t, len(testData.appInput.Bundles), len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())

			return true
		})
		require.NoError(t, err)
	})
}

func createApp(suffix string) directorSchema.ApplicationRegisterInput {
	return generateAppInputForDifferentTenants(directorSchema.ApplicationRegisterInput{
		Name:        "test-app",
		Description: ptr.String("my application"),
		Bundles: []*directorSchema.BundleCreateInput{
			{
				Name:        "foo-bndl",
				Description: ptr.String("foo-descr"),
				APIDefinitions: []*directorSchema.APIDefinitionInput{
					{
						Name:        "comments-v1",
						Description: ptr.String("api for adding comments"),
						TargetURL:   "http://mywordpress.com/comments",
						Group:       ptr.String("comments"),
						Version:     pkg.FixDepracatedVersion(),
						Spec: &directorSchema.APISpecInput{
							Type:   directorSchema.APISpecTypeOpenAPI,
							Format: directorSchema.SpecFormatYaml,
							Data:   ptr.CLOB(`{"openapi":"3.0.2"}`),
						},
					},
					{
						Name:        "reviews-v1",
						Description: ptr.String("api for adding reviews"),
						TargetURL:   "http://mywordpress.com/reviews",
						Version:     pkg.FixActiveVersion(),
						Spec: &directorSchema.APISpecInput{
							Type:   directorSchema.APISpecTypeOdata,
							Format: directorSchema.SpecFormatJSON,
							Data:   ptr.CLOB(`{"openapi":"3.0.1"}`),
						},
					},
					{
						Name:        "xml",
						Description: ptr.String("xml api"),
						Version:     pkg.FixDecomissionedVersion(),
						TargetURL:   "http://mywordpress.com/xml",
						Spec: &directorSchema.APISpecInput{
							Type:   directorSchema.APISpecTypeOdata,
							Format: directorSchema.SpecFormatXML,
							Data:   ptr.CLOB("odata"),
						},
					},
				},
				EventDefinitions: []*directorSchema.EventDefinitionInput{
					{
						Name:        "comments-v1",
						Description: ptr.String("comments events"),
						Version:     pkg.FixDepracatedVersion(),
						Group:       ptr.String("comments"),
						Spec: &directorSchema.EventSpecInput{
							Type:   directorSchema.EventSpecTypeAsyncAPI,
							Format: directorSchema.SpecFormatYaml,
							Data:   ptr.CLOB(`{"asyncapi":"1.2.0"}`),
						},
					},
					{
						Name:        "reviews-v1",
						Description: ptr.String("review events"),
						Version:     pkg.FixActiveVersion(),
						Spec: &directorSchema.EventSpecInput{
							Type:   directorSchema.EventSpecTypeAsyncAPI,
							Format: directorSchema.SpecFormatYaml,
							Data:   ptr.CLOB(`{"asyncapi":"1.1.0"}`),
						},
					},
				},
			},
		},
	}, suffix)
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

func generateAppInputForDifferentTenants(appInput directorSchema.ApplicationRegisterInput, suffix string) directorSchema.ApplicationRegisterInput {
	appInput.Name += "-" + suffix
	for _, bndl := range appInput.Bundles {
		bndl.Name = bndl.Name + "-" + suffix

		for _, apiDef := range bndl.APIDefinitions {
			apiDef.Name = apiDef.Name + "-" + suffix
		}

		for _, eventDef := range bndl.EventDefinitions {
			eventDef.Name = eventDef.Name + "-" + suffix
		}
	}
	return appInput
}

func WaitForFunction(interval time.Duration, timeout time.Duration, conditionalFunc func() bool) error {
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
