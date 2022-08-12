package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
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
	internalVisibilityScope = "internal_visibility:read"

	descriptionField      = "description"
	shortDescriptionField = "shortDescription"
	apisField             = "apis"
	eventsField           = "events"
	publicAPIsField       = "publicAPIs"
	publicEventsField     = "publicEvents"

	expectedSpecType                         = "openapi-v3"
	expectedSpecFormat                       = "application/json"
	expectedSystemInstanceName               = "test-app"
	expectedSecondSystemInstanceName         = "second-test-app"
	expectedThirdSystemInstanceName          = "third-test-app"
	expectedFourthSystemInstanceName         = "fourth-test-app"
	expectedFifthSystemInstanceName          = "fifth-test-app"
	expectedSixthSystemInstanceName          = "sixth-test-app"
	expectedSeventhSystemInstanceName        = "seventh-test-app"
	expectedSystemInstanceDescription        = "test-app1-description"
	expectedSecondSystemInstanceDescription  = "test-app2-description"
	expectedThirdSystemInstanceDescription   = "test-app3-description"
	expectedFourthSystemInstanceDescription  = "test-app4-description"
	expectedFifthSystemInstanceDescription   = "test-app5-description"
	expectedSixthSystemInstanceDescription   = "test-app6-description"
	expectedSeventhSystemInstanceDescription = "test-app7-description"
	expectedBundleTitle                      = "BUNDLE TITLE"
	secondExpectedBundleTitle                = "BUNDLE TITLE 2"
	expectedBundleDescription                = "lorem ipsum dolor nsq sme"
	secondExpectedBundleDescription          = ""
	firstBundleOrdIDRegex                    = "ns:consumptionBundle:BUNDLE_ID(.+):v1"
	expectedPackageTitle                     = "PACKAGE 1 TITLE"
	expectedPackageDescription               = "lorem ipsum dolor set"
	firstProductTitle                        = "PRODUCT TITLE"
	firstProductShortDescription             = "lorem ipsum"
	secondProductTitle                       = "SAP Business Technology Platform"
	secondProductShortDescription            = "Accelerate business outcomes with integration, data to value, and extensibility."
	firstAPIExpectedTitle                    = "API TITLE"
	firstAPIExpectedDescription              = "lorem ipsum dolor sit amet"
	firstAPIExpectedNumberOfSpecs            = 3
	secondAPIExpectedTitle                   = "API TITLE INTERNAL"
	secondAPIExpectedDescription             = "Test description internal"
	secondAPIExpectedNumberOfSpecs           = 2
	thirdAPIExpectedTitle                    = "API TITLE PRIVATE"
	thirdAPIExpectedDescription              = "Test description private"
	thirdAPIExpectedNumberOfSpecs            = 2
	firstEventTitle                          = "EVENT TITLE"
	firstEventDescription                    = "lorem ipsum dolor sit amet"
	secondEventTitle                         = "EVENT TITLE 2"
	secondEventDescription                   = "lorem ipsum dolor sit amet"
	thirdEventTitle                          = "EVENT TITLE INTERNAL"
	thirdEventDescription                    = "Test description internal"
	fourthEventTitle                         = "EVENT TITLE PRIVATE"
	fourthEventDescription                   = "Test description private"
	expectedTombstoneOrdIDRegex              = "ns:apiResource:API_ID2(.+):v1"
	expectedVendorTitle                      = "SAP SE"

	expectedNumberOfSystemInstances               = 7
	expectedNumberOfSystemInstancesInSubscription = 1
	expectedNumberOfPackages                      = 7
	expectedNumberOfPackagesInSubscription        = 1
	expectedNumberOfBundles                       = 14
	expectedNumberOfBundlesInSubscription         = 2
	expectedNumberOfAPIs                          = 21
	expectedNumberOfAPIsInSubscription            = 3
	expectedNumberOfEvents                        = 28
	expectedNumberOfEventsInSubscription          = 4
	expectedNumberOfTombstones                    = 7
	expectedNumberOfTombstonesInSubscription      = 1

	expectedNumberOfPublicAPIs   = 7
	expectedNumberOfPublicEvents = 14

	expectedNumberOfAPIsInFirstBundle    = 2
	expectedNumberOfAPIsInSecondBundle   = 2
	expectedNumberOfEventsInFirstBundle  = 3
	expectedNumberOfEventsInSecondBundle = 3

	expectedNumberOfPublicAPIsInFirstBundle    = 1
	expectedNumberOfPublicAPIsInSecondBundle   = 1
	expectedNumberOfPublicEventsInFirstBundle  = 2
	expectedNumberOfPublicEventsInSecondBundle = 2

	testTimeoutAdditionalBuffer = 7 * time.Minute

	firstCorrelationID  = "sap.s4:communicationScenario:SAP_COM_0001"
	secondCorrelationID = "sap.s4:communicationScenario:SAP_COM_0002"

	documentationLabelKey         = "Documentation label key"
	documentationLabelFirstValue  = "Markdown Documentation with links"
	documentationLabelSecondValue = "With multiple values"
)

