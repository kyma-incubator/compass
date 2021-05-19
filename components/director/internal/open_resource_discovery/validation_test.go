package open_resource_discovery_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

const (
	invalidOpenResourceDiscovery = "invalidOpenResourceDiscovery"
	invalidUrl                   = "invalidUrl"
	invalidOrdID                 = "invalidOrdId"
	invalidDescriptionLength     = 256
	invalidVersion               = "invalidVersion"
	invalidPolicyLevel           = "invalidPolicyLevel"
	invalidVendor                = "wrongVendor!"
	invalidType                  = "invalidType"
	invalidCustomType            = "wrongCustomType"
	invalidMediaType             = "invalid/type"
	invalidBundleOrdID           = "ns:wrongConsumptionBundle:v1"

	unknownVendorOrdID  = "nsUNKNOWN:vendor:id:"
	unknownProductOrdID = "nsUNKNOWN:product:id:"
	unknownPackageOrdID = "ns:package:UNKNOWN_PACKAGE_ID:v1"
	unknownBundleOrdID  = "ns:consumptionBundle:UNKNOWN_BUNDLE_ID:v1"
)

var (
	invalidJson = `[
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

	invalidPartnersWhenValuesAreNotArrayOfStrings = `{
  		"partners-key-1": [
    	  "partners-value-1",
    	  112
  		]
	}`

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

	invalidApiResourceLinksDueToMissingType = `[
        {
          "url": "https://example.com/shell/discover"
        },
		{
          "type": "console",
          "url": "%s/shell/discover/relative"
        }
      ]`
	invalidApiResourceLinksDueToWrongType = `[
        {
          "type": "wrongType",
          "url": "https://example.com/shell/discover"
        }
      ]`
	invalidApiResourceLinksDueToMissingCustomValueOfType = `[
        {
          "type": "console",
          "customType": "foo",
          "url": "https://example.com/shell/discover"
        }
      ]`
	invalidApiResourceLinksDueToMissingURL = `[
        {
          "type": "console"
        }
      ]`
	invalidApiResourceLinksDueToWrongURL = `[
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

	invalidCorrelationIdsElement          = `["foo.bar.baz:123456", "wrongID"]`
	invalidCorrelationIdsNonStringElement = `["foo.bar.baz:123456", 992]`

	invalidEntryPointURI               = `["invalidUrl"]`
	invalidEntryPointsDueToDuplicates  = `["/test/v1", "/test/v1"]`
	invalidEntryPointsNonStringElement = `["/test/v1", 992]`

	invalidExtensibleDueToInvalidSupportedType                       = `{"supported":true}`
	invalidExtensibleDueToNoSupportedProperty                        = `{"description":"Please find the extensibility documentation"}`
	invalidExtensibleDueToInvalidSupportedValue                      = `{"supported":"invalid"}`
	invalidExtensibleDueToSupportedAutomaticAndNoDescriptionProperty = `{"supported":"automatic"}`
	invalidExtensibleDueToSupportedManualAndNoDescriptionProperty    = `{"supported":"manual"}`
)

func TestDocuments_ValidateSystemInstance(t *testing.T) {
	var tests = []struct {
		Name                   string
		SystemInstanceProvider func() *model.Application
		ExpectedToBeValid      bool
	}{
		{
			Name: "Invalid value for `correlationIds` field for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.CorrelationIds = json.RawMessage(invalidCorrelationIdsElement)

				return sysInst
			},
		}, {
			Name: "Invalid `correlationIds` field when it is invalid JSON for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.CorrelationIds = json.RawMessage(invalidJson)

				return sysInst
			},
		}, {
			Name: "Invalid `correlationIds` field when it isn't a JSON array for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.CorrelationIds = json.RawMessage("{}")

				return sysInst
			},
		}, {
			Name: "Invalid `correlationIds` field when the JSON array is empty for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.CorrelationIds = json.RawMessage("[]")

				return sysInst
			},
		}, {
			Name: "Invalid `correlationIds` field when it contains non string value for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.CorrelationIds = json.RawMessage(invalidCorrelationIdsNonStringElement)

				return sysInst
			},
		}, {
			Name: "Invalid `baseUrl` for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.BaseURL = str.Ptr("http://test.com/test/v1")

				return sysInst
			},
		}, {
			Name: "Invalid JSON `Labels` field for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.Labels = json.RawMessage(invalidJson)

				return sysInst
			},
		}, {
			Name: "Invalid JSON object `Labels` field for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.Labels = json.RawMessage(`[]`)

				return sysInst
			},
		}, {
			Name: "`Labels` values are not array for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return sysInst
			},
		}, {
			Name: "`Labels` values are not array of strings for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return sysInst
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for SystemInstance",
			SystemInstanceProvider: func() *model.Application {
				sysInst := fixSystemInstance()
				sysInst.Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return sysInst
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			systemInstance := test.SystemInstanceProvider
			err := open_resource_discovery.ValidateSystemInstanceInput(systemInstance())
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `OpenResourceDiscovery` field for Document",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `OpenResourceDiscovery` field for Document",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = "wrongValue"

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Valid document",
			DocumentProvider: func() []*open_resource_discovery.Document {
				return []*open_resource_discovery.Document{fixORDDocument()}
			},
			ExpectedToBeValid: true,
		}, {
			Name: "Missing `openResourceDiscovery` field for a Document",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `openResourceDiscovery` field for a Document",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.OpenResourceDiscovery = invalidOpenResourceDiscovery

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `baseUrl` of describedSystemInstance Document field",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.DescribedSystemInstance.BaseURL = str.Ptr(invalidUrl)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `ordID` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].OrdID = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].OrdID = invalidOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Title = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = strings.Repeat("a", invalidDescriptionLength)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].ShortDescription = `newLine\n`

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `description` filed for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Description = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `version` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Version = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `version` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Version = invalidVersion

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `policyLevel` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `policyLevel` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = invalidPolicyLevel

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`policyLevel` field for Package is not of type `custom` when `customPolicyLevel` is set",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].CustomPolicyLevel = str.Ptr("myCustomPolicyLevel")
				doc.Packages[0].PolicyLevel = policyLevel

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `type` from `PackageLinks` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToMissingType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `type` key in `PackageLinks` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToWrongType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `url` from `PackageLinks` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToMissingURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `url` key in `PackageLinks` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Field `type` in `PackageLinks` is not set to `custom` when `customType` field is provided",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkTypeWhenProvidedCustomType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `type` set to `custom` in `PackageLinks` when `customType` field is not provided",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidPackageLinkCustomTypeWhenCustomTypeNotProvided)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `PackageLinks` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `PackageLinks` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `PackageLinks` field when it is an empty JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PackageLinks = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field in `Links` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidLinkDueToMissingTitle)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidLinkDueToMissingURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is an empty JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `url` field in `Links` for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Links = json.RawMessage(invalidLinkDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `vendor` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Vendor = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Vendor = str.Ptr(invalidVendor)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `partOfProducts` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when the JSON array is empty",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid element of `partOfProducts` array field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(invalidPartOfProductsElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it contains non string value",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(invalidPartOfProductsIntegerElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage(invalidTagsValue)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when the JSON array is empty",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it contains non string value",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Tags = json.RawMessage(invalidTagsValueIntegerElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage(invalidCountriesElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when JSON array contains non string element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage(invalidCountriesNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when the JSON array is empty",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Countries = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `lineOfBusiness` field element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when JSON array contains non string element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when the JSON array is empty",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].LineOfBusiness = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `industry` field element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(invalidIndustryElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when JSON array contains non string element for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(invalidIndustryNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it is invalid JSON for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it isn't a JSON array for Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when the JSON array is empty",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Industry = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		},

		// Test invalid entity relations

		{
			Name: "Package has a reference to unknown Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].Vendor = str.Ptr(unknownVendorOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Package has a reference to unknown Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PartOfProducts = json.RawMessage(fmt.Sprintf(`["%s"]`, unknownProductOrdID))

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `ordID` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].OrdID = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].OrdID = str.Ptr(invalidOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Name = ""

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `shortDescription` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = str.Ptr(strings.Repeat("a", invalidDescriptionLength))

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = str.Ptr("")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].ShortDescription = str.Ptr(`newLine\n`)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `description` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Exceeded length of `description` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = str.Ptr(strings.Repeat("a", invalidDescriptionLength))

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid empty `description` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = str.Ptr("")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "New lines in `description` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Description = str.Ptr(`newLine\n`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field in `Links` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidBundleLinksDueToMissingTitle)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidBundleLinksDueToMissingURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `url` field in `Links` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidBundleLinksDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `Links` field when it is invalid JSON for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `Links` field when it isn't a JSON array for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `Links` field when it is an empty JSON array for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Links = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid JSON `Labels` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `type` field of `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToMissingType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `type` field of `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`type` field is not with value `custom` when `customType` field is provided for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongCustomType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `customType` field when `type` field is set to `custom` for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongCustomType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`type` field is not with value `custom` when `customDescription` field is provided for `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToMissingCustomType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `callbackURL` field of `CredentialExchangeStrategies` field for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidCredentialsExchangeStrategyDueToWrongCallbackURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `CredentialExchangeStrategies` field when it is invalid JSON for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `CredentialExchangeStrategies` field when it isn't a JSON array for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `CredentialExchangeStrategies` field when it is an empty JSON array for Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `ordID` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdID = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdID = str.Ptr(invalidOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Name = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = str.Ptr(strings.Repeat("a", invalidDescriptionLength))

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = str.Ptr("")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ShortDescription = str.Ptr(`newLine\n`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `description` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Description = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `version` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].VersionInput.Value = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `version` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].VersionInput.Value = invalidVersion

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `partOfPackage` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdPackageID = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfPackage` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdPackageID = str.Ptr(invalidOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `apiProtocol` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ApiProtocol = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `apiProtocol` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ApiProtocol = str.Ptr("wrongApiProtocol")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `visibility` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Visibility = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `visibility` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Visibility = str.Ptr("wrongVisibility")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid element of `partOfProducts` array field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when the JSON array is empty for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it contains non string value for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsIntegerElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `tags` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage(invalidTagsValue)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when the JSON array is empty for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it contains non string value for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Tags = json.RawMessage(invalidTagsValueIntegerElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `countries` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage(invalidCountriesElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when the JSON array is empty for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it contains non string value for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Countries = json.RawMessage(invalidCountriesNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `lineOfBusiness` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when the JSON array is empty for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it contains non string value for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].LineOfBusiness = json.RawMessage(invalidCountriesNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `industry` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(invalidIndustryElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when the JSON array is empty for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it contains non string value for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Industry = json.RawMessage(invalidIndustryNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `type` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `type` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = invalidType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Field `type` value is not `custom` when field `customType` is provided for `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].CustomType = "test:test:v1"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `customType` value when field `type` has value `custom`for `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "custom"
				doc.APIResources[0].ResourceDefinitions[0].CustomType = invalidCustomType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `mediaType` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].MediaType = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].MediaType = invalidMediaType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `openapi-v2` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "openapi-v2"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `openapi-v3` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "openapi-v3"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `raml-v1` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "raml-v1"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `edmx` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "edmx"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `csdl-json` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "csdl-json"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/xml"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `wsdl-v1` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "wsdl-v1"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `wsdl-v2` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "wsdl-v2"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` when field `type` has value `sap-rfc-metadata-v1` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].Type = "sap-rfc-metadata-v1"
				doc.APIResources[0].ResourceDefinitions[0].MediaType = "application/json"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `url` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].URL = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].URL = invalidUrl

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `type` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `type` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = invalidType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = "foo"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `customDescription` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.APIResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomDescription = "foo"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `type` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidApiResourceLinksDueToMissingType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `type` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidApiResourceLinksDueToWrongType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is not `custom` for `apiResourceLink` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidApiResourceLinksDueToMissingCustomValueOfType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `url` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidApiResourceLinksDueToMissingURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `url` field for `apiResourceLink` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidApiResourceLinksDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `apiResourceLink` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `apiResourceLink` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `apiResourceLink` field when it is an empty JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].APIResourceLinks = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field in `Links` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidLinkDueToMissingTitle)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidLinkDueToMissingURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `url` field in `Links` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidLinkDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is an empty JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Links = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `releaseStatus` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `releaseStatus` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("wrongValue")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `sunsetDate` field when `releaseStatus` field has value `deprecated` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `sunsetDate` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("deprecated")
				doc.APIResources[0].SunsetDate = str.Ptr("0000-00-00T09:35:30+0000")
				doc.APIResources[0].Successor = str.Ptr(api2ORDID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `successor` field when `releaseStatus` field has value `deprecated` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `successor` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Successor = str.Ptr("invalidValue")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `version` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingVersion)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `version` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongVersion)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `releaseStatus` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingReleaseStatus)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `releaseStatus` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongReleaseStatus)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `date` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingDate)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `date` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongDate)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of field `changeLogEntries` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it is an empty JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ChangeLogEntries = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `entryPoints` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `entryPoints` when containing invalid URI for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidEntryPointURI)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `entryPoints` when containing duplicate URIs for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidEntryPointsDueToDuplicates)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when the JSON array is empty for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `entryPoints` field when it contains non string value for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].TargetURLs = json.RawMessage(invalidEntryPointsNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `implementationStandard` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr(invalidType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandard` field is valid but `implementationStandard` field is missing for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = nil
				doc.APIResources[0].CustomImplementationStandard = str.Ptr("sap.s4:ATTACHMENT_SRV:v1")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandard` field is valid but `implementationStandard` field is not set to `custom` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].CustomImplementationStandard = str.Ptr("sap.s4:ATTACHMENT_SRV:v1")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `customImplementationStandard` field when `implementationStandard` field is set to `custom` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr("custom")
				doc.APIResources[0].CustomImplementationStandard = str.Ptr(invalidType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `customImplementationStandard` field when `implementationStandard` field is set to `custom` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr("custom")
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandardDescription` is set but `implementationStandard` field is missing for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = nil
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandardDescription` is set but `implementationStandard` field is not `custom` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid when `customImplementationStandardDescription` is set but `implementationStandard` field is not `custom` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].CustomImplementationStandardDescription = str.Ptr("description")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `customImplementationStandardDescription` field when `implementationStandard` field is set to `custom` for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].ImplementationStandard = str.Ptr("custom")
				doc.APIResources[0].CustomImplementationStandard = str.Ptr("sap.s4:ATTACHMENT_SRV:v1")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `ordId` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].BundleOrdID = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordId` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].BundleOrdID = invalidBundleOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Duplicate `ordId` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles = append(doc.APIResources[0].PartOfConsumptionBundles, &model.ConsumptionBundleReference{BundleOrdID: bundleORDID})

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `defaultEntryPoint` field in `PartOfConsumptionBundles` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL = invalidUrl

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `defaultEntryPoint` field from `entryPoints` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfConsumptionBundles[0].DefaultTargetURL = "https://exmaple.com/test/v3"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Present `defaultEntryPoint` field even though there is a single element in `entryPoints` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[1].PartOfConsumptionBundles[0].DefaultTargetURL = "https://exmaple.com/test/v3"

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `supported` field in the `extensible` object for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToNoSupportedProperty)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `supported` field type in the `extensible` object for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedType)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `supported` field value in the `extensible` object for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedValue)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `description` field when `supported` has an `automatic` value",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedAutomaticAndNoDescriptionProperty)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `description` field when `supported` has a `manual` value",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedManualAndNoDescriptionProperty)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Valid `WSDL V1` and `WSDL V2` definitions when APIResources has policyLevel `sap` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSap
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV2
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Valid `WSDL V1` and `WSDL V2` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSapPartner
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV2
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Valid `WSDL V1` and `WSDL V2` definitions when APIResources has policyLevel `sap` and apiProtocol is `soap-outbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSap
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapOutbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapOutbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeWsdlV1
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatApplicationXML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeWsdlV2
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
			ExpectedToBeValid: true,
		},
		{
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSap
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap-partner` and apiProtocol is `soap-inbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSapPartner
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapInbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap` and apiProtocol is `soap-outbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSap
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapOutbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapOutbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `WSDL V1` or `WSDL V2` definition when APIResources has policyLevel `sap-partner` and apiProtocol is `soap-outbound`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSapPartner
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolSoapOutbound
				*doc.APIResources[1].ApiProtocol = open_resource_discovery.ApiProtocolSoapOutbound
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[1].ResourceDefinitions[0].Type = model.APISpecTypeEDMX
				doc.APIResources[1].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationXML
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap` and apiProtocol is `odata-v2`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSap
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolODataV2
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `odata-v2`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSapPartner
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolODataV2
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap` and apiProtocol is `odata-v4`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSap
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolODataV4
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `OpenAPI` and `EDMX` definitions when APIResources has policyLevel `sap-partner` and apiProtocol is `odata-v4`",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = open_resource_discovery.PolicyLevelSapPartner
				*doc.APIResources[0].ApiProtocol = open_resource_discovery.ApiProtocolODataV4
				doc.APIResources[0].ResourceDefinitions[0].Type = model.APISpecTypeOpenAPIV2
				doc.APIResources[0].ResourceDefinitions[0].MediaType = model.SpecFormatApplicationJSON
				doc.APIResources[0].ResourceDefinitions[1].Type = model.APISpecTypeRaml
				doc.APIResources[0].ResourceDefinitions[1].MediaType = model.SpecFormatTextYAML
				doc.APIResources[0].ResourceDefinitions[2] = &model.APIResourceDefinition{}
				return []*open_resource_discovery.Document{doc}
			},
		},
		// Test invalid entity relations

		{
			Name: "API has a reference to an unknown Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].OrdPackageID = str.Ptr(unknownPackageOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "API has a reference to an unknown Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles = fixBundleCreateInput()
				doc.APIResources[0].PartOfConsumptionBundles = fixAPIPartOfConsumptionBundles()
				doc.APIResources[0].PartOfConsumptionBundles[0].BundleOrdID = unknownBundleOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "API has a reference to an unknown Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.APIResources[0].PartOfProducts = json.RawMessage(fmt.Sprintf(`["%s"]`, unknownProductOrdID))

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `ordID` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdID = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordID` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdID = str.Ptr(invalidOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Name = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = str.Ptr(strings.Repeat("a", invalidDescriptionLength))

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid empty `shortDescription` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = str.Ptr("")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ShortDescription = str.Ptr(`newLine\n`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `description` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Description = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `version` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].VersionInput.Value = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `version` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].VersionInput.Value = invalidVersion

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `version` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingVersion)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `version` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongVersion)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `releaseStatus` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingReleaseStatus)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `releaseStatus` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongReleaseStatus)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `date` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToMissingDate)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `date` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongDate)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of field `changeLogEntries` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidChangeLogEntriesDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `changeLogEntries` field when it is an empty JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ChangeLogEntries = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `partOfPackage` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdPackageID = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfPackage` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdPackageID = str.Ptr(invalidOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `visibility` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Visibility = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `visibility` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Visibility = str.Ptr("wrongVisibility")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field in `Links` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidLinkDueToMissingTitle)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `url` field in `Links` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidLinkDueToMissingURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `url` field in `Links` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidLinkDueToWrongURL)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `links` field when it is an empty JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Links = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid element of `partOfProducts` array field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when the JSON array is empty for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `partOfProducts` field when it contains non string value for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(invalidPartOfProductsIntegerElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `type` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `type` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = invalidType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Field `type` value is not `custom` when field `customType` is provided for `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].CustomType = "test:test:v1"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `customType` value when field `type` has value `custom`for `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].Type = "custom"
				doc.EventResources[0].ResourceDefinitions[0].CustomType = invalidCustomType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `mediaType` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].MediaType = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `mediaType` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].MediaType = invalidMediaType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `url` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].URL = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `url` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].URL = invalidUrl

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy = nil

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing field `type` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `type` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = invalidType

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `customType` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomType = "foo"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid field `customDescription` when field `type` is not `custom` for `accessStrategies` of `resourceDefinitions` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].Type = "open"
				doc.EventResources[0].ResourceDefinitions[0].AccessStrategy[0].CustomDescription = "foo"

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `tags` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage(invalidTagsValue)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when the JSON array is empty for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `tags` field when it contains non string value for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Tags = json.RawMessage(invalidTagsValueIntegerElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `countries` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage(invalidCountriesElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when the JSON array is empty for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `countries` field when it contains non string value for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Countries = json.RawMessage(invalidCountriesNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `lineOfBusiness` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(invalidLineOfBusinessElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when the JSON array is empty for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `lineOfBusiness` field when it contains non string value for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].LineOfBusiness = json.RawMessage(invalidCountriesNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `industry` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(invalidIndustryElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it is invalid JSON for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it isn't a JSON array for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when the JSON array is empty for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `industry` field when it contains non string value for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Industry = json.RawMessage(invalidIndustryNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `releaseStatus` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `releaseStatus` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("wrongValue")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `sunsetDate` field when `releaseStatus` field has value `deprecated` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `sunsetDate` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("deprecated")
				doc.EventResources[0].SunsetDate = str.Ptr("0000-00-00T09:35:30+0000")
				doc.EventResources[0].Successor = str.Ptr(event2ORDID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `successor` field when `releaseStatus` field has value `deprecated` for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].ReleaseStatus = str.Ptr("deprecated")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `successor` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Successor = str.Ptr("invalidValue")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `ordId` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles[0].BundleOrdID = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordId` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles[0].BundleOrdID = invalidBundleOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Duplicate `ordId` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles = append(doc.EventResources[0].PartOfConsumptionBundles, &model.ConsumptionBundleReference{BundleOrdID: bundleORDID})

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Present `defaultEntryPoint` field in `PartOfConsumptionBundles` field for Event",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfConsumptionBundles[0].DefaultTargetURL = "https://exmaple.com/test/v3"

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `supported` field in the `extensible` object for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToNoSupportedProperty)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `supported` field type in the `extensible` object for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedType)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid `supported` field value in the `extensible` object for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToInvalidSupportedValue)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `description` field when `supported` has an `automatic` value",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedAutomaticAndNoDescriptionProperty)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Missing `description` field when `supported` has a `manual` value",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].Extensible = json.RawMessage(invalidExtensibleDueToSupportedManualAndNoDescriptionProperty)

				return []*open_resource_discovery.Document{doc}
			},
		},

		// Test invalid entity relations

		{
			Name: "Event has a reference to unknown Package",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].OrdPackageID = str.Ptr(unknownPackageOrdID)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Event has a reference to unknown Bundle",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.ConsumptionBundles = fixBundleCreateInput()
				doc.EventResources[0].PartOfConsumptionBundles = fixEventPartOfConsumptionBundles()
				doc.EventResources[0].PartOfConsumptionBundles[0].BundleOrdID = unknownBundleOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Event has a reference to unknown Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.EventResources[0].PartOfProducts = json.RawMessage(fmt.Sprintf(`["%s"]`, unknownProductOrdID))

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `id` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].OrdID = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `id` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].OrdID = invalidOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Title = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `shortDescription` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].ShortDescription = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Exceeded length of `shortDescription` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].ShortDescription = strings.Repeat("a", invalidDescriptionLength)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "New lines in `shortDescription` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].ShortDescription = `newLine\n`

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `vendor` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Vendor = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `vendor` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Vendor = invalidOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `parent` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Parent = str.Ptr(invalidType)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid value for `correlationIds` field for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIds = json.RawMessage(invalidCorrelationIdsElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it is invalid JSON for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIds = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it isn't a JSON array for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIds = json.RawMessage("{}")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when the JSON array is empty for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIds = json.RawMessage("[]")

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `correlationIds` field when it contains non string value for API",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].CorrelationIds = json.RawMessage(invalidCorrelationIdsNonStringElement)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*open_resource_discovery.Document{doc}
			},
		},

		// Test invalid entity relations

		{
			Name: "Product has a reference to unknown Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Vendor = unknownVendorOrdID

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `id` field for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `id` field for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].OrdID = invalidOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `title` field for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Title = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON `Labels` field for Product",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Products[0].Labels = json.RawMessage(invalidJson)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid JSON object `Labels` field for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidLabelsWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Labels` values are not array of strings for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidLabelsWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid key for JSON `Labels` field for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidLabelsWhenKeyIsWrong)

				return []*open_resource_discovery.Document{doc}
			},
		},
		{
			Name: "Invalid JSON object `Partners` field for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Partners = json.RawMessage(`[]`)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Partners` values are not array for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidPartnersWhenValueIsNotArray)

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "`Partners` values are not array of strings for Vendor",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Vendors[0].Labels = json.RawMessage(invalidPartnersWhenValuesAreNotArrayOfStrings)

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
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
		DocumentProvider  func() []*open_resource_discovery.Document
		ExpectedToBeValid bool
	}{
		{
			Name: "Missing `ordId` field for Tombstone",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `ordId` field for Tombstone",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].OrdID = invalidOrdID

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Missing `removalDate` field for Tombstone",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].RemovalDate = ""

				return []*open_resource_discovery.Document{doc}
			},
		}, {
			Name: "Invalid `removalDate` field for Tombstone",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Tombstones[0].RemovalDate = "0000-00-00T15:04:05Z"

				return []*open_resource_discovery.Document{doc}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			docs := open_resource_discovery.Documents{test.DocumentProvider()[0]}
			err := docs.Validate(baseURL)
			if test.ExpectedToBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
