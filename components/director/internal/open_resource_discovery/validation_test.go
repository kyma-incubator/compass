package ord_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

const (
	invalidOpenResourceDiscovery  = "invalidOpenResourceDiscovery"
	invalidURL                    = "invalidURL"
	invalidOrdID                  = "invalidOrdId"
	invalidShortDescriptionLength = 257 // max allowed: 256
	maxDescriptionLength          = 5000
	invalidVersion                = "invalidVersion"
	invalidPolicyLevel            = "invalidPolicyLevel"
	invalidVendor                 = "wrongVendor!"
	invalidType                   = "invalidType"
	invalidCustomType             = "wrongCustomType"
	invalidMediaType              = "invalid/type"
	invalidBundleOrdID            = "ns:wrongConsumptionBundle:v1"

	unknownVendorOrdID  = "nsUNKNOWN:vendor:id:"
	unknownProductOrdID = "nsUNKNOWN:product:id:"
	unknownPackageOrdID = "ns:package:UNKNOWN_PACKAGE_ID:v1"
	unknownBundleOrdID  = "ns:consumptionBundle:UNKNOWN_BUNDLE_ID:v1"
)

var (
	invalidJSON = `[
        {
          foo: bar,
        }
      ]`

	invalidPackageLinkDueToMissingType = `[
        {
          "url": "https://example.com/en/legal/terms-of-use.html"
        },
        {
          "type": "client-registration",
          "url": "https://example2.com/en/legal/terms-of-use.html"
        }
      ]`

	invalidPackageLinkDueToWrongType = `[
        {
          "type": "wrongType",
          "url": "https://example.com/en/legal/terms-of-use.html"
        },
        {
          "type": "client-registration",
          "url": "https://example2.com/en/legal/terms-of-use.html"
        }
      ]`

	invalidPackageLinkDueToMissingURL = `[
        {
          "type": "payment"
        },
        {
          "type": "client-registration",
          "url": "https://example2.com/en/legal/terms-of-use.html"
        }
      ]`

	invalidPackageLinkDueToWrongURL = `[
        {
          "type": "payment",
          "url": "wrongUrl"
        },
        {
          "type": "client-registration",
          "url": "https://example2.com/en/legal/terms-of-use.html"
        }
      ]`

	invalidPackageLinkTypeWhenProvidedCustomType = `[
        {
          "type": "payment",
          "url": "https://example2.com/en/legal/terms-of-use.html",
          "customType": "myCustomType"
        }
      ]`

	invalidPackageLinkCustomTypeWhenCustomTypeNotProvided = `[
        {
          "type": "custom",
          "url": "https://example2.com/en/legal/terms-of-use.html",
        }
      ]`

	invalidLinkDueToMissingTitle = `[
        {
          "url": "https://example2.com/en/legal/terms-of-use.html",
          "description": "foo bar"
        }
      ]`
	invalidLinkDueToMissingURL = `[
        {
          "title": "myTitle"
        }
      ]`
	invalidLinkDueToWrongURL = `[
        {
          "url": "wrongURL",
          "title": "myTitle"
        }
      ]`
	invalidLinkDueToInvalidLengthOfDescription = `[
        {
          "title": "myTitle",
          "url": "https://example2.com/en/legal/terms-of-use.html",
          "description": "%s"
        }
      ]`

	invalidPartOfProductsElement = `["invalidValue"]`

	invalidPartOfProductsIntegerElement = `["sap:S4HANA_OD", 992]`

	invalidTagsValue = `["invalid!@#"]`

	invalidTagsValueIntegerElement = `["storage", 992]`

	invalidLabelsWhenValueIsNotArray = `{
  		"label-key-1": "label-value-1"
		}`

	invalidLabelsWhenValuesAreNotArrayOfStrings = `{
  		"label-key-1": [
    	  "label-value-1",
    	  992
  		]
	}`

	invalidLabelsWhenKeyIsWrong = `{
  		"invalidKey!@#": [
    	  "label-value-1",
    	  "label-value-2"
  		]
	}`

	invalidPartnersWhenValueIsNotArray = `{
  		"partner-key-1": "partner-value-1"
	}`

	invalidPartnersWhenValuesAreNotArrayOfStrings = `[
    	  "microsoft:vendor:Microsoft",
    	  112
	]`

	invalidPartnersWhenValuesDoNotSatisfyRegex = `[
		"partner:partner:partner",
	]`

	invalidCountriesElement          = `["DE", "wrongCountry"]`
	invalidCountriesNonStringElement = `["DE", 992]`

	invalidLineOfBusinessElement          = `["sales", "wrongLineOfBusiness!@#"]`
	invalidLineOfBusinessNonStringElement = `["sales", 992]`

	invalidIndustryElement          = `["banking", "wrongIndustry!@#"]`
	invalidIndustryNonStringElement = `["banking", 992]`

	invalidBundleLinksDueToMissingTitle = `[
        {
		  "description": "foo bar",
          "url": "https://example.com/2018/04/11/testing/"
        }
      ]`

	invalidBundleLinksDueToMissingURL = `[
        {
		  "description": "foo bar",
		  "title": "myTitle"
        }
      ]`
	invalidBundleLinksDueToWrongURL = `[
        {
		  "description": "foo bar",
		  "title": "myTitle",
          "url": "wrongURL"
        }
      ]`

	invalidCredentialsExchangeStrategyDueToMissingType = `[
        {
		  "callbackUrl": "http://localhost:8080/credentials/relative"
        }
      ]`
	invalidCredentialsExchangeStrategyDueToWrongType = `[
        {
          "type": "wrongType",
		  "callbackUrl": "http://localhost:8080/credentials/relative"
        }
      ]`
	invalidCredentialsExchangeStrategyDueToMissingCustomType = `[
        {
          "type": "wrongType",
		  "customType": "ns:credential-exchange:v1",
		  "customDescription": "foo bar"
        }
      ]`
	invalidCredentialsExchangeStrategyDueToInvalidLenOfCustomDescription = `[
        {
		  "type": "custom",
		  "customType": "ns:credential-exchange:v1",
		  "customDescription": "%s"
        }
      ]`
	invalidCredentialsExchangeStrategyDueToWrongCustomType = `[
        {
          "type": "custom",
		  "customType": "wrongCustomType"
        }
      ]`
	invalidCredentialsExchangeStrategyDueToWrongCallbackURL = `[
        {
          "type": "custom",
		  "callbackUrl": "wrongURL"		  
        }
      ]`

	invalidAPIResourceLinksDueToMissingType = `[
        {
          "url": "https://example.com/shell/discover"
        },
		{
          "type": "console",
          "url": "%s/shell/discover/relative"
        }
      ]`
	invalidAPIResourceLinksDueToWrongType = `[
        {
          "type": "wrongType",
          "url": "https://example.com/shell/discover"
        }
      ]`
	invalidAPIResourceLinksDueToMissingCustomValueOfType = `[
        {
          "type": "console",
          "customType": "foo",
          "url": "https://example.com/shell/discover"
        }
      ]`
	invalidAPIResourceLinksDueToMissingURL = `[
        {
          "type": "console"
        }
      ]`
	invalidAPIResourceLinksDueToWrongURL = `[
        {
          "type": "console",
          "url": "wrongURL"
        }
      ]`

	invalidChangeLogEntriesDueToMissingVersion = `[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1"
        }
      ]`
	invalidChangeLogEntriesDueToWrongVersion = `[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "wrongValue"
        }
      ]`
	invalidChangeLogEntriesDueToMissingReleaseStatus = `[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`
	invalidChangeLogEntriesDueToWrongReleaseStatus = `[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "wrongValue",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`
	invalidChangeLogEntriesDueToMissingDate = `[
        {
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`
	invalidChangeLogEntriesDueToWrongDate = `[
        {
		  "date": "0000-00-00",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`
	invalidChangeLogEntriesDueToWrongURL = `[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "wrongValue",
          "version": "1.0.0"
        }
      ]`
	invalidChangeLogEntriesDueToInvalidLengthOfDescription = `[
        {
		  "date": "2020-04-29",
		  "description": "%s",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
		  "version": "1.0.0"
        }
      ]`

	invalidCorrelationIDsElement          = `["foo.bar.baz:123456", "wrongID"]`
	invalidCorrelationIDsNonStringElement = `["foo.bar.baz:123456", 992]`

	invalidEntryPointURI               = `["invalidUrl"]`
	invalidEntryPointsDueToDuplicates  = `["/test/v1", "/test/v1"]`
	invalidEntryPointsNonStringElement = `["/test/v1", 992]`

	invalidExtensibleDueToInvalidJSON                                 = `{invalid}`
	invalidExtensibleDueToInvalidSupportedType                        = `{"supported":true}`
	invalidExtensibleDueToNoSupportedProperty                         = `{"description":"Please find the extensibility documentation"}`
	invalidExtensibleDueToInvalidSupportedValue                       = `{"supported":"invalid"}`
	invalidExtensibleDueToSupportedAutomaticAndNoDescriptionProperty  = `{"supported":"automatic"}`
	invalidExtensibleDueToSupportedManualAndNoDescriptionProperty     = `{"supported":"manual"}`
	invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength = `{"supported":"%s", "description": "%s"}`

	invalidSuccessorsDueToInvalidAPIRegex   = `["sap.s4:apiResource:API_BILL_OF_MATERIAL_SRV:v2", "invalid-api-successor"]`
	invalidSuccessorsDueToInvalidEventRegex = `["sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1", "invalid-event-successor"]`

	invalidDescriptionFieldWithExceedingMaxLength = strings.Repeat("a", maxDescriptionLength+1)
)