var (
	// The expected number is increased with initial number of global vendors/products before test execution
	expectedNumberOfProducts               = 7
	expectedNumberOfProductsInSubscription = 1
	expectedNumberOfVendors                = 7
	expectedNumberOfVendorsInSubscription  = 1
)

func TestORDAggregator(stdT *testing.T) {
	t := testingx.NewT(stdT)

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

	var appInput, secondAppInput, thirdAppInput, fourthAppInput, fifthAppInput, sixthAppInput, seventhAppInput directorSchema.ApplicationRegisterInput
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
		sixthAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSixthSystemInstanceName, expectedSixthSystemInstanceDescription, testConfig.ExternalServicesMockOrdCertSecuredURL, accessStrategyConfigSecurity)
		// Unsecured config endpoint with automatic .well-known/open-resource-discovery; unsecured document; doc baseURL from the webhook; with additional content
		seventhAppInput = fixtures.FixSampleApplicationRegisterInputWithORDWebhooks(expectedSeventhSystemInstanceName, expectedSeventhSystemInstanceDescription, testConfig.ExternalServicesMockUnsecuredWithAdditionalContentURL, nil)

		systemInstancesMap := make(map[string]string)
		systemInstancesMap[expectedSystemInstanceName] = expectedSystemInstanceDescription
		systemInstancesMap[expectedSecondSystemInstanceName] = expectedSecondSystemInstanceDescription
		systemInstancesMap[expectedThirdSystemInstanceName] = expectedThirdSystemInstanceDescription
		systemInstancesMap[expectedFourthSystemInstanceName] = expectedFourthSystemInstanceDescription
		systemInstancesMap[expectedFifthSystemInstanceName] = expectedFifthSystemInstanceDescription
		systemInstancesMap[expectedSixthSystemInstanceName] = expectedSixthSystemInstanceDescription
		systemInstancesMap[expectedSeventhSystemInstanceName] = expectedSeventhSystemInstanceDescription

		apisMap := make(map[string]string)
		apisMap[firstAPIExpectedTitle] = firstAPIExpectedDescription
		apisMap[secondAPIExpectedTitle] = secondAPIExpectedDescription
		apisMap[thirdAPIExpectedTitle] = thirdAPIExpectedDescription

		publicApisMap := make(map[string]string)
		publicApisMap[firstAPIExpectedTitle] = firstAPIExpectedDescription

		apisDefaultBundleMap := make(map[string]string)
		apisDefaultBundleMap[firstAPIExpectedTitle] = firstBundleOrdIDRegex

		apiSpecsMap := make(map[string]int)
		apiSpecsMap[firstAPIExpectedTitle] = firstAPIExpectedNumberOfSpecs
		apiSpecsMap[secondAPIExpectedTitle] = secondAPIExpectedNumberOfSpecs
		apiSpecsMap[thirdAPIExpectedTitle] = thirdAPIExpectedNumberOfSpecs

		eventsMap := make(map[string]string)
		eventsMap[firstEventTitle] = firstEventDescription
		eventsMap[secondEventTitle] = secondEventDescription
		eventsMap[thirdEventTitle] = thirdEventDescription
		eventsMap[fourthEventTitle] = fourthEventDescription

		publicEventsMap := make(map[string]string)
		publicEventsMap[firstEventTitle] = firstEventDescription
		publicEventsMap[secondEventTitle] = secondEventDescription

		eventsDefaultBundleMap := make(map[string]string)
		eventsDefaultBundleMap[firstEventTitle] = firstBundleOrdIDRegex

		apisAndEventsNumber := make(map[string]int)
		apisAndEventsNumber[apisField] = expectedNumberOfAPIsInFirstBundle + expectedNumberOfAPIsInSecondBundle
		apisAndEventsNumber[publicAPIsField] = expectedNumberOfPublicAPIsInFirstBundle + expectedNumberOfPublicAPIsInSecondBundle
		apisAndEventsNumber[eventsField] = expectedNumberOfEventsInFirstBundle + expectedNumberOfEventsInSecondBundle
		apisAndEventsNumber[publicEventsField] = expectedNumberOfPublicEventsInFirstBundle + expectedNumberOfPublicEventsInSecondBundle

		bundlesMap := make(map[string]string)
		bundlesMap[expectedBundleTitle] = expectedBundleDescription
		bundlesMap[secondExpectedBundleTitle] = secondExpectedBundleDescription

		bundlesAPIsNumberMap := make(map[string]int)
		bundlesAPIsNumberMap[expectedBundleTitle] = expectedNumberOfAPIsInFirstBundle
		bundlesAPIsNumberMap[secondExpectedBundleTitle] = expectedNumberOfAPIsInSecondBundle

		bundlesAPIsData := make(map[string][]string)
		bundlesAPIsData[expectedBundleTitle] = []string{firstAPIExpectedTitle, secondAPIExpectedTitle}
		bundlesAPIsData[secondExpectedBundleTitle] = []string{firstAPIExpectedTitle, thirdAPIExpectedTitle}

		bundlesEventsNumberMap := make(map[string]int)
		bundlesEventsNumberMap[expectedBundleTitle] = expectedNumberOfEventsInFirstBundle
		bundlesEventsNumberMap[secondExpectedBundleTitle] = expectedNumberOfEventsInSecondBundle

		bundlesEventsData := make(map[string][]string)
		bundlesEventsData[expectedBundleTitle] = []string{firstEventTitle, secondEventTitle, thirdEventTitle}
		bundlesEventsData[secondExpectedBundleTitle] = []string{firstEventTitle, secondEventTitle, fourthEventTitle}

		bundlesCorrelationIDs := make(map[string][]string)
		bundlesCorrelationIDs[expectedBundleTitle] = []string{firstCorrelationID, secondCorrelationID}
		bundlesCorrelationIDs[secondExpectedBundleTitle] = []string{firstCorrelationID, secondCorrelationID}

		documentationLabelsPossibleValues := []string{documentationLabelFirstValue, documentationLabelSecondValue}

		productsMap := make(map[string]string)
		productsMap[firstProductTitle] = firstProductShortDescription
		productsMap[secondProductTitle] = secondProductShortDescription

		ctx := context.Background()

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", "test-int-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys.ID)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSystemCredentials.ID)

		unsecuredHttpClient := http.DefaultClient
		unsecuredHttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		oauthCredentialData, ok := intSystemCredentials.Auth.Credential.(*directorSchema.OAuthCredentialData)
		require.True(t, ok)

		cfgWithInternalVisibilityScope := &clientcredentials.Config{
			ClientID:     oauthCredentialData.ClientID,
			ClientSecret: oauthCredentialData.ClientSecret,
			TokenURL:     oauthCredentialData.URL,
			Scopes:       []string{internalVisibilityScope},
		}

		cfgWithoutScopes := &clientcredentials.Config{
			ClientID:     oauthCredentialData.ClientID,
			ClientSecret: oauthCredentialData.ClientSecret,
			TokenURL:     oauthCredentialData.URL,
		}

		ctx = context.WithValue(ctx, oauth2.HTTPClient, unsecuredHttpClient)
		httpClient := cfgWithInternalVisibilityScope.Client(ctx)
		httpClient.Timeout = 20 * time.Second

		httpClientWithoutVisibilityScope := cfgWithoutScopes.Client(ctx)
		httpClientWithoutVisibilityScope.Timeout = 20 * time.Second

		// create client to call Director graphql api with internal_visibility:read scope
		accessTokenWithInternalVisibility := token.GetAccessToken(t, oauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClientWithInternalVisibility := gql.NewAuthorizedGraphQLClientWithCustomURL(accessTokenWithInternalVisibility, testConfig.DirectorGraphqlOauthURL)

		// create client to call Director graphql api without internal_visibility:read scope
		accessTokenWithoutInternalVisibility := token.GetAccessToken(t, oauthCredentialData, token.IntegrationSystemScopesWithoutInternalVisibility)
		oauthGraphQLClientWithoutInternalVisibility := gql.NewAuthorizedGraphQLClientWithCustomURL(accessTokenWithoutInternalVisibility, testConfig.DirectorGraphqlOauthURL)

		globalProductsNumber, globalVendorsNumber := getGlobalResourcesNumber(ctx, t, unsecuredHttpClient)
		t.Logf("Global products number: %d, Global vendors number: %d", globalProductsNumber, globalVendorsNumber)

		expectedTotalNumberOfProducts := expectedNumberOfProducts + globalProductsNumber
		expectedTotalNumberOfVendors := expectedNumberOfVendors + globalVendorsNumber

		// Register systems
		app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)

		secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, secondAppInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &secondApp)
		require.NoError(t, err)

		thirdApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, thirdAppInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &thirdApp)
		require.NoError(t, err)

		fourthApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, fourthAppInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &fourthApp)
		require.NoError(t, err)

		fifthApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, fifthAppInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &fifthApp)
		require.NoError(t, err)

		sixthApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, sixthAppInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &sixthApp)
		require.NoError(t, err)

		seventhApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, seventhAppInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &seventhApp)
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

			if len(gjson.Get(respBody, "value").Array()) < expectedTotalNumberOfProducts {
				t.Log("Missing Products...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedTotalNumberOfProducts)
			assertions.AssertProducts(t, respBody, productsMap, expectedTotalNumberOfProducts, shortDescriptionField)
			t.Log("Successfully verified products")

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfAPIs {
				t.Log("Missing APIs...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfAPIs)
			// In the document there are actually 4 APIs but there is a tombstone for one of them so in the end there will be 3 APIs
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, apisMap, expectedNumberOfAPIs, descriptionField)
			t.Log("Successfully verified apis")

			// Verify defaultBundle for apis
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfAPIs, apisDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for apis")

			// Verify the api spec
			specs := assertions.AssertSpecsFromORDService(t, respBody, expectedNumberOfAPIs, apiSpecsMap)
			t.Log("Successfully verified specs for apis")

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

			// verify public apis via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicApisMap, apisField, expectedNumberOfPublicAPIs)

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

			// verify public events via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicEventsMap, eventsField, expectedNumberOfPublicEvents)

			// verify apis and events visibility via Director's graphql
			verifyEntitiesVisibilityViaGraphql(t, oauthGraphQLClientWithInternalVisibility, oauthGraphQLClientWithoutInternalVisibility, mergeMaps(apisMap, eventsMap), mergeMaps(publicApisMap, publicEventsMap), apisAndEventsNumber, app.ID)

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

			if len(gjson.Get(respBody, "value").Array()) < expectedTotalNumberOfVendors {
				t.Log("Missing Vendors...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedTotalNumberOfVendors)
			assertions.AssertVendorFromORDService(t, respBody, expectedTotalNumberOfVendors, expectedNumberOfVendors, expectedVendorTitle)
			t.Log("Successfully verified vendors")

			return true
		})
		require.NoError(t, err)
	})
	t.Run("Verifying ORD Document for subscribed tenant", func(t *testing.T) {
		ctx := context.Background()

		apisMap := make(map[string]string)
		apisMap[firstAPIExpectedTitle] = firstAPIExpectedDescription
		apisMap[secondAPIExpectedTitle] = secondAPIExpectedDescription
		apisMap[thirdAPIExpectedTitle] = thirdAPIExpectedDescription

		publicApisMap := make(map[string]string)
		publicApisMap[firstAPIExpectedTitle] = firstAPIExpectedDescription

		apisDefaultBundleMap := make(map[string]string)
		apisDefaultBundleMap[firstAPIExpectedTitle] = firstBundleOrdIDRegex

		apiSpecsMap := make(map[string]int)
		apiSpecsMap[firstAPIExpectedTitle] = firstAPIExpectedNumberOfSpecs
		apiSpecsMap[secondAPIExpectedTitle] = secondAPIExpectedNumberOfSpecs
		apiSpecsMap[thirdAPIExpectedTitle] = thirdAPIExpectedNumberOfSpecs

		eventsMap := make(map[string]string)
		eventsMap[firstEventTitle] = firstEventDescription
		eventsMap[secondEventTitle] = secondEventDescription
		eventsMap[thirdEventTitle] = thirdEventDescription
		eventsMap[fourthEventTitle] = fourthEventDescription

		publicEventsMap := make(map[string]string)
		publicEventsMap[firstEventTitle] = firstEventDescription
		publicEventsMap[secondEventTitle] = secondEventDescription

		eventsDefaultBundleMap := make(map[string]string)
		eventsDefaultBundleMap[firstEventTitle] = firstBundleOrdIDRegex

		apisAndEventsNumber := make(map[string]int)
		apisAndEventsNumber[apisField] = expectedNumberOfAPIsInFirstBundle + expectedNumberOfAPIsInSecondBundle
		apisAndEventsNumber[publicAPIsField] = expectedNumberOfPublicAPIsInFirstBundle + expectedNumberOfPublicAPIsInSecondBundle
		apisAndEventsNumber[eventsField] = expectedNumberOfEventsInFirstBundle + expectedNumberOfEventsInSecondBundle
		apisAndEventsNumber[publicEventsField] = expectedNumberOfPublicEventsInFirstBundle + expectedNumberOfPublicEventsInSecondBundle

		bundlesMap := make(map[string]string)
		bundlesMap[expectedBundleTitle] = expectedBundleDescription
		bundlesMap[secondExpectedBundleTitle] = secondExpectedBundleDescription

		bundlesAPIsNumberMap := make(map[string]int)
		bundlesAPIsNumberMap[expectedBundleTitle] = expectedNumberOfAPIsInFirstBundle
		bundlesAPIsNumberMap[secondExpectedBundleTitle] = expectedNumberOfAPIsInSecondBundle

		bundlesAPIsData := make(map[string][]string)
		bundlesAPIsData[expectedBundleTitle] = []string{firstAPIExpectedTitle, secondAPIExpectedTitle}
		bundlesAPIsData[secondExpectedBundleTitle] = []string{firstAPIExpectedTitle, thirdAPIExpectedTitle}

		bundlesEventsNumberMap := make(map[string]int)
		bundlesEventsNumberMap[expectedBundleTitle] = expectedNumberOfEventsInFirstBundle
		bundlesEventsNumberMap[secondExpectedBundleTitle] = expectedNumberOfEventsInSecondBundle

		bundlesEventsData := make(map[string][]string)
		bundlesEventsData[expectedBundleTitle] = []string{firstEventTitle, secondEventTitle, thirdEventTitle}
		bundlesEventsData[secondExpectedBundleTitle] = []string{firstEventTitle, secondEventTitle, fourthEventTitle}

		bundlesCorrelationIDs := make(map[string][]string)
		bundlesCorrelationIDs[expectedBundleTitle] = []string{firstCorrelationID, secondCorrelationID}
		bundlesCorrelationIDs[secondExpectedBundleTitle] = []string{firstCorrelationID, secondCorrelationID}

		documentationLabelsPossibleValues := []string{documentationLabelFirstValue, documentationLabelSecondValue}

		productsMap := make(map[string]string)
		productsMap[firstProductTitle] = firstProductShortDescription
		productsMap[secondProductTitle] = secondProductShortDescription

		appTemplateName := createAppTemplateName("ORD-aggregator-test-app-template")
		appTemplateInput := fixAppTemplateInput(appTemplateName, testConfig.ExternalServicesMockUnsecuredMultiTenantURL)
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &appTemplate)
		require.NoError(t, err)
		require.NotEmpty(t, appTemplate)

		selfRegLabelValue, ok := appTemplate.Labels[testConfig.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, testConfig.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTemplate.ID)

		httpClient := &http.Client{
			Timeout: time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: testConfig.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, testConfig.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		require.NoError(t, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(t, http.StatusOK, response.StatusCode)

		subscriptionProviderSubaccountID := testConfig.TestProviderSubaccountID
		subscriptionConsumerSubaccountID := testConfig.TestConsumerSubaccountID
		subscriptionConsumerTenantID := testConfig.TestConsumerTenantID

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
		subscribeReq, err := http.NewRequest(http.MethodPost, testConfig.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, testConfig.SubscriptionConfig.TokenURL+testConfig.TokenPath, testConfig.SubscriptionConfig.ClientID, testConfig.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(subscription.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(subscription.ContentTypeHeader, subscription.ContentTypeApplicationJson)
		subscribeReq.Header.Add(testConfig.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		//unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		//In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, appTemplate.ID, appTemplate.Name, httpClient, testConfig.SubscriptionConfig.URL, apiPath, subscriptionToken, testConfig.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTemplate.Name, appTemplate.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTemplate.ID, appTemplate.Name, httpClient, testConfig.SubscriptionConfig.URL, apiPath, subscriptionToken, testConfig.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		err = resp.Body.Close()
		require.NoError(t, err)

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := testConfig.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTemplate.Name, appTemplate.ID, subscriptionProviderSubaccountID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", "test-int-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys.ID)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSystemCredentials.ID)

		oauthCredentialData, ok := intSystemCredentials.Auth.Credential.(*directorSchema.OAuthCredentialData)
		require.True(t, ok)

		unsecuredHttpClient := http.DefaultClient
		unsecuredHttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		cfgWithInternalVisibilityScope := &clientcredentials.Config{
			ClientID:     oauthCredentialData.ClientID,
			ClientSecret: oauthCredentialData.ClientSecret,
			TokenURL:     oauthCredentialData.URL,
			Scopes:       []string{internalVisibilityScope},
		}

		ctx = context.WithValue(ctx, oauth2.HTTPClient, unsecuredHttpClient)
		httpClient = cfgWithInternalVisibilityScope.Client(ctx)
		httpClient.Timeout = 20 * time.Second

		actualAppPage := directorSchema.ApplicationPage{}
		getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)

		require.NoError(t, err)

		require.Len(t, actualAppPage.Data, 1)
		require.Equal(t, appTemplate.ID, *actualAppPage.Data[0].ApplicationTemplateID)

		scheduleTime, err := parseCronTime(testConfig.AggregatorSchedule)
		require.NoError(t, err)

		defaultTestTimeout := 2*scheduleTime + testTimeoutAdditionalBuffer
		defaultCheckInterval := defaultTestTimeout / 20

		err = verifyORDDocument(defaultCheckInterval, defaultTestTimeout, func() bool {
			var respBody string

			// Verify system instances
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfSystemInstancesInSubscription {
				t.Log("Missing System Instances...will try again")
				return false
			}

			// Verify packages
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfPackagesInSubscription {
				t.Log("Missing Packages...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfSystemInstancesInSubscription)
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfSystemInstancesInSubscription, expectedPackageTitle, expectedPackageDescription, descriptionField)
			t.Log("Successfully verified packages")

			// Verify bundles
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfBundlesInSubscription {
				t.Log("Missing Bundles...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfBundlesInSubscription)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, bundlesMap, expectedNumberOfBundlesInSubscription, descriptionField)
			assertions.AssertBundleCorrelationIds(t, respBody, bundlesCorrelationIDs, expectedNumberOfBundlesInSubscription)
			ordAndInternalIDsMappingForBundles := storeMappingBetweenORDAndInternalBundleID(t, respBody, expectedNumberOfBundlesInSubscription)
			t.Log("Successfully verified bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=apis&$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, apisField, bundlesAPIsNumberMap, bundlesAPIsData)
			t.Log("Successfully verified relation between apis and bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events&$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, eventsField, bundlesEventsNumberMap, bundlesEventsData)
			t.Log("Successfully verified relation between events and bundles")

			globalProductsNumber, globalVendorsNumber := getGlobalResourcesNumber(ctx, t, unsecuredHttpClient)
			t.Logf("Global products number: %d, Global vendors number: %d", globalProductsNumber, globalVendorsNumber)

			expectedTotalNumberOfProducts := expectedNumberOfProductsInSubscription + globalProductsNumber
			expectedTotalNumberOfVendors := expectedNumberOfVendorsInSubscription + globalVendorsNumber

			// Verify products
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/products?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedTotalNumberOfProducts {
				t.Log("Missing Products...will try again")
				return false
			}
			t.Logf("Expected total number of product: %d", expectedTotalNumberOfProducts)
			t.Logf("Products response body: %s", respBody)
			t.Logf("Expected products map: %v", productsMap)
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedTotalNumberOfProducts)
			assertions.AssertProducts(t, respBody, productsMap, expectedTotalNumberOfProducts, shortDescriptionField)
			t.Log("Successfully verified products")

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfAPIsInSubscription {
				t.Log("Missing APIs...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfAPIsInSubscription)
			// In the document there are actually 4 APIs but there is a tombstone for one of them so in the end there will be 3 APIs
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, apisMap, expectedNumberOfAPIsInSubscription, descriptionField)
			t.Log("Successfully verified apis")

			// Verify defaultBundle for apis
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfAPIsInSubscription, apisDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for apis")

			// Verify the api spec
			specs := assertions.AssertSpecsFromORDService(t, respBody, expectedNumberOfAPIsInSubscription, apiSpecsMap)
			t.Log("Successfully verified specs for apis")

			var specURL string
			for _, s := range specs {
				specType := s.Get("type").String()
				specFormat := s.Get("mediaType").String()
				if specType == expectedSpecType && specFormat == expectedSpecFormat {
					specURL = s.Get("url").String()
					break
				}
			}

			respBody = makeRequestWithHeaders(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(respBody) == 0 || !strings.Contains(respBody, "swagger") {
				t.Logf("Spec %s not successfully fetched... will try again", specURL)
				return false
			}
			t.Log("Successfully verified api spec")

			// Verify events
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEventsInSubscription {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfEventsInSubscription)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, eventsMap, expectedNumberOfEventsInSubscription, descriptionField)
			t.Log("Successfully verified events")

			// Verify defaultBundle for events
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfEventsInSubscription, eventsDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for events")

			// Verify tombstones
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/tombstones?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfTombstonesInSubscription {
				t.Log("Missing Tombstones...will try again")
				return false
			}
			assertions.AssertTombstoneFromORDService(t, respBody, expectedNumberOfTombstonesInSubscription, expectedTombstoneOrdIDRegex)
			t.Log("Successfully verified tombstones")

			// Verify vendors
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})

			if len(gjson.Get(respBody, "value").Array()) < expectedTotalNumberOfVendors {
				t.Log("Missing Vendors...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedTotalNumberOfVendors)
			assertions.AssertVendorFromORDService(t, respBody, expectedTotalNumberOfVendors, expectedNumberOfVendorsInSubscription, expectedVendorTitle)
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

func verifyEntitiesWithPublicVisibilityInORD(t *testing.T, httpClient *http.Client, publicEntitiesMap map[string]string, entity string, expectedNumberOfPublicEntities int) {
	respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+fmt.Sprintf("/%s?$format=json", entity), map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

	assertions.AssertMultipleEntitiesFromORDService(t, respBody, publicEntitiesMap, expectedNumberOfPublicEntities, descriptionField)
	t.Logf("Successfully verified public %s", entity)
}

func verifyEntitiesVisibilityViaGraphql(t *testing.T, clientWithInternalScope, clientWithoutInternalScope *gcli.Client, entitiesMap, publicEntitiesMap map[string]string, expectedNumberOfEntities map[string]int, appID string) {
	appWithAllEntities := fixtures.GetApplication(t, context.Background(), clientWithInternalScope, testConfig.DefaultTestTenant, appID)
	appWithPublicEntities := fixtures.GetApplication(t, context.Background(), clientWithoutInternalScope, testConfig.DefaultTestTenant, appID)

	var allAPIs []*directorSchema.APIDefinitionExt
	var allEvents []*directorSchema.EventAPIDefinitionExt

	for _, bndl := range appWithAllEntities.Bundles.Data {
		allAPIs = append(allAPIs, bndl.APIDefinitions.Data...)
		allEvents = append(allEvents, bndl.EventDefinitions.Data...)
	}

	var publicAPIs []*directorSchema.APIDefinitionExt
	var publicEvents []*directorSchema.EventAPIDefinitionExt

	for _, bndl := range appWithPublicEntities.Bundles.Data {
		publicAPIs = append(publicAPIs, bndl.APIDefinitions.Data...)
		publicEvents = append(publicEvents, bndl.EventDefinitions.Data...)
	}

	t.Log("Start verifying all APIs via Director graphql api")
	for _, api := range allAPIs {
		require.Equal(t, entitiesMap[api.Name], *api.Description)
	}
	require.Equal(t, len(allAPIs), expectedNumberOfEntities[apisField])
	t.Log("Successfully verified all APIs via Director graphql api")

	t.Log("Start verifying public APIs via Director graphql api")
	for _, api := range publicAPIs {
		require.Equal(t, publicEntitiesMap[api.Name], *api.Description)
	}
	require.Equal(t, len(publicAPIs), expectedNumberOfEntities[publicAPIsField])
	t.Log("Successfully verified public APIs via Director graphql api")

	t.Log("Start verifying all Events via Director graphql api")
	for _, event := range allEvents {
		require.Equal(t, entitiesMap[event.Name], *event.Description)
	}
	require.Equal(t, len(allEvents), expectedNumberOfEntities[eventsField])
	t.Log("Successfully verified all Events via Director graphql api")

	t.Log("Start verifying public Events via Director graphql api")
	for _, event := range publicEvents {
		require.Equal(t, publicEntitiesMap[event.Name], *event.Description)
	}
	require.Equal(t, len(publicEvents), expectedNumberOfEntities[publicEventsField])
	t.Log("Successfully verified public Events via Director graphql api")

}

func getGlobalResourcesNumber(ctx context.Context, t *testing.T, httpClient *http.Client) (int, int) {
	accessStrategyExecutorProvider := accessstrategy.NewDefaultExecutorProvider(certCache)
	ordClient := NewGlobalRegistryClient(httpClient, accessStrategyExecutorProvider)

	products, vendors, err := ordClient.GetGlobalProductsAndVendorsNumber(ctx, testConfig.GlobalRegistryURL)
	if err != nil {
		t.Fatalf("while fetching global registry resources from %s %v", testConfig.GlobalRegistryURL, err)
	}
	return products, vendors
}

func mergeMaps(first, second map[string]string) map[string]string {
	for k, v := range second {
		first[k] = v
	}
	return first
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}

func fixAppTemplateInput(name, webhookURL string) directorSchema.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplateWithORDWebhook(name, webhookURL)
	input.Labels[testConfig.SubscriptionConfig.SelfRegDistinguishLabelKey] = testConfig.SubscriptionConfig.SelfRegDistinguishLabelValue

	return input
}
