package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"

	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"golang.org/x/oauth2"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/util"

	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/request"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	tenantHeader            = "Tenant"
	internalVisibilityScope = "internal_visibility:read"

	descriptionField             = "description"
	shortDescriptionField        = "shortDescription"
	apisField                    = "apis"
	eventsField                  = "events"
	capabilitiesField            = "capabilities"
	integrationDependenciesField = "integrationDependencies"
	publicAPIsField              = "publicAPIs"
	publicEventsField            = "publicEvents"

	expectedSpecType                         = "openapi-v3"
	expectedCapabilitySpecType               = "sap.mdo:mdi-capability-definition:v1"
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
	expectedEntityTypeTitle                  = "ENTITYTYPE 1 TITLE"
	expectedEntityTypeDescription            = "lorem ipsum dolor set"
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
	expectedCapabilityTitle                  = "CAPABILITY TITLE"
	expectedCapabilityDescription            = "Optional, longer description"
	expectedCapabilityNumberOfSpecs          = 1
	expectedIntegrationDependencyTitle       = "INTEGRATION DEPENDENCY TITLE"
	expectedIntegrationDependencyDescription = "longer description of an integration dependency"
	expectedTombstoneOrdIDRegex              = "ns:apiResource:API_ID2(.+):v1"
	expectedVendorTitle                      = "SAP SE"

	expectedNumberOfSystemInstances                       = 7
	expectedNumberOfSystemInstancesInSubscription         = 1
	expectedNumberOfPackages                              = 7
	expectedNumberOfPackagesInSubscription                = 1
	expectedNumberOfEntityTypes                           = 7
	expectedNumberOfEntityTypesInSubscription             = 1
	expectedNumberOfBundles                               = 14
	expectedNumberOfBundlesInSubscription                 = 2
	expectedNumberOfAPIs                                  = 21
	expectedNumberOfAPIsInSubscription                    = 3
	expectedNumberOfEvents                                = 28
	expectedNumberOfEventsInSubscription                  = 4
	expectedNumberOfCapabilities                          = 7
	expectedNumberOfCapabilitiesInSubscription            = 1
	expectedNumberOfIntegrationDependencies               = 7
	expectedNumberOfIntegrationDependenciesInSubscription = 1
	expectedNumberOfTombstones                            = 7
	expectedNumberOfTombstonesInSubscription              = 1

	expectedNumberOfPublicAPIs                    = 7
	expectedNumberOfPublicEvents                  = 14
	expectedNumberOfPublicCapabilities            = 7
	expectedNumberOfPublicIntegrationDependencies = 7

	expectedNumberOfAPIsInFirstBundle    = 2
	expectedNumberOfAPIsInSecondBundle   = 2
	expectedNumberOfEventsInFirstBundle  = 3
	expectedNumberOfEventsInSecondBundle = 3

	expectedNumberOfPublicAPIsInFirstBundle    = 1
	expectedNumberOfPublicAPIsInSecondBundle   = 1
	expectedNumberOfPublicEventsInFirstBundle  = 2
	expectedNumberOfPublicEventsInSecondBundle = 2

	firstCorrelationID  = "sap.s4:communicationScenario:SAP_COM_0001"
	secondCorrelationID = "sap.s4:communicationScenario:SAP_COM_0002"

	documentationLabelKey         = "Documentation label key"
	documentationLabelFirstValue  = "Markdown Documentation with links"
	documentationLabelSecondValue = "With multiple values"

	apiResourceDefinitionsFieldName        = "resourceDefinitions"
	capabilityResourceDefinitionsFieldName = "definitions"
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

		capabilitiesMap := make(map[string]string)
		capabilitiesMap[expectedCapabilityTitle] = expectedCapabilityDescription

		publicCapabilitiesMap := make(map[string]string)
		publicCapabilitiesMap[expectedCapabilityTitle] = expectedCapabilityDescription

		capabilitySpecsMap := make(map[string]int)
		capabilitySpecsMap[expectedCapabilityTitle] = expectedCapabilityNumberOfSpecs

		integrationDependenciesMap := make(map[string]string)
		integrationDependenciesMap[expectedIntegrationDependencyTitle] = expectedIntegrationDependencyDescription

		publicIntegrationDependenciesMap := make(map[string]string)
		publicIntegrationDependenciesMap[expectedIntegrationDependencyTitle] = expectedIntegrationDependencyDescription

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

		defaultTestTimeout := 5 * time.Minute
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

			// Verify entity types
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/entityTypes?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEntityTypes {
				t.Log("Missing Entity Types...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfEntityTypes)
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfEntityTypes, expectedEntityTypeTitle, expectedEntityTypeDescription, descriptionField)
			t.Log("Successfully verified EntityTypes")

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
			if !assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, apisField, bundlesAPIsNumberMap, bundlesAPIsData) {
				t.Logf("Relation between bundles and %s does not match..will try again", apisField)
				return false
			}
			t.Log("Successfully verified relation between apis and bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if !assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, eventsField, bundlesEventsNumberMap, bundlesEventsData) {
				t.Logf("Relation between bundles and %s does not match..will try again", eventsField)
				return false
			}
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
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$expand=entityTypeMappings&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfAPIs {
				t.Log("Missing APIs...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfAPIs)
			// In the document there are actually 4 APIs but there is a tombstone for one of them so in the end there will be 3 APIs
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, apisMap, expectedNumberOfAPIs, descriptionField)
			t.Log("Successfully verified apis")

			// Verify EntityTypeMappings
			apiTitlesWithEntityTypeMappingsCountToCheck := map[string]int{}
			apiTitlesWithEntityTypeMappingsExpectedContent := map[string]string{}
			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE"] = 1
			apiTitlesWithEntityTypeMappingsExpectedContent["API TITLE"] = "A_OperationalAcctgDocItemCube"

			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE INTERNAL"] = 2
			apiTitlesWithEntityTypeMappingsExpectedContent["API TITLE INTERNAL"] = "B_OperationalAcctgDocItemCube"

			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE PRIVATE"] = 0
			assertions.AssertEntityTypeMappings(t, respBody, expectedNumberOfAPIsInSubscription, apiTitlesWithEntityTypeMappingsCountToCheck, apiTitlesWithEntityTypeMappingsExpectedContent)
			t.Log("Successfully verified api entity type mappings")

			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfAPIs, apisDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for apis")

			// Verify the api spec
			specs := assertions.AssertSpecsFromORDService(t, respBody, expectedNumberOfAPIs, apiSpecsMap, apiResourceDefinitionsFieldName)
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
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$expand=entityTypeMappings&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})

			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEvents {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfEvents)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, eventsMap, expectedNumberOfEvents, descriptionField)
			t.Log("Successfully verified events")

			// Verify EntityTypeMappings
			eventTitlesWithEntityTypeMappingsCountToCheck := map[string]int{}
			eventTitlesWithEntityTypeMappingsExpectedContent := map[string]string{}
			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE"] = 1
			eventTitlesWithEntityTypeMappingsExpectedContent["EVENT TITLE"] = "sap.odm:entityType:CostCenter:v1"

			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE INTERNAL"] = 2
			eventTitlesWithEntityTypeMappingsExpectedContent["EVENT TITLE INTERNAL"] = "sap.odm:entityType:CostCenter:v2"

			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE PRIVATE"] = 0
			assertions.AssertEntityTypeMappings(t, respBody, expectedNumberOfEventsInSubscription, eventTitlesWithEntityTypeMappingsCountToCheck, eventTitlesWithEntityTypeMappingsExpectedContent)
			t.Log("Successfully verified api entity type mappings")

			// Verify defaultBundle for events
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfEvents, eventsDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for events")

			// verify public events via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicEventsMap, eventsField, expectedNumberOfPublicEvents)

			// verify apis and events visibility via Director's graphql
			verifyEntitiesVisibilityViaGraphql(t, oauthGraphQLClientWithInternalVisibility, oauthGraphQLClientWithoutInternalVisibility, mergeMaps(apisMap, eventsMap), mergeMaps(publicApisMap, publicEventsMap), apisAndEventsNumber, app.ID)

			// Verify capabilities
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/capabilities?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfCapabilities {
				t.Log("Missing Capabilities...will try again")
				return false
			}

			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfCapabilities)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, capabilitiesMap, expectedNumberOfCapabilities, descriptionField)
			t.Log("Successfully verified capabilities")

			// Verify the capability spec
			capabilitySpecs := assertions.AssertSpecsFromORDService(t, respBody, expectedNumberOfCapabilities, capabilitySpecsMap, capabilityResourceDefinitionsFieldName)
			t.Log("Successfully verified specs for capabilities")

			var capabilitySpecURL string
			for _, s := range capabilitySpecs {
				specType := s.Get("type").String()
				specFormat := s.Get("mediaType").String()
				if specType == expectedCapabilitySpecType && specFormat == expectedSpecFormat {
					capabilitySpecURL = s.Get("url").String()
					break
				}
			}

			respBody = makeRequestWithHeaders(t, httpClient, capabilitySpecURL, map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(respBody) == 0 || !strings.Contains(respBody, "swagger") {
				t.Logf("Spec %s not successfully fetched... will try again", specURL)
				return false
			}
			t.Log("Successfully verified capability spec")

			// verify public capabilities via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicCapabilitiesMap, capabilitiesField, expectedNumberOfPublicCapabilities)

			// Verify integration dependencies
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/integrationDependencies?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfIntegrationDependencies {
				t.Log("Missing Integration Dependencies...will try again")
				return false
			}

			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfIntegrationDependencies)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, integrationDependenciesMap, expectedNumberOfIntegrationDependencies, descriptionField)
			t.Log("Successfully verified integration dependencies")

			// verify public integration dependencies via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicIntegrationDependenciesMap, integrationDependenciesField, expectedNumberOfPublicIntegrationDependencies)

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
		t.Log("Successfully verified all ORD documents")
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

		capabilitiesMap := make(map[string]string)
		capabilitiesMap[expectedCapabilityTitle] = expectedCapabilityDescription

		publicCapabilitiesMap := make(map[string]string)
		publicCapabilitiesMap[expectedCapabilityTitle] = expectedCapabilityDescription

		capabilitySpecsMap := make(map[string]int)
		capabilitySpecsMap[expectedCapabilityTitle] = expectedCapabilityNumberOfSpecs

		integrationDependenciesMap := make(map[string]string)
		integrationDependenciesMap[expectedIntegrationDependencyTitle] = expectedIntegrationDependencyDescription

		publicIntegrationDependenciesMap := make(map[string]string)
		publicIntegrationDependenciesMap[expectedIntegrationDependencyTitle] = expectedIntegrationDependencyDescription

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

		appTemplateName := fixtures.CreateAppTemplateName("ORD-aggregator-test-app-template")
		appTemplateInput := fixAppTemplateInputWitSelfRegLabel(appTemplateName, testConfig.ExternalServicesMockUnsecuredMultiTenantURL)
		placeholderName := "name"
		placeholderDisplayName := "display-name"
		appTemplateInput.Placeholders = []*directorSchema.PlaceholderDefinitionInput{
			{
				Name:        "name",
				Description: &placeholderName,
				JSONPath:    str.Ptr(fmt.Sprintf("$.%s", testConfig.SubscriptionProviderAppNameProperty)),
			},
			{
				Name:        "display-name",
				Description: &placeholderDisplayName,
				JSONPath:    str.Ptr(fmt.Sprintf("$.%s", testConfig.SubscriptionProviderAppNameProperty)),
			},
		}
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestSubaccount, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestSubaccount, appTemplate)
		require.NoError(t, err)
		require.NotEmpty(t, appTemplate)

		selfRegLabelValue, ok := appTemplate.Labels[testConfig.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, testConfig.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTemplate.ID)

		httpClient := &http.Client{
			Timeout: 2 * time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: testConfig.SkipSSLValidation},
			},
		}

		deps, err := json.Marshal([]string{selfRegLabelValue})
		require.NoError(t, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, testConfig.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		subscriptionProviderSubaccountID := testConfig.TestProviderSubaccountID
		subscriptionConsumerSubaccountID := testConfig.TestConsumerSubaccountID

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", testConfig.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, testConfig.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, testConfig.SubscriptionConfig.TokenURL+testConfig.TokenPath, testConfig.SubscriptionConfig.ClientID, testConfig.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(testConfig.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)
		subscribeReq.Header.Add(testConfig.SubscriptionConfig.SubscriptionFlowHeaderKey, testConfig.SubscriptionConfig.StandardFlow)
		//unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		//In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, appTemplate.ID, appTemplate.Name, httpClient, testConfig.SubscriptionConfig.URL, apiPath, subscriptionToken, testConfig.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, testConfig.SubscriptionConfig.StandardFlow, testConfig.SubscriptionConfig.SubscriptionFlowHeaderKey)

		t.Logf("Creating a subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, appTemplate.Name, appTemplate.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTemplate.ID, appTemplate.Name, httpClient, testConfig.SubscriptionConfig.URL, apiPath, subscriptionToken, testConfig.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, testConfig.SubscriptionConfig.StandardFlow, testConfig.SubscriptionConfig.SubscriptionFlowHeaderKey)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
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
		t.Logf("Successfully created subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, appTemplate.Name, appTemplate.ID, subscriptionProviderSubaccountID)

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

		defaultTestTimeout := 5 * time.Minute
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
			if !assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, apisField, bundlesAPIsNumberMap, bundlesAPIsData) {
				t.Logf("Relation between bundles and %s does not match..will try again", apisField)
				return false
			}
			t.Log("Successfully verified relation between apis and bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events&$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if !assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, eventsField, bundlesEventsNumberMap, bundlesEventsData) {
				t.Logf("Relation between bundles and %s does not match..will try again", eventsField)
				return false
			}
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
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedTotalNumberOfProducts)
			assertions.AssertProducts(t, respBody, productsMap, expectedTotalNumberOfProducts, shortDescriptionField)
			t.Log("Successfully verified products")

			// Verify apis
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$expand=entityTypeMappings&$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
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

			// Verify EntityTypeMappings
			apiTitlesWithEntityTypeMappingsCountToCheck := map[string]int{}
			apiTitlesWithEntityTypeMappingsExpectedContent := map[string]string{}
			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE"] = 1
			apiTitlesWithEntityTypeMappingsExpectedContent["API TITLE"] = "A_OperationalAcctgDocItemCube"

			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE INTERNAL"] = 2
			apiTitlesWithEntityTypeMappingsExpectedContent["API TITLE INTERNAL"] = "B_OperationalAcctgDocItemCube"

			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE PRIVATE"] = 0
			assertions.AssertEntityTypeMappings(t, respBody, expectedNumberOfAPIsInSubscription, apiTitlesWithEntityTypeMappingsCountToCheck, apiTitlesWithEntityTypeMappingsExpectedContent)
			t.Log("Successfully verified api entity type mappings")

			// Verify the api spec
			specs := assertions.AssertSpecsFromORDService(t, respBody, expectedNumberOfAPIsInSubscription, apiSpecsMap, apiResourceDefinitionsFieldName)
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
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$expand=entityTypeMappings&$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEventsInSubscription {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfEventsInSubscription)

			// Verify EntityTypeMappings
			eventTitlesWithEntityTypeMappingsCountToCheck := map[string]int{}
			eventTitlesWithEntityTypeMappingsExpectedContent := map[string]string{}
			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE"] = 1
			eventTitlesWithEntityTypeMappingsExpectedContent["EVENT TITLE"] = "sap.odm:entityType:CostCenter:v1"

			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE INTERNAL"] = 2
			eventTitlesWithEntityTypeMappingsExpectedContent["EVENT TITLE INTERNAL"] = "sap.odm:entityType:CostCenter:v2"

			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE PRIVATE"] = 0
			assertions.AssertEntityTypeMappings(t, respBody, expectedNumberOfEventsInSubscription, eventTitlesWithEntityTypeMappingsCountToCheck, eventTitlesWithEntityTypeMappingsExpectedContent)
			t.Log("Successfully verified event entity type mappings")

			assertions.AssertMultipleEntitiesFromORDService(t, respBody, eventsMap, expectedNumberOfEventsInSubscription, descriptionField)
			t.Log("Successfully verified events")

			// Verify entity types
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/entityTypes?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfEntityTypesInSubscription {
				t.Log("Missing Entity Types...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfEntityTypesInSubscription)
			assertions.AssertSingleEntityFromORDService(t, respBody, expectedNumberOfEntityTypesInSubscription, expectedEntityTypeTitle, expectedEntityTypeDescription, descriptionField)
			t.Log("Successfully verified EntityTypes")

			// Verify defaultBundle for events
			assertions.AssertDefaultBundleID(t, respBody, expectedNumberOfEventsInSubscription, eventsDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for events")

			// Verify capabilities
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/capabilities?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfCapabilitiesInSubscription {
				t.Log("Missing Capabilities...will try again")
				return false
			}

			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfCapabilitiesInSubscription)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, capabilitiesMap, expectedNumberOfCapabilitiesInSubscription, descriptionField)
			t.Log("Successfully verified capabilities")

			// Verify the capability spec
			capabilitySpecs := assertions.AssertSpecsFromORDService(t, respBody, expectedNumberOfCapabilitiesInSubscription, capabilitySpecsMap, capabilityResourceDefinitionsFieldName)
			t.Log("Successfully verified specs for capabilities")

			var capabilitySpecURL string
			for _, s := range capabilitySpecs {
				specType := s.Get("type").String()
				specFormat := s.Get("mediaType").String()
				if specType == expectedCapabilitySpecType && specFormat == expectedSpecFormat {
					capabilitySpecURL = s.Get("url").String()
					break
				}
			}

			respBody = makeRequestWithHeaders(t, httpClient, capabilitySpecURL, map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(respBody) == 0 || !strings.Contains(respBody, "swagger") {
				t.Logf("Spec %s not successfully fetched... will try again", specURL)
				return false
			}
			t.Log("Successfully verified capability spec")

			// Verify integration dependencies
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/integrationDependencies?$format=json", map[string][]string{tenantHeader: {testConfig.TestConsumerSubaccountID}})
			if len(gjson.Get(respBody, "value").Array()) < expectedNumberOfIntegrationDependenciesInSubscription {
				t.Log("Missing Integration Dependencies...will try again")
				return false
			}

			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedNumberOfIntegrationDependenciesInSubscription)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, integrationDependenciesMap, expectedNumberOfIntegrationDependenciesInSubscription, descriptionField)
			t.Log("Successfully verified integration dependencies")

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
		t.Log("Successfully verified all ORD documents")
	})
	t.Run("Verify ORD document that is accessible through a proxy", func(t *testing.T) {
		numberOfProducts := 1
		numberOfVendors := 1
		numberOfSystemInstances := 1
		numberOfPackages := 1
		numberOfBundles := 2
		numberOfAPIs := 3
		numberOfPublicAPIs := 1
		numberOfEvents := 4
		numberOfEntityTypes := 1
		numberOfPublicEvents := 2
		numberOfCapabilities := 1
		numberOfPublicCapabilities := 1
		numberOfIntegrationDependencies := 1
		numberOfPublicIntegrationDependencies := 1
		numberOfTombstones := 1

		ctx := context.Background()

		// Create integration system for credentials
		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", "test-int-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys.ID)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSystemCredentials.ID)

		// Create Application Template
		appTemplateInput := fixtures.FixApplicationTemplate(testConfig.ProxyApplicationTemplateName)

		appTemplate, err := fixtures.CreateApplicationTemplateFromInputWithoutTenant(t, ctx, certSecuredGraphQLClient, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, "", appTemplate)
		require.NoError(t, err)
		require.NotEmpty(t, appTemplate)

		// Create Application from Template
		app := fixtures.RegisterApplicationFromTemplate(t, ctx, certSecuredGraphQLClient, testConfig.ProxyApplicationTemplateName, expectedSystemInstanceName, expectedSystemInstanceName, testConfig.DefaultTestTenant)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)

		// Update Application to simulate successful Pairing flow and to add an ORD webhook from the mappings in Director
		updatedDescription := "appUpdated"
		updateAppInput := fixtures.FixSampleApplicationUpdateInput(updatedDescription)
		updateAppInput.BaseURL = str.Ptr("http://test.com/test/v1")
		_, err = fixtures.UpdateApplicationWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, app.ID, updateAppInput)
		require.NoError(t, err)

		// Define assertion data
		systemInstancesMap := make(map[string]string)
		systemInstancesMap[expectedSystemInstanceName] = updatedDescription

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

		capabilitiesMap := make(map[string]string)
		capabilitiesMap[expectedCapabilityTitle] = expectedCapabilityDescription

		publicCapabilitiesMap := make(map[string]string)
		publicCapabilitiesMap[expectedCapabilityTitle] = expectedCapabilityDescription

		capabilitySpecsMap := make(map[string]int)
		capabilitySpecsMap[expectedCapabilityTitle] = expectedCapabilityNumberOfSpecs

		integrationDependenciesMap := make(map[string]string)
		integrationDependenciesMap[expectedIntegrationDependencyTitle] = expectedIntegrationDependencyDescription

		publicIntegrationDependenciesMap := make(map[string]string)
		publicIntegrationDependenciesMap[expectedIntegrationDependencyTitle] = expectedIntegrationDependencyDescription

		entityTypesMap := make(map[string]string)
		entityTypesMap[expectedEntityTypeTitle] = expectedEntityTypeDescription

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

		expectedTotalNumberOfProducts := numberOfProducts + globalProductsNumber
		expectedTotalNumberOfVendors := numberOfVendors + globalVendorsNumber

		defaultTestTimeout := 5 * time.Minute
		defaultCheckInterval := defaultTestTimeout / 20

		err = verifyORDDocument(defaultCheckInterval, defaultTestTimeout, func() bool {
			var respBody string

			// Verify system instances
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfSystemInstances {
				t.Log("Missing System Instances...will try again")
				return false
			}
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, systemInstancesMap, numberOfSystemInstances, descriptionField)

			// Verify packages
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < 1 {
				t.Log("Missing Packages...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, numberOfPackages)
			assertions.AssertSingleEntityFromORDService(t, respBody, numberOfPackages, expectedPackageTitle, expectedPackageDescription, descriptionField)
			t.Log("Successfully verified packages")

			// Verify bundles
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfBundles {
				t.Log("Missing Bundles...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, 2)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, bundlesMap, numberOfBundles, descriptionField)
			assertions.AssertBundleCorrelationIds(t, respBody, bundlesCorrelationIDs, numberOfBundles)
			ordAndInternalIDsMappingForBundles := storeMappingBetweenORDAndInternalBundleID(t, respBody, 2)
			t.Log("Successfully verified bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=apis&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if !assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, apisField, bundlesAPIsNumberMap, bundlesAPIsData) {
				t.Logf("Relation between bundles and %s does not match..will try again", apisField)
				return false
			}
			t.Log("Successfully verified relation between apis and bundles")

			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if !assertions.AssertRelationBetweenBundleAndEntityFromORDService(t, respBody, eventsField, bundlesEventsNumberMap, bundlesEventsData) {
				t.Logf("Relation between bundles and %s does not match..will try again", eventsField)
				return false
			}
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
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$expand=entityTypeMappings&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < 3 {
				t.Log("Missing APIs...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, numberOfAPIs)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, apisMap, numberOfAPIs, descriptionField)
			t.Log("Successfully verified apis")

			// Verify EntityTypeMappings
			apiTitlesWithEntityTypeMappingsCountToCheck := map[string]int{}
			apiTitlesWithEntityTypeMappingsExpectedContent := map[string]string{}
			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE"] = 1
			apiTitlesWithEntityTypeMappingsExpectedContent["API TITLE"] = "A_OperationalAcctgDocItemCube"

			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE INTERNAL"] = 2
			apiTitlesWithEntityTypeMappingsExpectedContent["API TITLE INTERNAL"] = "B_OperationalAcctgDocItemCube"

			apiTitlesWithEntityTypeMappingsCountToCheck["API TITLE PRIVATE"] = 0
			assertions.AssertEntityTypeMappings(t, respBody, expectedNumberOfAPIsInSubscription, apiTitlesWithEntityTypeMappingsCountToCheck, apiTitlesWithEntityTypeMappingsExpectedContent)
			t.Log("Successfully verified api entity type mappings")

			// Verify defaultBundle for apis
			assertions.AssertDefaultBundleID(t, respBody, numberOfAPIs, apisDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for apis")

			// Verify the api spec
			specs := assertions.AssertSpecsFromORDService(t, respBody, numberOfAPIs, apiSpecsMap, apiResourceDefinitionsFieldName)
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
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicApisMap, apisField, numberOfPublicAPIs)

			// Verify events
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$expand=entityTypeMappings&$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfEvents {
				t.Log("Missing Events...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, numberOfEvents)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, eventsMap, numberOfEvents, descriptionField)
			t.Log("Successfully verified events")

			// Verify EntityTypeMappings
			eventTitlesWithEntityTypeMappingsCountToCheck := map[string]int{}
			eventTitlesWithEntityTypeMappingsExpectedContent := map[string]string{}
			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE"] = 1
			eventTitlesWithEntityTypeMappingsExpectedContent["EVENT TITLE"] = "sap.odm:entityType:CostCenter:v1"

			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE INTERNAL"] = 2
			eventTitlesWithEntityTypeMappingsExpectedContent["EVENT TITLE INTERNAL"] = "sap.odm:entityType:CostCenter:v2"

			eventTitlesWithEntityTypeMappingsCountToCheck["EVENT TITLE PRIVATE"] = 0
			assertions.AssertEntityTypeMappings(t, respBody, numberOfEvents, eventTitlesWithEntityTypeMappingsCountToCheck, eventTitlesWithEntityTypeMappingsExpectedContent)
			t.Log("Successfully verified api entity type mappings")

			// Verify defaultBundle for events
			assertions.AssertDefaultBundleID(t, respBody, numberOfEvents, eventsDefaultBundleMap, ordAndInternalIDsMappingForBundles)
			t.Log("Successfully verified defaultBundles for events")

			// verify public events via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicEventsMap, eventsField, numberOfPublicEvents)

			// verify apis and events visibility via Director's graphql
			verifyEntitiesVisibilityViaGraphql(t, oauthGraphQLClientWithInternalVisibility, oauthGraphQLClientWithoutInternalVisibility, mergeMaps(apisMap, eventsMap), mergeMaps(publicApisMap, publicEventsMap), apisAndEventsNumber, app.ID)

			// Verify entity types
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/entityTypes?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfEntityTypes {
				t.Log("Missing Entity Types...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, numberOfEntityTypes)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, entityTypesMap, numberOfEntityTypes, descriptionField)
			t.Log("Successfully verified EntityTypes")

			// Verify capabilities
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/capabilities?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfCapabilities {
				t.Log("Missing Capabilities...will try again")
				return false
			}

			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, numberOfCapabilities)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, capabilitiesMap, numberOfCapabilities, descriptionField)
			t.Log("Successfully verified capabilities")

			// Verify the capability spec
			capabilitySpecs := assertions.AssertSpecsFromORDService(t, respBody, numberOfCapabilities, capabilitySpecsMap, capabilityResourceDefinitionsFieldName)
			t.Log("Successfully verified specs for capabilities")

			var capabilitySpecURL string
			for _, s := range capabilitySpecs {
				specType := s.Get("type").String()
				specFormat := s.Get("mediaType").String()
				if specType == expectedCapabilitySpecType && specFormat == expectedSpecFormat {
					capabilitySpecURL = s.Get("url").String()
					break
				}
			}

			respBody = makeRequestWithHeaders(t, httpClient, capabilitySpecURL, map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(respBody) == 0 || !strings.Contains(respBody, "swagger") {
				t.Logf("Spec %s not successfully fetched... will try again", specURL)
				return false
			}
			t.Log("Successfully verified capability spec")

			// verify public capabilities via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicCapabilitiesMap, capabilitiesField, numberOfPublicCapabilities)

			// Verify integration dependencies
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/integrationDependencies?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfIntegrationDependencies {
				t.Log("Missing Integration Dependencies...will try again")
				return false
			}

			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, numberOfIntegrationDependencies)
			assertions.AssertMultipleEntitiesFromORDService(t, respBody, integrationDependenciesMap, numberOfIntegrationDependencies, descriptionField)
			t.Log("Successfully verified integration dependencies")

			// verify public integration dependencies via ORD Service
			verifyEntitiesWithPublicVisibilityInORD(t, httpClientWithoutVisibilityScope, publicIntegrationDependenciesMap, integrationDependenciesField, numberOfPublicIntegrationDependencies)

			// Verify tombstones
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/tombstones?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < numberOfTombstones {
				t.Log("Missing Tombstones...will try again")
				return false
			}
			assertions.AssertTombstoneFromORDService(t, respBody, numberOfTombstones, expectedTombstoneOrdIDRegex)
			t.Log("Successfully verified tombstones")

			// Verify vendors
			respBody = makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/vendors?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			if len(gjson.Get(respBody, "value").Array()) < expectedTotalNumberOfVendors {
				t.Log("Missing Vendors...will try again")
				return false
			}
			assertions.AssertDocumentationLabels(t, respBody, documentationLabelKey, documentationLabelsPossibleValues, expectedTotalNumberOfVendors)
			assertions.AssertVendorFromORDService(t, respBody, expectedTotalNumberOfVendors, numberOfVendors, expectedVendorTitle)
			t.Log("Successfully verified vendors")

			return true
		})
		require.NoError(t, err)
		t.Log("Successfully verified all ORD documents")
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
	accessStrategyExecutorProvider := accessstrategy.NewDefaultExecutorProvider(certCache, testConfig.ExternalClientCertSecretName, testConfig.ExtSvcClientCertSecretName)
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

func fixAppTemplateInput(name, webhookURL string) directorSchema.ApplicationTemplateInput {
	return fixtures.FixApplicationTemplateWithORDWebhook(name, webhookURL)
}

func fixAppTemplateInputWitSelfRegLabel(name, webhookURL string) directorSchema.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplateWithORDWebhook(name, webhookURL)
	input.Labels[testConfig.SubscriptionConfig.SelfRegDistinguishLabelKey] = testConfig.SubscriptionConfig.SelfRegDistinguishLabelValue

	return input
}
