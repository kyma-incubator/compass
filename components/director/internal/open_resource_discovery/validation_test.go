package open_resource_discovery_test

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	invalidOpenResourceDiscovery = "invalidOpenResourceDiscovery"
	invalidUrl                   = "invalidUrl"
	invalidOrdID                 = "invalidOrdId"
	invalidDescriptionLength     = 256
	invalidVersion               = "invalidVersion"
	invalidPolicyLevel           = "invalidPolicyLevel"
	customPolicyLevel            = "custom"
	invalidVendor                = "wrongVendor!"
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
)

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
			Name: "`policyLevel` field for Package is set to `custom` but `customPolicyLevel` field is nil",
			DocumentProvider: func() []*open_resource_discovery.Document {
				doc := fixORDDocument()
				doc.Packages[0].PolicyLevel = customPolicyLevel

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
