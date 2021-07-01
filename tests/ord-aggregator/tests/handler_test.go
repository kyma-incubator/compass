package tests

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
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

	descriptionField      = "description"
	shortDescriptionField = "shortDescription"
	apisField             = "apis"
	eventsField           = "events"

	expectedExternalServicesMockSpecURL = "expectedExternal"
	expectedSystemInstanceName              = "test-app"
	expectedSecondSystemInstanceName        = "second-test-app"
	expectedSystemInstanceDescription       = "test-app-description"
	expectedSecondSystemInstanceDescription = "test-app-description"
	expectedBundleTitle                     = "BUNDLE TITLE"
	secondExpectedBundleTitle               = "BUNDLE TITLE 2"
	expectedBundleDescription               = "lorem ipsum dolor nsq sme"
	secondExpectedBundleDescription         = "foo bar"
	expectedPackageTitle                    = "PACKAGE 1 TITLE"
	expectedPackageDescription              = "lorem ipsum dolor set"
	expectedProductTitle                    = "PRODUCT TITLE"
	expectedProductShortDescription         = "lorem ipsum"
	firstAPIExpectedTitle                   = "API TITLE"
	firstAPIExpectedDescription             = "lorem ipsum dolor sit amet"
	firstEventTitle                         = "EVENT TITLE"
	firstEventDescription                   = "lorem ipsum dolor sit amet"
	secondEventTitle                        = "EVENT TITLE 2"
	secondEventDescription                  = "lorem ipsum dolor sit amet"
	expectedTombstoneOrdID                  = "ns:apiResource:API_ID2:v1"
	expectedVendorTitle                     = "SAP"

	expectedNumberOfSystemInstances           = 2
	expectedNumberOfPackages                  = 2
	expectedNumberOfBundles                   = 4
	expectedNumberOfProducts                  = 2
	expectedNumberOfAPIs                      = 2
	expectedNumberOfResourceDefinitionsPerAPI = 3
	expectedNumberOfEvents                    = 4
	expectedNumberOfTombstones                = 2
	expectedNumberOfVendors                   = 4

	expectedNumberOfAPIsInFirstBundle    = 1
	expectedNumberOfAPIsInSecondBundle   = 1
	expectedNumberOfEventsInFirstBundle  = 2
	expectedNumberOfEventsInSecondBundle = 2

	testTimeoutAdditionalBuffer = 5 * time.Minute
)

func TestORDAggregator(t *testing.T) {
	appInput := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSystemInstanceName, expectedSystemInstanceDescription, testConfig.ExternalServicesMockBaseURL)
	secondAppInput := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSecondSystemInstanceName, expectedSecondSystemInstanceDescription, testConfig.ExternalServicesMockBaseURL)

	systemInstancesMap := make(map[string]string)
	systemInstancesMap[expectedSystemInstanceName] = expectedSystemInstanceDescription
	systemInstancesMap[expectedSecondSystemInstanceName] = expectedSecondSystemInstanceDescription

	eventsMap := make(map[string]string)
	eventsMap[firstEventTitle] = firstEventDescription
	eventsMap[secondEventTitle] = secondEventDescription

	bundlesMap := make(map[string]string)
	bundlesMap[expectedBundleTitle] = expectedBundleDescription
	bundlesMap[secondExpectedBundleTitle] = secondExpectedBundleDescription

	bundlesAPIsNumberMap := make(map[string]int)
	bundlesAPIsNumberMap[expectedBundleTitle] = expectedNumberOfAPIsInFirstBundle
	bundlesAPIsNumberMap[secondExpectedBundleTitle] = expectedNumberOfAPIsInSecondBundle

	bundlesAPIsData := make(map[string][]string)
	bundlesAPIsData[expectedBundleTitle] = []string{firstAPIExpectedTitle}
	bundlesAPIsData[secondExpectedBundleTitle] = []string{firstAPIExpectedTitle}

	bundlesEventsNumberMap := make(map[string]int)
	bundlesEventsNumberMap[expectedBundleTitle] = expectedNumberOfEventsInFirstBundle
	bundlesEventsNumberMap[secondExpectedBundleTitle] = expectedNumberOfEventsInSecondBundle

	bundlesEventsData := make(map[string][]string)
	bundlesEventsData[expectedBundleTitle] = []string{firstEventTitle, secondEventTitle}
	bundlesEventsData[secondExpectedBundleTitle] = []string{firstEventTitle, secondEventTitle}

	ctx := context.Background()

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, appInput)
	require.NoError(t, err)
	secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, secondAppInput)
	require.NoError(t, err)

	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, app.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, secondApp.ID)

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
	defaultCheckInterval := defaultTestTimeout / 20

	t.Run("Verifying ORD Document to be valid", func(t *testing.T) {
		err = verifyORDDocument(defaultCheckInterval, defaultTestTimeout, func() bool {
			var respBody string

			// Verify system instances
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfSystemInstances {
				t.Log("Missing System Instances...will try again")
				return false
			}
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, systemInstancesMap, expectedNumberOfSystemInstances)

			// Verify packages
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfPackages {
				t.Log("Missing Packages...will try again")
				return false
			}
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfPackages, expectedPackageTitle, expectedPackageDescription, descriptionField)
			t.Log("Successfully verified packages")

			// Verify bundles
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfBundles {
				t.Log("Missing Bundles...will try again")
				return false
			}
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, bundlesMap, expectedNumberOfBundles)
			t.Log("Successfully verified bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=apis&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, apisField, bundlesAPIsNumberMap, bundlesAPIsData)
			t.Log("Successfully verified relation between apis and bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, eventsField, bundlesEventsNumberMap, bundlesEventsData)
			t.Log("Successfully verified relation between events and bundles")

			// Verify products
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/products?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfProducts {
				t.Log("Missing Products...will try again")
				return false
			}
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfProducts, expectedProductTitle, expectedProductShortDescription, shortDescriptionField)
			t.Log("Successfully verified products")

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfAPIs {
				t.Log("Missing APIs...will try again")
				return false
			}
			// In the document there are actually 2 APIs but there is a tombstone for the second one so in the end there will be only one API
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfAPIs, firstAPIExpectedTitle, firstAPIExpectedDescription, descriptionField)
			t.Log("Successfully verified apis")

			// Verify the api spec
			specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
			require.Equal(t, expectedNumberOfResourceDefinitionsPerAPI, len(specs))

			var specURL string
			for _, s := range specs {
				url := s.Get("url").String()
				if strings.Contains(url, expectedExternalServicesMockSpecURL) {
					specURL = url
					break
				}
			}

			respBody = makeRequestWithHeaders(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(respBody) == 0 || !strings.Contains(respBody, "swagger") {
				t.Logf("Spec %s not successfully fetched... will try again", specURL)
				return false
			}

			// Verify events
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEvents {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, eventsMap, expectedNumberOfEvents)
			t.Log("Successfully verified events")

			// Verify tombstones
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/tombstones?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfTombstones {
				t.Log("Missing Tombstones...will try again")
				return false
			}
			assertions.AssertTombstoneFromORDService(t, respBody, expectedNumberOfTombstones, expectedTombstoneOrdID)
			t.Log("Successfully verified tombstones")

			// Verify vendors
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfVendors {
				t.Log("Missing Vendors...will try again")
				return false
			}
			assertions.AssertVendorFromORDService(t, respBody, expectedNumberOfVendors, expectedVendorTitle)
			t.Log("Successfully verified vendors")

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
