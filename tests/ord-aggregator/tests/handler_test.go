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
	tenantHeader            = "Tenant"
	applicationTypeLabelKey = "applicationType"

	descriptionField      = "description"
	shortDescriptionField = "shortDescription"
	apisField             = "apis"
	eventsField           = "events"

	expectedSpecType                        = "openapi-v3"
	expectedSpecFormat                      = "application/json"
	expectedSystemInstanceName              = "test-app"
	expectedSecondSystemInstanceName        = "second-test-app"
	expectedThirdSystemInstanceName         = "third-test-app"
	expectedFourthSystemInstanceName        = "fourth-test-app"
	expectedFifthSystemInstanceName         = "fifth-test-app"
	expectedSixthSystemInstanceName         = "sixth-test-app"
	expectedSystemInstanceDescription       = "test-app1-description"
	expectedSecondSystemInstanceDescription = "test-app2-description"
	expectedThirdSystemInstanceDescription  = "test-app3-description"
	expectedFourthSystemInstanceDescription = "test-app4-description"
	expectedFifthSystemInstanceDescription  = "test-app5-description"
	expectedSixthSystemInstanceDescription  = "test-app6-description"
	expectedBundleTitle                     = "BUNDLE TITLE"
	secondExpectedBundleTitle               = "BUNDLE TITLE 2"
	expectedBundleDescription               = "lorem ipsum dolor nsq sme"
	secondExpectedBundleDescription         = ""
	firstBundleOrdIDRegex                   = "ns:consumptionBundle:BUNDLE_ID(.+):v1"
	expectedPackageTitle                    = "PACKAGE 1 TITLE"
	expectedPackageDescription              = "lorem ipsum dolor set"
	firstProductTitle                       = "PRODUCT TITLE"
	firstProductShortDescription            = "lorem ipsum"
	secondProductTitle                      = "SAP Business Technology Platform"
	secondProductShortDescription           = "Accelerate business outcomes with integration, data to value, and extensibility."
	firstAPIExpectedTitle                   = "API TITLE"
	firstAPIExpectedDescription             = "lorem ipsum dolor sit amet"
	firstEventTitle                         = "EVENT TITLE"
	firstEventDescription                   = "lorem ipsum dolor sit amet"
	secondEventTitle                        = "EVENT TITLE 2"
	secondEventDescription                  = "lorem ipsum dolor sit amet"
	expectedTombstoneOrdIDRegex             = "ns:apiResource:API_ID2(.+):v1"
	expectedVendorTitle                     = "SAP SE"

	expectedNumberOfSystemInstances           = 6
	expectedNumberOfPackages                  = 6
	expectedNumberOfBundles                   = 12
	expectedNumberOfAPIs                      = 6
	expectedNumberOfResourceDefinitionsPerAPI = 3
	expectedNumberOfEvents                    = 12
	expectedNumberOfTombstones                = 6

	expectedNumberOfAPIsInFirstBundle    = 1
	expectedNumberOfAPIsInSecondBundle   = 1
	expectedNumberOfEventsInFirstBundle  = 2
	expectedNumberOfEventsInSecondBundle = 2

	testTimeoutAdditionalBuffer = 5 * time.Minute

	firstCorrelationID  = "sap.s4:communicationScenario:SAP_COM_0001"
	secondCorrelationID = "sap.s4:communicationScenario:SAP_COM_0002"

	documentationLabelKey         = "Documentation label key"
	documentationLabelFirstValue  = "Markdown Documentation with links"
	documentationLabelSecondValue = "With multiple values"
)

var (
	// The expected number is increased with initial number of global vendors/products before test execution
	expectedNumberOfProducts = 6
	expectedNumberOfVendors  = 6
)