func TestConfig_ValidateConfig(t *testing.T) {
	var tests = []struct {
		Name              string
		ConfigProvider    func() ord.WellKnownConfig
		BaseURL           string
		ExpectedToBeValid bool
	}{
		{
			Name: "Invalid 'baseURL' field for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.BaseURL = baseURL + "/full/path"
				return *config
			},
		},
		{
			Name: "Missing 'OpenResourceDiscoveryV1' field for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1 = ord.OpenResourceDiscoveryV1{}
				return *config
			},
		},
		{
			Name: "Missing 'url' field for document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].URL = ""
				return *config
			},
		},
		{
			Name: "Missing 'accessStrategies' field for document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies = nil
				return *config
			},
		},
		{
			Name: "Missing 'type' field for 'accessStrategies' field of document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = ""
				return *config
			},
		},
		{
			Name: "Invalid field `type` for `accessStrategies` field of document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = invalidType
				return *config
			},
		},
		{
			Name: "Invalid field `customType` when field `type` is not `custom` for `accessStrategies` field of document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = accessstrategy.OpenAccessStrategy
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomType = "foo"

				return *config
			},
		},
		{
			Name: "Invalid field `customType` when field `type` is `custom` for `accessStrategies` field of document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = accessstrategy.CustomAccessStrategy
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomType = invalidCustomType

				return *config
			},
		},
		{
			Name: "Field `type` is not `custom` when `customType` is valid for `accessStrategies` field of document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = accessstrategy.OpenAccessStrategy
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomType = "sap:custom-definition-format:v1"

				return *config
			},
		},
		{
			Name: "Invalid field `customDescription` when field `type` is not `custom` for `accessStrategies` field of document for config",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = accessstrategy.OpenAccessStrategy
				config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomDescription = "foo"

				return *config
			},
		},
		{
			Name: "Invalid when webhookURL is not /well-known, no config baseURL is set => empty baseURL and documents have relative URLs",
			ConfigProvider: func() ord.WellKnownConfig {
				config := fixWellKnownConfig()
				config.BaseURL = ""

				return *config
			},
			BaseURL: "",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			cfg := test.ConfigProvider()
			err := cfg.Validate(test.BaseURL)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateSystemInstance(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		CalculatedBaseURL *string
		ExpectedToBeValid bool
	}{
		{
			Name: "Invalid value for `correlationIds` field for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.CorrelationIDs = json.RawMessage(invalidCorrelationIDsElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it is invalid JSON for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.CorrelationIDs = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it isn't a JSON array for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.CorrelationIDs = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `correlationIds` field when the JSON array is empty for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.CorrelationIDs = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `correlationIds` field when it contains non string value for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.CorrelationIDs = json.RawMessage(invalidCorrelationIDsNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `baseUrl` for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.BaseURL = str.Ptr("http://test.com/test/v1")

				return []*ord.Document{doc}
			},
		}, {
			Name: "`baseUrl` of `DescribedSystemInstance` does not match the calculated baseURL",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.BaseURL = str.Ptr(baseURL2)

				return []*ord.Document{doc}
			},
		}, {
			Name: "No `baseUrl` of `DescribedSystemInstance` is provided when the calculated baseURL is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance = nil

				return []*ord.Document{doc}
			},
			CalculatedBaseURL: str.Ptr(""),
		}, {
			Name: "`baseUrl` of `DescribedSystemInstance` is different for each document when the calculated baseURL is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc2 := fixORDDocument()
				doc2.DescribedSystemInstance.BaseURL = str.Ptr(baseURL2)

				return []*ord.Document{doc, doc2}
			},
			CalculatedBaseURL: str.Ptr(""),
		}, {
			Name: "Invalid JSON `Labels` field for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.OrdLabels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.OrdLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.OrdLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.OrdLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.OrdLabels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `DocumentationLabels` field for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.DocumentationLabels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `DocumentationLabels` field for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array of strings for SystemInstance",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var docs ord.Documents
			if len(test.DocumentProvider()) == 0 {
				docs = ord.Documents{test.DocumentProvider()[0]}
			} else {
				docs = test.DocumentProvider()
			}

			var url string
			if test.CalculatedBaseURL != nil {
				url = *test.CalculatedBaseURL
			} else {
				url = baseURL
			}

			err := docs.Validate(url, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateDocument(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `OpenResourceDiscovery` field for Document",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `OpenResourceDiscovery` field for Document",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = "wrongValue"

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Only major version is checked for `OpenResourceDiscovery` field for Document",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = "1.4"

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidatePackage(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
		AfterTest         func()
	}{
		{
			Name: "Valid document",
			DocumentProvider: func() []*ord.Document {
				return []*ord.Document{fixORDDocument()}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `openResourceDiscovery` field for a Document",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `openResourceDiscovery` field for a Document",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = invalidOpenResourceDiscovery

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `baseUrl` of describedSystemInstance Document field",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.BaseURL = str.Ptr(invalidURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `ordID` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].OrdID = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].OrdID = invalidOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Title = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = strings.Repeat("a", invalidShortDescriptionLength)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = `newLine\n`

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `description` filed for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Description = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Exceeded length of `description` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Description = invalidDescriptionFieldWithExceedingMaxLength

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `version` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Version = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `version` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Version = invalidVersion

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Not incremented `version` field when Package has been changed",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(`["Mining"]`)

				newHash, err := ord.HashObject(doc.Packages[0])
				require.NoError(t, err)

				resourceHashes[packageORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		},
		{
			Name: "Valid incremented `version` field when package has changed",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(`["Utilities"]`)
				doc.Packages[0].Version = "2.1.4"

				hash, err := ord.HashObject(doc.Packages[0])
				require.NoError(t, err)

				resourceHashes[packageORDID] = hash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Missing `policyLevel` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `policyLevel` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = invalidPolicyLevel

				return []*ord.Document{doc}
			},
		}, {
			Name: "`policyLevel` field for Package is not of type `custom` when `customPolicyLevel` is set",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].CustomPolicyLevel = str.Ptr("myCustomPolicyLevel")
				doc.Packages[0].PolicyLevel = policyLevel

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `CustomPolicyLevel` field value for Package when `PolicyLevel` is set to `custom`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].CustomPolicyLevel = str.Ptr("invalid-value")
				doc.Packages[0].PolicyLevel = "custom"

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `type` from `PackageLinks` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToMissingType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `type` key in `PackageLinks` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToWrongType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `url` from `PackageLinks` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToMissingURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `url` key in `PackageLinks` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Field `type` in `PackageLinks` is not set to `custom` when `customType` field is provided",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkTypeWhenProvidedCustomType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `type` set to `custom` in `PackageLinks` when `customType` field is not provided",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkCustomTypeWhenCustomTypeNotProvided)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `PackageLinks` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `PackageLinks` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Valid `PackageLinks` field when it is an empty JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `title` field in `Links` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidLinkDueToMissingTitle)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidLinkDueToMissingURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `links` field when it is an empty JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `url` field in `Links` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidLinkDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding length in `Links` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty `description` field in `Links` for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `vendor` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Vendor = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Vendor = str.Ptr(invalidVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field for Package when `policyLevel` is sap-partner",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.SapVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field for Package when `policyLevel` is sap",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `partOfProducts` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `partOfProducts` field when the JSON array is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid element of `partOfProducts` array field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(invalidPartOfProductsElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it contains non string value",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(invalidPartOfProductsIntegerElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage(invalidTagsValue)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `tags` field when the JSON array is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `tags` field when it contains non string value",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage(invalidTagsValueIntegerElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `DocumentationLabels` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array of strings for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage(invalidCountriesElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when JSON array contains non string element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage(invalidCountriesNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `countries` field when the JSON array is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Invalid `lineOfBusiness` field element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when JSON array contains non string element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `lineOfBusiness` field when the JSON array is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `lineOfBusiness` field when `policyLevel` is `sap`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when `policyLevel` is `sap partner`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name:              "Valid `lineOfBusiness` field when `policyLevel` is `custom`",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelCustom

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(invalidIndustryElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when JSON array contains non string element for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(invalidIndustryNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it is invalid JSON for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `industry` field when the JSON array is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `industry` field when `policyLevel` is `sap`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when `policyLevel` is `sap partner`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name:              "Valid `industry` field when `policyLevel` is `custom`",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelCustom

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid empty `supportInfo` field for Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				emptyStr := ""
				doc.Packages[0].SupportInfo = &emptyStr

				return []*ord.Document{doc}
			},
		},

		// Test invalid entity relations

		{
			Name: "Package has a reference to unknown Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].Vendor = str.Ptr(unknownVendorOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Package has a reference to unknown Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(fmt.Sprintf(`["%s"]`, unknownProductOrdID))

				return []*ord.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)

			if test.AfterTest != nil {
				test.AfterTest()
			}

			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateBundle(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `ordID` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].OrdID = nil

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `ordID` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].OrdID = str.Ptr(invalidOrdID)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `title` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = ""

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Valid missing `shortDescription` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = nil

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Exceeded length of `shortDescription` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = str.Ptr(strings.Repeat("a", invalidShortDescriptionLength))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid empty `shortDescription` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = str.Ptr("")

				return []*ord.Document{doc}
			},
		},
		{
			Name: "New lines in `shortDescription` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = str.Ptr(`newLine\n`)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Valid missing `description` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = nil

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Exceeded length of `description` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = str.Ptr(invalidDescriptionFieldWithExceedingMaxLength)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid empty `description` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = str.Ptr("")

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `title` field in `Links` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidBundleLinksDueToMissingTitle)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `url` field in `Links` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidBundleLinksDueToMissingURL)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `url` field in `Links` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidBundleLinksDueToWrongURL)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `description` field with exceeding length in `Links` for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid empty `description` field in `Links` for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, ""))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `Links` field when it is invalid JSON for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `Links` field when it isn't a JSON array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Valid `Links` field when it is an empty JSON array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Invalid JSON `Labels` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid JSON object `Labels` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`Labels` values are not array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`Labels` values are not array of strings for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid key for JSON `Labels` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid JSON `DocumentationLabels` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].DocumentationLabels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid JSON object `DocumentationLabels` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`DocumentationLabels` values are not array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`DocumentationLabels` values are not array of strings for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `type` field of `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToMissingType)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `type` field of `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongType)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`type` field is not with value `custom` when `customType` field is provided for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongCustomType)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `customType` field when `type` field is set to `custom` for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongCustomType)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`type` field is not with value `custom` when `customDescription` field is provided for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToMissingCustomType)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`type` field is with value `custom` but `customDescription` field is empty for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(invalidCredentialsExchangeStrategyDueToInvalidLenOfCustomDescription, ""))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "`type` field is with value `custom` but `customDescription` field is with exceeding length for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(fmt.Sprintf(invalidCredentialsExchangeStrategyDueToInvalidLenOfCustomDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `callbackURL` field of `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongCallbackURL)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `CredentialExchangeStrategies` field when it is invalid JSON for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `CredentialExchangeStrategies` field when it isn't a JSON array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Valid `CredentialExchangeStrategies` field when it is an empty JSON array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Invalid `correlationIds` field when it is invalid JSON for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CorrelationIDs = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `correlationIds` field when it isn't a JSON array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CorrelationIDs = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Valid `correlationIds` field when it is an empty JSON array for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CorrelationIDs = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Invalid `correlationIds` field when it contains non string value for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CorrelationIDs = json.RawMessage(invalidCorrelationIDsNonStringElement)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid value for `correlationIds` field for Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CorrelationIDs = json.RawMessage(invalidCorrelationIDsElement)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Success when `correlationIds` are valid",
			DocumentProvider: func() []*ord.Document {
				return []*ord.Document{fixORDDocument()}
			},
			ExpectedToBeValid: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateAPI(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
		AfterTest         func()
	}{
		{
			Name: "Missing `ordID` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdID = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdID = str.Ptr(invalidOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `title` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Name = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = str.Ptr(strings.Repeat("a", invalidShortDescriptionLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = str.Ptr("")

				return []*ord.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = str.Ptr(`newLine\n`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `description` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Description = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding max length for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Description = str.Ptr(invalidDescriptionFieldWithExceedingMaxLength)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `version` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].VersionInput.Value = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `version` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].VersionInput.Value = invalidVersion

				return []*ord.Document{doc}
			},
		}, {
			Name: "Not incremented `version` field when resource definition's URL has changed for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].URL = "http://newurl.com/odata/$metadata"

				newHash, err := ord.HashObject(doc.APIResources[0])
				require.NoError(t, err)

				resourceHashes[api1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource definition's MediaType has changed for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatTextYAML

				newHash, err := ord.HashObject(doc.APIResources[0])
				require.NoError(t, err)

				resourceHashes[api1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource definition's Type has changed for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2

				newHash, err := ord.HashObject(doc.APIResources[0])
				require.NoError(t, err)

				resourceHashes[api1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource definition's CustomType has changed for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeCustom
				doc.APIResources[0].ResourceDefinitions[0].CustomType = "sap:custom-definition-format:v1"

				newHash, err := ord.HashObject(doc.APIResources[0])
				require.NoError(t, err)

				resourceHashes[api1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource has changed for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(`["Utilities"]`)

				newHash, err := ord.HashObject(doc.APIResources[0])
				require.NoError(t, err)

				resourceHashes[api1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Valid incremented `version` field when resource definition has changed for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeCustom
				doc.APIResources[0].ResourceDefinitions[0].CustomType = "sap:custom-definition-format:v1"
				doc.APIResources[0].VersionInput.Value = "2.1.4"

				newHash, err := ord.HashObject(doc.APIResources[0])
				require.NoError(t, err)

				resourceHashes[api1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `partOfPackage` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdPackageID = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfPackage` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdPackageID = str.Ptr(invalidOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `apiProtocol` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIProtocol = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `apiProtocol` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIProtocol = str.Ptr("wrongAPIProtocol")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `visibility` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Visibility = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `visibility` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Visibility = str.Ptr("wrongVisibility")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid element of `partOfProducts` array field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `partOfProducts` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `partOfProducts` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsIntegerElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `tags` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage(invalidTagsValue)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `tags` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `tags` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage(invalidTagsValueIntegerElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `countries` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage(invalidCountriesElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `countries` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `countries` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage(invalidCountriesNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `lineOfBusiness` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `lineOfBusiness` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `lineOfBusiness` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(invalidCountriesNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when `policyLevel` is `sap` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when `policyLevel` is `sap partner` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name:              "Valid `lineOfBusiness` field when `policyLevel` is `custom` for API",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelCustom

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `industry` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(invalidIndustryElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `industry` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `industry` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(invalidIndustryNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when `policyLevel` is `sap` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when `policyLevel` is `sap partner` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name:              "Valid `industry` field when `policyLevel` is `custom`",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelCustom

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid missing `resourceDefinitions` field for API when `policyLevel` is sap and `visibility` is private",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions = nil
				doc.APIResources[0].Visibility = str.Ptr(ord.APIVisibilityPrivate)
				doc.Packages[0].PolicyLevel = policyLevel

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing field `type` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `type` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = invalidType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Field `type` value is not `custom` when field `customType` is provided for `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].CustomType = "test:test:v1"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `customType` value when field `type` has value `custom`for `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "custom"
				doc.APIResources[0].ResourceDefinitions[0].CustomType = invalidCustomType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `mediaType` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].MediaType = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].MediaType = invalidMediaType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `openapi-v2` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "openapi-v2"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `openapi-v3` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "openapi-v3"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `raml-v1` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "raml-v1"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `edmx` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "edmx"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `csdl-json` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "csdl-json"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `wsdl-v1` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "wsdl-v1"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `wsdl-v2` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "wsdl-v2"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `sap-rfc-metadata-v1` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "sap-rfc-metadata-v1"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `url` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].URL = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].URL = invalidURL

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `type` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `type` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = invalidType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = "foo"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is `custom` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "custom"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = invalidCustomType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Field `type` is not `custom` when `customType` is valid for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = "sap:custom-definition-format:v1"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `customType` value when field `type` has value `custom` for `accessStrategies` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "custom"
				doc.APIResources[0].ResourceDefinitions[0].CustomType = invalidCustomType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customDescription` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomDescription = "foo"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customDescription` with exceeding length when field `type` is `custom` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "custom"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomDescription = invalidDescriptionFieldWithExceedingMaxLength

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `type` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidAPIResourceLinksDueToMissingType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `type` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidAPIResourceLinksDueToWrongType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is not `custom` for `apiResourceLink` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidAPIResourceLinksDueToMissingCustomValueOfType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `url` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidAPIResourceLinksDueToMissingURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `url` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidAPIResourceLinksDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `apiResourceLink` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `apiResourceLink` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `apiResourceLink` field when it is an empty JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `title` field in `Links` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidLinkDueToMissingTitle)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidLinkDueToMissingURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `url` field in `Links` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidLinkDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding length in `Links` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty `description` field in `Links` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `links` field when it is an empty JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `releaseStatus` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `releaseStatus` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("wrongValue")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `sunsetDate` field when `releaseStatus` field has value `deprecated` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `sunsetDate` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("deprecated")
				doc.APIResources[0].SunsetDate = str.Ptr("0000-00-00T09:35:30+0000")
				doc.APIResources[0].Successors = json.RawMessage(fmt.Sprintf(`["%s"]`, api2ORDID))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `successors` field when `releaseStatus` field has value `deprecated` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `successors` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Successors = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `successors` when values do not match the regex for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Successors = json.RawMessage(invalidSuccessorsDueToInvalidAPIRegex)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `version` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingVersion)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `version` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongVersion)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `releaseStatus` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingReleaseStatus)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `releaseStatus` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongReleaseStatus)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `date` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingDate)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `date` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongDate)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty field `description` of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(fmt.Sprintf(invalidChangeLogEntriesDueToInvalidLengthOfDescription, ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `description` with exceeding length of field `changeLogEntries` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(fmt.Sprintf(invalidChangeLogEntriesDueToInvalidLengthOfDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `changeLogEntries` field when it is an empty JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid when `entryPoints` field is empty but `PartOfConsumptionBundles` field is not for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid when `entryPoints` field is empty and `PartOfConsumptionBundles` field is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = nil
				doc.APIResources[0].PartOfConsumptionBundles = nil

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid when `defaultConsumptionBundle` field doesn't match the required regex for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].DefaultConsumptionBundle = str.Ptr(invalidBundleOrdID)
				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `defaultConsumptionBundle` field is not part of any bundles in the `partOfConsumptionBundles` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].DefaultConsumptionBundle = str.Ptr(secondBundleORDID)
				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `entryPoints` when containing invalid URI for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidEntryPointURI)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `entryPoints` when containing duplicate URIs for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidEntryPointsDueToDuplicates)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidEntryPointsNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `DocumentationLabels` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].DocumentationLabels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `DocumentationLabels` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array of strings for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `implementationStandard` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr(invalidType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandard` field is valid but `implementationStandard` field is missing for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = nil
				doc.APIResources[0].CustomImplementationStandard = str.Ptr("sap.s4:ATTACHMENT_SRV:v1")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandard` field is valid but `implementationStandard` field is not set to `custom` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].CustomImplementationStandard = str.Ptr("sap.s4:ATTACHMENT_SRV:v1")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `customImplementationStandard` field when `implementationStandard` field is set to `custom` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr("custom")
				doc.APIResources[0].CustomImplementationStandard = str.Ptr(invalidType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `customImplementationStandard` field when `implementationStandard` field is set to `custom` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr("custom")
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandardDescription` is set but `implementationStandard` field is missing for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = nil
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandardDescription` is set but `implementationStandard` field is not `custom` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandardDescription` is set but `implementationStandard` field is not `custom` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `customImplementationStandardDescription` field when `implementationStandard` field is set to `custom` for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr("custom")
				doc.APIResources[0].CustomImplementationStandard = str.Ptr("sap.s4:ATTACHMENT_SRV:v1")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `ordId` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].BundleOrdID = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `ordId` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].BundleOrdID = invalidBundleOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Duplicate `ordId` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles = append(doc.APIResources[0].PartOfConsumptionBundles, &model.ConsumptionBundleReference{BundleOrdID: bundleORDID})

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `defaultEntryPoint` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL = invalidURL

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `defaultEntryPoint` field from `entryPoints` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL = "https://exmaple.com/test/v3"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Present `defaultEntryPoint` field even though there is a single element in `entryPoints` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[1].PartOfConsumptionBundles[0].DefaultTargetURL = "https://exmaple.com/test/v3"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Empty `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles = []*model.ConsumptionBundleReference{}

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `Extensible` field when `policyLevel` is sap for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = nil
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `Extensible` field when `policyLevel` is sap partner for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = nil
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `Extensible` field due to empty json object for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(`{}`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `Extensible` field due to invalid json for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `supported` field in the `extensible` object for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToNoSupportedProperty)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `supported` field type in the `extensible` object for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `supported` field value in the `extensible` object for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedValue)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `description` field when `supported` has an `automatic` value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedAutomaticAndNoDescriptionProperty)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `description` field when `supported` has a `manual` value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedManualAndNoDescriptionProperty)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Empty `description` field when `supported` has a `manual` value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "manual", ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Empty `description` field when `supported` has a `automatic` value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "automatic", ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding length when `supported` has a `manual` value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "manual", invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding length when `supported` has a `automatic` value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "automatic", invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `WSDL V1` and `WSDL V2` definitions when APIResources has policyLevel `sap` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapInbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV2
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Valid `WSDL V1` and `WSDL V2` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapInbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapInbound
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV2
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Valid `WSDL V1` and `WSDL V2` definitions when APIResources has policyLevel `sap` and apiProtocol is `soap-outbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapOutbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapOutbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV2
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Valid `SAP RFC Metadata` definitions when APIResources has policyLevel `sap` and apiProtocol is `sap-rfc`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSapRfc
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeRfcMetadata
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Valid `SAP RFC Metadata` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `sap-rfc`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSapRfc
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeRfcMetadata
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapInbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap-partner` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapInbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap` and apiProtocol is `soap-outbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapOutbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapOutbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap-partner` and apiProtocol is `soap-outbound`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSoapOutbound
				*doc.APIResources[1].APIProtocol = ord.APIProtocolSoapOutbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap` and apiProtocol is `odata-v2`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolODataV2
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `odata-v2`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolODataV2
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap` and apiProtocol is `odata-v4`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolODataV4
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `odata-v4`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolODataV4
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `OpenAPI` definitions when APIResources has policyLevel `sap` and apiProtocol is `rest`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolRest
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `OpenAPI` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `rest`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolRest
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `SAP RFC` definitions when APIResources has policyLevel `sap` and apiProtocol is `sap-rfc-metadata-v1`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSapRfc
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `SAP RFC` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `sap-rfc-metadata-v1`",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)
				*doc.APIResources[0].APIProtocol = ord.APIProtocolSapRfc
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*ord.Document{doc}
			},
		},
		// Test invalid entity relations

		{
			Name: "API has a reference to an unknown Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdPackageID = str.Ptr(unknownPackageOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "API has a reference to an unknown Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles = fixBundleCreateInput()
				doc.APIResources[0].PartOfConsumptionBundles = fixAPIPartOfConsumptionBundles()
				doc.APIResources[0].PartOfConsumptionBundles[0].BundleOrdID = unknownBundleOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "API has a reference to an unknown Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(fmt.Sprintf(`["%s"]`, unknownProductOrdID))

				return []*ord.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)

			if test.AfterTest != nil {
				test.AfterTest()
			}

			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateEvent(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
		AfterTest         func()
	}{
		{
			Name: "Missing `ordID` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdID = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdID = str.Ptr(invalidOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Name = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = str.Ptr(strings.Repeat("a", invalidShortDescriptionLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = str.Ptr("")

				return []*ord.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = str.Ptr(`newLine\n`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `description` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Description = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding length for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Description = str.Ptr(invalidDescriptionFieldWithExceedingMaxLength)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `version` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].VersionInput.Value = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Not incremented `version` field when resource definition's URL has changed for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].URL = "http://newurl.com/odata/$metadata"

				newHash, err := ord.HashObject(doc.EventResources[0])
				require.NoError(t, err)

				resourceHashes[event1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource definition's MediaType has changed for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatTextYAML

				newHash, err := ord.HashObject(doc.EventResources[0])
				require.NoError(t, err)

				resourceHashes[event1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource definition's Type has changed for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = model.EventSpecTypeCustom
				doc.EventResources[0].ResourceDefinitions[0].CustomType = "sap:custom-definition-format:v1"

				newHash, err := ord.HashObject(doc.EventResources[0])
				require.NoError(t, err)

				resourceHashes[event1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Not incremented `version` field when resource has changed for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(`["Utilities"]`)

				newHash, err := ord.HashObject(doc.EventResources[0])
				require.NoError(t, err)

				resourceHashes[event1ORDID] = newHash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
		}, {
			Name: "Valid incremented `version` field when resource definition has changed for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = model.EventSpecTypeCustom
				doc.EventResources[0].ResourceDefinitions[0].CustomType = "sap:custom-definition-format:v1"
				doc.EventResources[0].VersionInput.Value = "2.1.4"

				hash, err := ord.HashObject(doc.EventResources[0])
				require.NoError(t, err)

				resourceHashes[event1ORDID] = hash

				return []*ord.Document{doc}
			},
			AfterTest: func() {
				resourceHashes = fixResourceHashes()
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `version` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].VersionInput.Value = invalidVersion

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `version` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingVersion)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `version` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongVersion)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `releaseStatus` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingReleaseStatus)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `releaseStatus` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongReleaseStatus)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `date` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingDate)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `date` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongDate)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty field `description` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(fmt.Sprintf(invalidChangeLogEntriesDueToInvalidLengthOfDescription, ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `description` with exceeding length of field `changeLogEntries` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(fmt.Sprintf(invalidChangeLogEntriesDueToInvalidLengthOfDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `changeLogEntries` field when it is an empty JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `partOfPackage` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdPackageID = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfPackage` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdPackageID = str.Ptr(invalidOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `visibility` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Visibility = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `visibility` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Visibility = str.Ptr("wrongVisibility")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `title` field in `Links` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidLinkDueToMissingTitle)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidLinkDueToMissingURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `url` field in `Links` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidLinkDueToWrongURL)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `description` field with exceeding length in `Links` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid empty `description` field in `Links` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(fmt.Sprintf(invalidLinkDueToInvalidLengthOfDescription, ""))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `links` field when it is an empty JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid element of `partOfProducts` array field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `partOfProducts` field when the JSON array is empty for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `partOfProducts` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it contains non string value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsIntegerElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid missing `resourceDefinitions` field for Event when `policyLevel` is sap and `visibility` is private",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions = nil
				doc.EventResources[0].Visibility = str.Ptr(ord.APIVisibilityPrivate)
				doc.Packages[0].PolicyLevel = policyLevel

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `type` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `type` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = invalidType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Field `type` value is not `custom` when field `customType` is provided for `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].CustomType = "test:test:v1"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `customType` value when field `type` has value `custom`for `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = "custom"
				doc.EventResources[0].ResourceDefinitions[0].CustomType = invalidCustomType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `mediaType` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].MediaType = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].MediaType = invalidMediaType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `url` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].URL = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].URL = invalidURL

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy = nil

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing field `type` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `type` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = invalidType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = "foo"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is `custom` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "custom"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = invalidCustomType

				return []*ord.Document{doc}
			},
		}, {
			Name: "Field `type` is not `custom` when `customType` is valid for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = "sap:custom-definition-format:v1"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customDescription` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomDescription = "foo"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid field `customDescription` with exceeding length when field `type` is `custom` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "custom"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomDescription = invalidDescriptionFieldWithExceedingMaxLength

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `tags` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage(invalidTagsValue)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `tags` field when the JSON array is empty for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `tags` field when it contains non string value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage(invalidTagsValueIntegerElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `DocumentationLabels` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array of strings for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `countries` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage(invalidCountriesElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `countries` field when the JSON array is empty for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `countries` field when it contains non string value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage(invalidCountriesNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `lineOfBusiness` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `lineOfBusiness` field when the JSON array is empty for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `lineOfBusiness` field when it contains non string value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(invalidCountriesNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when `policyLevel` is `sap` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when `policyLevel` is `sap partner` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name:              "Valid `lineOfBusiness` field when `policyLevel` is `custom`",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(`["LoB"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelCustom

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `industry` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(invalidIndustryElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it is invalid JSON for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `industry` field when the JSON array is empty for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `industry` field when it contains non string value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(invalidIndustryNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when `policyLevel` is `sap` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when `policyLevel` is `sap partner` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		}, {
			Name:              "Valid `industry` field when `policyLevel` is `custom`",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(`["SomeIndustry"]`)
				doc.Packages[0].PolicyLevel = ord.PolicyLevelCustom

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `releaseStatus` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `releaseStatus` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("wrongValue")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `sunsetDate` field when `releaseStatus` field has value `deprecated` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("deprecated")
				doc.EventResources[0].Successors = json.RawMessage(fmt.Sprintf(`["%s"]`, event2ORDID))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `sunsetDate` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("deprecated")
				doc.EventResources[0].SunsetDate = str.Ptr("0000-00-00T09:35:30+0000")
				doc.EventResources[0].Successors = json.RawMessage(fmt.Sprintf(`["%s"]`, event2ORDID))

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `successors` field when `releaseStatus` field has value `deprecated` for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid json field `successors` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Successors = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `successors` when values do not match the regex for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Successors = json.RawMessage(invalidSuccessorsDueToInvalidEventRegex)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `ordId` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles[0].BundleOrdID = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `ordId` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles[0].BundleOrdID = invalidBundleOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Duplicate `ordId` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles = append(doc.EventResources[0].PartOfConsumptionBundles, &model.ConsumptionBundleReference{BundleOrdID: bundleORDID})

				return []*ord.Document{doc}
			},
		}, {
			Name: "Present `defaultEntryPoint` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles[0].DefaultTargetURL = "https://exmaple.com/test/v3"

				return []*ord.Document{doc}
			},
		}, {
			Name: "Empty `PartOfConsumptionBundle` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles = []*model.ConsumptionBundleReference{}

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid when `defaultConsumptionBundle` field doesn't match the required regex for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].DefaultConsumptionBundle = str.Ptr(invalidBundleOrdID)
				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid when `defaultConsumptionBundle` field is not part of any bundles in the `partOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].DefaultConsumptionBundle = str.Ptr(secondBundleORDID)
				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `Extensible` field when `policyLevel` is sap",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = nil
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSap

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `Extensible` field when `policyLevel` is sap partner",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = nil
				doc.Packages[0].PolicyLevel = ord.PolicyLevelSapPartner
				doc.Packages[0].Vendor = str.Ptr(ord.PartnerVendor)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `Extensible` field due to empty json object",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(`{}`)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `Extensible` field due to invalid json",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidJSON)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `supported` field in the `extensible` object for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToNoSupportedProperty)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `supported` field type in the `extensible` object for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedType)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `supported` field value in the `extensible` object for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedValue)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `description` field when `supported` has an `automatic` value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedAutomaticAndNoDescriptionProperty)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Missing `description` field when `supported` has a `manual` value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedManualAndNoDescriptionProperty)

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Empty `description` field when `supported` has a `manual` value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "manual", ""))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Empty `description` field when `supported` has a `automatic` value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "automatic", ""))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `description` field with exceeding length when `supported` has a `manual` value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "manual", invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		},
		{
			Name: "Invalid `description` field with exceeding length when `supported` has a `automatic` value for Event",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(fmt.Sprintf(invalidExtensibleDueToCorrectSupportedButInvalidDescriptionLength, "automatic", invalidDescriptionFieldWithExceedingMaxLength))

				return []*ord.Document{doc}
			},
		},

		// Test invalid entity relations

		{
			Name: "Event has a reference to unknown Package",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdPackageID = str.Ptr(unknownPackageOrdID)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Event has a reference to unknown Bundle",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles = fixBundleCreateInput()
				doc.EventResources[0].PartOfConsumptionBundles = fixEventPartOfConsumptionBundles()
				doc.EventResources[0].PartOfConsumptionBundles[0].BundleOrdID = unknownBundleOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Event has a reference to unknown Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(fmt.Sprintf(`["%s"]`, unknownProductOrdID))

				return []*ord.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)

			if test.AfterTest != nil {
				test.AfterTest()
			}

			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateProduct(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
	}{
		{
			Name:              "Valid `id` field for Product",
			ExpectedToBeValid: true,
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products = append(doc.Products, &model.ProductInput{
					OrdID:            "sap:product:test:",
					Title:            "title",
					ShortDescription: "Description",
					Vendor:           ord.SapVendor,
					Parent:           nil,
					CorrelationIDs:   nil,
					Labels:           nil,
				})

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `id` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].OrdID = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `id` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].OrdID = invalidOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Title = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].ShortDescription = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].ShortDescription = strings.Repeat("a", invalidShortDescriptionLength)

				return []*ord.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].ShortDescription = `newLine\n`

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `vendor` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Vendor = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Vendor = invalidOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field when namespace in the `id` is `sap` for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].OrdID = "sap:product:S4HANA_OD:"
				doc.Products[0].Vendor = vendor2ORDID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field when namespace in the `id` is not `sap` for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].OrdID = "strange:product:S4HANA_OD:"
				doc.Products[0].Vendor = vendorORDID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `parent` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Parent = str.Ptr(invalidType)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid value for `correlationIds` field for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIDs = json.RawMessage(invalidCorrelationIDsElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it is invalid JSON for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIDs = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it isn't a JSON array for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIDs = json.RawMessage("{}")

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `correlationIds` field when the JSON array is empty for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIDs = json.RawMessage("[]")

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Invalid `correlationIds` field when it contains non string value for API",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIDs = json.RawMessage(invalidCorrelationIDsNonStringElement)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `DocumentationLabels` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].DocumentationLabels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `DocumentationLabels` field for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array of strings for Product",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		},

		// Test invalid entity relations

		{
			Name: "Product has a reference to unknown Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Products[0].Vendor = unknownVendorOrdID

				return []*ord.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateVendor(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `id` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `id` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = invalidOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Title = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON `DocumentationLabels` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].DocumentationLabels = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `DocumentationLabels` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].DocumentationLabels = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`DocumentationLabels` values are not array of strings for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].DocumentationLabels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Partners` field for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Partners = json.RawMessage(invalidJSON)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Partners` values are not array for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Partners = json.RawMessage(invalidPartnersWhenValueIsNotArray)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Partners` values are not array of strings for Vendor",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Partners = json.RawMessage(invalidPartnersWhenValuesAreNotArrayOfStrings)

				return []*ord.Document{doc}
			},
		}, {
			Name: "`Partners` values do not match the regex rule",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Partners = json.RawMessage(invalidPartnersWhenValuesDoNotSatisfyRegex)

				return []*ord.Document{doc}
			},
		}, {
			Name: "Valid `Partners` field when the JSON array is empty",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Partners = json.RawMessage(`[]`)

				return []*ord.Document{doc}
			},
			ExpectedToBeValid: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateTombstone(t *testing.T) {
	var tests = []struct {
		Name              string
		DocumentProvider  func() []*ord.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `ordId` field for Tombstone",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `ordId` field for Tombstone",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = invalidOrdID

				return []*ord.Document{doc}
			},
		}, {
			Name: "Missing `removalDate` field for Tombstone",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].RemovalDate = ""

				return []*ord.Document{doc}
			},
		}, {
			Name: "Invalid `removalDate` field for Tombstone",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].RemovalDate = "0000-00-00T15:04:05Z"

				return []*ord.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, nil)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDocuments_ValidateMultipleErrors(t *testing.T) {
	var tests = []struct {
		Name                   string
		DocumentProvider       func() []*ord.Document
		ExpectedStringsInError []string
	}{
		{
			Name: "Invalid value for `correlationIds` field for SystemInstance and invalid `baseUrl` for SystemInstance in one document",
			DocumentProvider: func() []*ord.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.CorrelationIDs = json.RawMessage(invalidCorrelationIDsElement)
				doc.DescribedSystemInstance.BaseURL = str.Ptr("http://test.com/test/v1")

				return []*ord.Document{doc}
			},
			ExpectedStringsInError: []string{"baseUrl", "correlationIds"},
		},
		{
			Name: "Invalid value for `correlationIds` field for SystemInstance in first doc and invalid `baseUrl` for SystemInstance in second doc",
			DocumentProvider: func() []*ord.Document {
				doc1 := fixORDDocument()
				doc1.DescribedSystemInstance.CorrelationIDs = json.RawMessage(invalidCorrelationIDsElement)
				doc2 := fixORDDocument()
				doc2.DescribedSystemInstance.BaseURL = str.Ptr("http://test.com/test/v1")

				return []*ord.Document{doc1, doc2}
			},
			ExpectedStringsInError: []string{"baseUrl", "correlationIds"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := ord.Documents(test.DocumentProvider())
			err := docs.Validate(baseURL, apisFromDB, eventsFromDB, pkgsFromDB, resourceHashes, map[string]bool{})
			if len(test.ExpectedStringsInError) != 0 {
				require.Error(t, err)
				for _, expectedStr := range test.ExpectedStringsInError {
					require.Contains(t, err.Error(), expectedStr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
