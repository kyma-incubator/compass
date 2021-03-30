package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/request"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	tenantHeader = "Tenant"

	descriptionField      = "value.0.description"
	shortDescriptionField = "value.0.shortDescription"

	expectedSystemInstanceName        = "test-app"
	expectedSystemInstanceDescription = "test-app-description"
	expectedBundleTitle               = "BUNDLE TITLE"
	expectedBundleDescription         = "lorem ipsum dolor nsq sme"
	expectedPackageTitle              = "PACKAGE 1 TITLE"
	expectedPackageDescription        = "lorem ipsum dolor set"
	expectedProductTitle              = "PRODUCT TITLE"
	expectedProductShortDescription   = "lorem ipsum"
	firstAPIExpectedTitle             = "API TITLE"
	firstAPIExpectedDescription       = "lorem ipsum dolor sit amet"
	firstEventTitle                   = "EVENT TITLE"
	firstEventDescription             = "lorem ipsum dolor sit amet"
	secondEventTitle                  = "EVENT TITLE 2"
	secondEventDescription            = "lorem ipsum dolor sit amet"
	expectedTombstoneOrdID            = "ns:apiResource:API_ID2:v1"
	expectedVendorTitle               = "SAP"

	expectedNumberOfSystemInstances = 1
	expectedNumberOfPackages        = 1
	expectedNumberOfBundles         = 1
	expectedNumberOfProducts        = 1
	expectedNumberOfAPIs            = 1
	expectedNumberOfEvents          = 2
	expectedNumberOfTombstones      = 1
	expectedNumberOfVendors         = 1

	testTimeoutAdditionalBuffer = 1 * time.Minute
)

func TestORDAggregator(t *testing.T) {
	appInput := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSystemInstanceName, expectedSystemInstanceDescription, testConfig.ExternalServicesMockBaseURL)

	eventsMap := make(map[string]string, 0)
	eventsMap[firstEventTitle] = firstEventDescription
	eventsMap[secondEventTitle] = secondEventDescription

	ctx := context.Background()

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NoError(t, err)

	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, "", "test-int-system")
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys.ID)

	intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys.ID)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSystemCredentials.ID)

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

	scheduleTime, err := parseCronTime(testConfig.AggregatorSchedule)
	require.NoError(t, err)

	defaultTestTimeout := scheduleTime + testTimeoutAdditionalBuffer
	defaultCheckInterval := scheduleTime / 20

	t.Run("Verifying ORD Document to be valid", func(t *testing.T) {
		err = verifyORDDocument(defaultCheckInterval, defaultTestTimeout, func() bool {
			var respBody string

			// Verify system instances
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})
			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing System Instances...will try again")
				return false
			}
			assertions.AssertEntityFromORDService(t, respBody, expectedNumberOfSystemInstances, expectedSystemInstanceName, expectedSystemInstanceDescription, descriptionField)

			// Verify packages
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Packages...will try again")
				return false
			}
			assertions.AssertEntityFromORDService(t, respBody, expectedNumberOfPackages, expectedPackageTitle, expectedPackageDescription, descriptionField)

			// Verify bundles
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Bundles...will try again")
				return false
			}
			assertions.AssertEntityFromORDService(t, respBody, expectedNumberOfBundles, expectedBundleTitle, expectedBundleDescription, descriptionField)

			// Verify products
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/products?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Products...will try again")
				return false
			}
			assertions.AssertEntityFromORDService(t, respBody, expectedNumberOfProducts, expectedProductTitle, expectedProductShortDescription, shortDescriptionField)

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing APIs...will try again")
				return false
			}
			// In the document there are actually 2 APIs but there is a tombstone for the second one so in the end there will be only one API
			assertions.AssertEntityFromORDService(t, respBody, expectedNumberOfAPIs, firstAPIExpectedTitle, firstAPIExpectedDescription, descriptionField)

			// Verify events
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertEventFromORDService(t, respBody, eventsMap, expectedNumberOfEvents)

			// Verify tombstones
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/tombstones?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Tombstones...will try again")
				return false
			}
			assertions.AssertTombstoneFromORDService(t, respBody, expectedNumberOfTombstones, expectedTombstoneOrdID)

			// Verify vendors
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}})

			if len(gjson.Get(respBody, "value").Array()) == 0 {
				t.Log("Missing Vendors...will try again")
				return false
			}
			assertions.AssertVendorFromORDService(t, respBody, expectedNumberOfVendors, expectedVendorTitle)

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

func makeRequestWithHeaders(t *testing.T, httpClient *http.Client, url string, headers map[string][]string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
}
