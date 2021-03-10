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

	applicationName         = "test-app"
	applicationDescription  = "test-app-description"
	bundleTitle             = "BUNDLE TITLE"
	bundleDescription       = "lorem ipsum dolor nsq sme"
	packageTitle            = "PACKAGE 1 TITLE"
	packageDescription      = "lorem ipsum dolor set"
	productTitle            = "PRODUCT TITLE"
	productShortDescription = "lorem ipsum"
	firstAPITitle           = "API TITLE"
	firstAPIDescription     = "lorem ipsum dolor sit amet"
	firstEventTitle         = "EVENT TITLE"
	firstEventDescription   = "lorem ipsum dolor sit amet"
	secondEventTitle        = "EVENT TITLE 2"
	secondEventDescription  = "lorem ipsum dolor sit amet"
	tombstoneOrdID          = "ns:apiResource:API_ID2:v1"
	vendorTitle             = "SAP"

	expectedNumberOfSystemInstances = 1
	expectedNumberOfPackages        = 1
	expectedNumberOfBundles         = 1
	expectedNumberOfProducts        = 1
	expectedNumberOfAPIs            = 1
	expectedNumberOfEvents          = 2
	expectedNumberOfTombstones      = 1
	expectedNumberOfVendors         = 1
)

func TestORDAggregator(t *testing.T) {
	appInput := createApp()

	eventsMap := make(map[string]string, 0)
	eventsMap[firstEventTitle] = firstEventDescription
	eventsMap[secondEventTitle] = secondEventDescription

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

	t.Run("Verifying ORD Document to be valid", func(t *testing.T) {
		err = verifyORDDocument(testConfig.DefaultCheckInterval, testConfig.DefaultTestTimeout, func() bool {
			var respBody string

			// Verify system instances
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})
			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing System Instances...will try again")
				return false
			}

			require.Equal(t, expectedNumberOfSystemInstances, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, applicationName, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, applicationDescription, gjson.Get(respBody, "value.0.description").String())

			// Verify packages
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Packages...will try again")
				return false
			}

			require.Equal(t, expectedNumberOfPackages, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, packageTitle, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, packageDescription, gjson.Get(respBody, "value.0.description").String())

			// Verify bundles
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Bundles...will try again")
				return false
			}

			require.Equal(t, expectedNumberOfBundles, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, bundleTitle, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, bundleDescription, gjson.Get(respBody, "value.0.description").String())

			// Verify products
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/products?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Products...will try again")
				return false
			}

			require.Equal(t, expectedNumberOfProducts, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, productTitle, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, productShortDescription, gjson.Get(respBody, "value.0.shortDescription").String())

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing APIs...will try again")
				return false
			}

			// In the document there are actually 2 APIs but there is a tombstone for the second one so in the end there will be only one API
			require.Equal(t, expectedNumberOfAPIs, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, firstAPITitle, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, firstAPIDescription, gjson.Get(respBody, "value.0.description").String())

			// Verify events
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Events...will try again")
				return false
			}

			numberOfEvents := len(gjson.Get(respBody, "value").Array())
			require.Equal(t, expectedNumberOfEvents, numberOfEvents)

			for i := 0; i < numberOfEvents; i++ {
				eventTitle := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
				require.NotEmpty(t, eventTitle)

				eventDescription, exists := eventsMap[eventTitle]
				require.True(t, exists)

				require.Equal(t, eventDescription, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
			}

			// Verify tombstones
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/tombstones?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Tombstones...will try again")
				return false
			}

			require.Equal(t, expectedNumberOfTombstones, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, tombstoneOrdID, gjson.Get(respBody, "value.0.ordId").String())

			// Verify vendors
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Vendors...will try again")
				return false
			}

			require.Equal(t, expectedNumberOfVendors, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, vendorTitle, gjson.Get(respBody, "value.0.title").String())

			return true
		})
		require.NoError(t, err)
	})
}

func verifyORDDocument(interval time.Duration, timeout time.Duration, conditionalFunc func() bool) error {
	done := time.After(timeout)
	ticker := time.Tick(interval)

	for {
		if conditionalFunc() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for entities to be present in DB")
		case <-ticker:
		}
	}
}

func createApp() directorSchema.ApplicationRegisterInput {
	return directorSchema.ApplicationRegisterInput{
		Name:        applicationName,
		Description: ptr.String(applicationDescription),
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