func TestORDAggregator(t *testing.T) {
	basicORDConfigSecurity := &fixtures.ORDConfigSecurity{
		Username: testConfig.BasicUsername,
		Password: testConfig.BasicPassword,
	}

	oauthORDConfigSecurity := &fixtures.ORDConfigSecurity{
		Username: testConfig.ClientID,
		Password: testConfig.ClientSecret,
		TokenURL: testConfig.ExternalServicesMockBaseURL + "/secured/oauth/token",
	}

	accessStrategyConfigSecurity := &fixtures.ORDConfigSecurity{
		AccessStrategy: "sap:cmp-mtls:v1",
	}

	var appInput, secondAppInput, thirdAppInput, fourthAppInput, fifthAppInput, sixthAppInput directorSchema.ApplicationRegisterInput
	t.Run("Verifying ORD Document to be valid", func(t *testing.T) {
		// Unsecured config endpoint with full absolute URL in the webhook; unsecured document; doc baseURL from the webhook
		appInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSystemInstanceName, expectedSystemInstanceDescription, testConfig.ExternalServicesMockAbsoluteURL, nil)
		// Unsecured config endpoint with automatic .well-known/open-resource-discovery; unsecured document; doc baseURL from the webhook
		secondAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSecondSystemInstanceName, expectedSecondSystemInstanceDescription, testConfig.ExternalServicesMockUnsecuredURL, nil)
		// Basic secured config endpoint; unsecured document; doc baseURL from the webhook
		thirdAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedThirdSystemInstanceName, expectedThirdSystemInstanceDescription, testConfig.ExternalServicesMockBasicURL, basicORDConfigSecurity)
		// Oauth secured config endpoint; unsecured document; doc baseURL from the webhook
		fourthAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedFourthSystemInstanceName, expectedFourthSystemInstanceDescription, testConfig.ExternalServicesMockOauthURL, oauthORDConfigSecurity)
		// Unsecured config endpoint with full absolute URL in the webhook; cert secured document; doc baseURL configured in the config response
		fifthAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedFifthSystemInstanceName, expectedFifthSystemInstanceDescription, testConfig.ExternalServicesMockBaseURL+"/cert", nil)
		// Cert secured config endpoint with automatic .well-known/open-resource-discovery; cert secured document; doc baseURL from the webhook
		sixthAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSixthSystemInstanceName, expectedSixthSystemInstanceDescription, testConfig.ExternalServicesMockCertSecuredURL, accessStrategyConfigSecurity)

		systemInstancesMap := make(map[string]string)
		systemInstancesMap[expectedSystemInstanceName] = expectedSystemInstanceDescription
		systemInstancesMap[expectedSecondSystemInstanceName] = expectedSecondSystemInstanceDescription
		systemInstancesMap[expectedThirdSystemInstanceName] = expectedThirdSystemInstanceDescription
		systemInstancesMap[expectedFourthSystemInstanceName] = expectedFourthSystemInstanceDescription
		systemInstancesMap[expectedFifthSystemInstanceName] = expectedFifthSystemInstanceDescription
		systemInstancesMap[expectedSixthSystemInstanceName] = expectedSixthSystemInstanceDescription

		apisDefaultBundleMap := make(map[string]string)
		apisDefaultBundleMap[firstAPIExpectedTitle] = firstBundleOrdIDRegex

		eventsMap := make(map[string]string)
		eventsMap[firstEventTitle] = firstEventDescription
		eventsMap[secondEventTitle] = secondEventDescription

		eventsDefaultBundleMap := make(map[string]string)
		eventsDefaultBundleMap[firstEventTitle] = firstBundleOrdIDRegex

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

		bundlesCorrelationIDs := make(map[string][]string)
		bundlesCorrelationIDs[expectedBundleTitle] = []string{firstCorrelationID, secondCorrelationID}
		bundlesCorrelationIDs[secondExpectedBundleTitle] = []string{firstCorrelationID, secondCorrelationID}

		documentationLabelsPossibleValues := []string{documentationLabelFirstValue, documentationLabelSecondValue}

		productsMap := make(map[string]string)
		productsMap[firstProductTitle] = firstProductShortDescription
		productsMap[secondProductTitle] = secondProductShortDescription

		ctx := context.Background()

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, "", "test-int-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

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
		httpClient.Timeout = 20 * time.Second

		// Get vendors
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		numberOfGlobalVendors := len(gjson.Get(respBody, "value").Array())
		expectedNumberOfVendors += numberOfGlobalVendors

		// Get products
		respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/products?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		numberOfGlobalProducts := len(gjson.Get(respBody, "value").Array())
		expectedNumberOfProducts += numberOfGlobalProducts

		// Register systems
		app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)

		secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, secondAppInput)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &secondApp)
		require.NoError(t, err)

		thirdApp, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, thirdAppInput)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &thirdApp)
		require.NoError(t, err)

		fourthApp, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, fourthAppInput)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &fourthApp)
		require.NoError(t, err)

		fifthApp, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, fifthAppInput)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &fifthApp)
		require.NoError(t, err)

		sixthApp, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, sixthAppInput)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &sixthApp)
		require.NoError(t, err)

		scheduleTime, err := parseCronTime(testConfig.AggregatorSchedule)
		require.NoError(t, err)

		defaultTestTimeout := 2*scheduleTime + testTimeoutAdditionalBuffer
		defaultCheckInterval := defaultTestTimeout / 20

		err = verifyORDDocument(defaultCheckInterval, defaultTestTimeout, func() bool {
			var respBody string

			// Verify system instances
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfSystemInstances {
				t.Log("Missing System Instances...will try again")
				return false
			}
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, systemInstancesMap, expectedNumberOfSystemInstances, descriptionField)

			// Verify packages
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfPackages {
				t.Log("Missing Packages...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfPackages)
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfPackages, expectedPackageTitle, expectedPackageDescription, descriptionField)
			t.Log("Successfully verified packages")

			// Verify bundles
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfBundles {
				t.Log("Missing Bundles...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfBundles)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, bundlesMap, expectedNumberOfBundles, descriptionField)
			assertions.AssertBundleCorrelationIds(t, respBody, bundlesCorrelationIDs, expectedNumberOfBundles)
			ordAndInternalIDsMappingForBundles := storeMappingBetweenORDAndInternalBundleID(t, respBody, expectedNumberOfBundles)
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
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfProducts)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, productsMap, expectedNumberOfProducts, shortDescriptionField)
			t.Log("Successfully verified products")

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfAPIs {
				t.Log("Missing APIs...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfAPIs)
			// In the document there are actually 2 APIs but there is a tombstone for the second one so in the end there will be only one API
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfAPIs, firstAPIExpectedTitle, firstAPIExpectedDescription, descriptionField)
			t.Log("Successfully verified apis")

			// Verify defaultBundle for apis
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfAPIs, apisDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for apis")

			// Verify the api spec
			specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
			require.Equal(t, expectedNumberOfResourceDefinitionsPerAPI, len(specs))

			var specURL string
			for _, s := range specs {
				specType := s.Get("type").String()
				specFormat := s.Get("mediaType").String()
				if specType == expectedSpecType && specFormat == expectedSpecFormat {
					specURL = s.Get("url").String()
					break
				}
			}

			respBody = makeRequestWithHeaders(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(respBody) == 0 || !strings.Contains(respBody, "swagger") {
				t.Logf("Spec %s not successfully fetched... will try again", specURL)
				return false
			}
			t.Log("Successfully verified api spec")

			// Verify events
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEvents {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfEvents)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, eventsMap, expectedNumberOfEvents, descriptionField)
			t.Log("Successfully verified events")

			// Verify defaultBundle for events
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfEvents, eventsDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for events")

			// Verify tombstones
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/tombstones?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfTombstones {
				t.Log("Missing Tombstones...will try again")
				return false
			}
			assertions.AssertTombstoneFromORDService(t, respBody, expectedNumberOfTombstones, expectedTombstoneOrdIDRegex)
			t.Log("Successfully verified tombstones")

			// Verify vendors
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfVendors {
				t.Log("Missing Vendors...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfVendors)
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

func storeMappingBetweenORDAndInternalBundleID(t *testing.T, respBody string, numberOfEntities int) map[string]string {
	ordAndInternalIDsMapping := make(map[string]string)

	for i := 0; i < numberOfEntities; i++ {
		internalBundleID := gjson.Get(respBody, fmt.Sprintf("value.%d.id", i)).String()
		require.NotEmpty(t, internalBundleID)

		ordID := gjson.Get(respBody, fmt.Sprintf("value.%d.ordId", i)).String()
		require.NotEmpty(t, ordID)

		ordAndInternalIDsMapping[internalBundleID] = ordID
	}

	return ordAndInternalIDsMapping
}
