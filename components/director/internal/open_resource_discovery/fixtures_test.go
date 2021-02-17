package open_resource_discovery_test

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"strings"
)

const (
	ordDocURI     = "/open-resource-discovery/v1/documents/example1"
	baseURL       = "http://localhost:8080"
	packageORDID  = "ns:package:PACKAGE_ID:v1"
	productORDID  = "ns:PRODUCT_ID"
	product2ORDID = "ns:PRODUCT_ID2"
	bundleORDID   = "ns:consumptionBundle:BUNDLE_ID:v1"
	vendorORDID   = "sap"
	api1ORDID     = "ns:apiResource:API_ID:v2"
	api2ORDID     = "ns:apiResource:API_ID2:v1"
	event1ORDID   = "ns:eventResource:EVENT_ID:v1"
	event2ORDID   = "ns2:eventResource:EVENT_ID:v1"
)

var (
	packageLinks = removeWhitespace(`[
        {
          "type": "terms-of-service",
          "url": "https://example.com/en/legal/terms-of-use.html"
        },
        {
          "type": "client-registration",
          "url": "/ui/public/showRegisterForm"
        }
      ]`)

	links = removeWhitespace(`[
        {
          "title": "Link Title",
          "description": "lorem ipsum dolor nem",
          "url": "https://example.com/2018/04/11/testing/"
        },
		{
          "title": "Link Title",
          "description": "lorem ipsum dolor nem",
          "url": "/testing/relative"
        }
      ]`)

	labels = removeWhitespace(`{
        "label-key-1": [
          "label-value-1",
          "label-value-2"
        ]
      }`)

	credentialExchangeStrategies = removeWhitespace(`[
        {
          "type": "custom",
          "customType": "ns:credential-exchange:v1",
          "callbackUrl": "/credentials/relative"
        },
		{
          "type": "custom",
          "customType": "ns:credential-exchange2:v3",
          "callbackUrl": "http://example.com/credentials"
        }
      ]`)

	apiResourceLinks = removeWhitespace(`[
        {
          "type": "console",
          "url": "https://example.com/shell/discover"
        },
		{
          "type": "console",
          "url": "/shell/discover/relative"
        }
      ]`)

	changeLogEntries = removeWhitespace(`[
        {
          "version": "1.0.0",
          "releaseStatus": "active",
          "date": "2020-04-29",
          "description": "lorem ipsum dolor sit amet",
          "url": "https://example.com/changelog/v1"
        }
      ]`)
)

func fixWellKnownConfig() *open_resource_discovery.WellKnownConfig {
	return &open_resource_discovery.WellKnownConfig{
		Schema: "../spec/v1/generated/Configuration.schema.json",
		OpenResourceDiscoveryV1: open_resource_discovery.OpenResourceDiscoveryV1{
			Documents: []open_resource_discovery.DocumentDetails{
				{
					URL:                 ordDocURI,
					SystemInstanceAware: true,
					AccessStrategies: []open_resource_discovery.AccessStrategy{
						{
							Type: open_resource_discovery.OpenAccessStrategy,
						},
					},
				},
			},
		},
	}
}

func fixORDDocument() *open_resource_discovery.Document {
	true := true
	return &open_resource_discovery.Document{
		Schema:                "./spec/v1/generated/Document.schema.json",
		OpenResourceDiscovery: "v1",
		Description:           "Test Document",
		SystemInstanceAware:   true,
		DescribedSystemInstance: &model.Application{
			BaseURL: str.Ptr(baseURL),
			Labels:  json.RawMessage(labels),
		},
		ProviderSystemInstance: nil,
		Packages: []*model.PackageInput{
			{
				OrdID:            packageORDID,
				Vendor:           str.Ptr(vendorORDID),
				Title:            "PACKAGE 1 TITLE",
				ShortDescription: "lorem ipsum",
				Description:      "lorem ipsum dolor set",
				Version:          "1.1.2",
				PackageLinks:     json.RawMessage(packageLinks),
				Links:            json.RawMessage(links),
				LicenseType:      str.Ptr("licence"),
				Tags:             json.RawMessage(`["testTag"]`),
				Countries:        json.RawMessage(`["BG","EN"]`),
				Labels:           json.RawMessage(labels),
				PolicyLevel:      "sap",
				PartOfProducts:   json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:   json.RawMessage(`["lineOfBusiness"]`),
				Industry:         json.RawMessage(`["automotive","finance"]`),
			},
		},
		ConsumptionBundles: []*model.BundleCreateInput{
			{
				Name:                         "BUNDLE TITLE",
				Description:                  str.Ptr("lorem ipsum dolor nsq sme"),
				OrdID:                        str.Ptr(bundleORDID),
				ShortDescription:             str.Ptr("lorem ipsum"),
				Links:                        json.RawMessage(links),
				Labels:                       json.RawMessage(labels),
				CredentialExchangeStrategies: json.RawMessage(credentialExchangeStrategies),
			},
		},
		Products: []*model.ProductInput{
			{
				OrdID:            productORDID,
				Title:            "PRODUCT TITLE",
				ShortDescription: "lorem ipsum",
				Vendor:           vendorORDID,
				Parent:           str.Ptr(product2ORDID),
				PPMSObjectID:     str.Ptr("12391293812"),
				Labels:           json.RawMessage(labels),
			},
		},
		APIResources: []*model.APIDefinitionInput{
			{
				OrdID:               str.Ptr(api1ORDID),
				OrdBundleID:         str.Ptr(bundleORDID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "API TITLE",
				Description:         str.Ptr("lorem ipsum dolor sit amet"),
				TargetURL:           "/test/v1",
				ShortDescription:    str.Ptr("lorem ipsum"),
				SystemInstanceAware: nil,
				ApiProtocol:         str.Ptr("odata-v2"),
				Tags:                json.RawMessage(`["apiTestTag"]`),
				Countries:           json.RawMessage(`["BG","US"]`),
				Links:               json.RawMessage(links),
				APIResourceLinks:    json.RawMessage(apiResourceLinks),
				ReleaseStatus:       str.Ptr("active"),
				SunsetDate:          nil,
				Successor:           nil,
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				Labels:              json.RawMessage(labels),
				Visibility:          str.Ptr("public"),
				Disabled:            &true,
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:      json.RawMessage(`["lineOfBusiness2"]`),
				Industry:            json.RawMessage(`["automotive","test"]`),
				ResourceDefinitions: []*model.APIResourceDefinition{
					{
						Type:      "openapi-v3",
						MediaType: "application/json",
						URL:       "/odata/1.0/catalog.svc/$value?type=json",
						AccessStrategy: []model.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
					{
						Type:      "openapi-v3",
						MediaType: "text/yaml",
						URL:       "https://test.com/odata/1.0/catalog",
						AccessStrategy: []model.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
			},
			{
				OrdID:               str.Ptr(api2ORDID),
				OrdBundleID:         str.Ptr(bundleORDID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "Gateway Sample Service",
				Description:         str.Ptr("lorem ipsum dolor sit amet"),
				TargetURL:           "/some-api/v1",
				ShortDescription:    str.Ptr("lorem ipsum"),
				SystemInstanceAware: nil,
				ApiProtocol:         str.Ptr("odata-v2"),
				Tags:                json.RawMessage(`["ZGWSAMPLE"]`),
				Countries:           json.RawMessage(`["BR"]`),
				Links:               json.RawMessage(links),
				APIResourceLinks:    json.RawMessage(apiResourceLinks),
				ReleaseStatus:       str.Ptr("deprecated"),
				SunsetDate:          str.Ptr("2020-12-08T15:47:04+0000"),
				Successor:           str.Ptr(api1ORDID),
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				Labels:              json.RawMessage(labels),
				Visibility:          str.Ptr("public"),
				Disabled:            nil,
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:      json.RawMessage(`["lineOfBusiness2"]`),
				Industry:            json.RawMessage(`["automotive","test"]`),
				ResourceDefinitions: []*model.APIResourceDefinition{
					{
						Type:      "edmx",
						MediaType: "application/xml",
						URL:       "https://TEST:443//odata/$metadata",
						AccessStrategy: []model.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				VersionInput: &model.VersionInput{
					Value: "1.1.0",
				},
			},
		},
		EventResources: []*model.EventDefinitionInput{
			{
				OrdID:               str.Ptr(event1ORDID),
				OrdBundleID:         str.Ptr(bundleORDID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "EVENT TITLE",
				Description:         str.Ptr("lorem ipsum dolor sit amet"),
				ShortDescription:    str.Ptr("lorem ipsum"),
				SystemInstanceAware: nil,
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				Links:               json.RawMessage(links),
				Tags:                json.RawMessage(`["eventTestTag"]`),
				Countries:           json.RawMessage(`["BG","US"]`),
				ReleaseStatus:       str.Ptr("active"),
				SunsetDate:          nil,
				Successor:           nil,
				Labels:              json.RawMessage(labels),
				Visibility:          str.Ptr("public"),
				Disabled:            &true,
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:      json.RawMessage(`["lineOfBusiness2"]`),
				Industry:            json.RawMessage(`["automotive","test"]`),
				ResourceDefinitions: []*model.EventResourceDefinition{
					{
						Type:      "asyncapi-v2",
						MediaType: "application/json",
						URL:       "http://localhost:8080/asyncApi2.json",
						AccessStrategy: []model.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
			},
			{
				OrdID:               str.Ptr(event2ORDID),
				OrdBundleID:         str.Ptr(bundleORDID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "EVENT TITLE 2",
				Description:         str.Ptr("lorem ipsum dolor sit amet"),
				ShortDescription:    str.Ptr("lorem ipsum"),
				SystemInstanceAware: nil,
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				Links:               json.RawMessage(links),
				Tags:                json.RawMessage(`["eventTestTag2"]`),
				Countries:           json.RawMessage(`["BR"]`),
				ReleaseStatus:       str.Ptr("deprecated"),
				SunsetDate:          str.Ptr("2020-12-08T15:47:04+0000"),
				Successor:           str.Ptr(event2ORDID),
				Labels:              json.RawMessage(labels),
				Visibility:          str.Ptr("public"),
				Disabled:            nil,
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:      json.RawMessage(`["lineOfBusiness2"]`),
				Industry:            json.RawMessage(`["automotive","test"]`),
				ResourceDefinitions: []*model.EventResourceDefinition{
					{
						Type:      "asyncapi-v2",
						MediaType: "application/json",
						URL:       "/api/eventCatalog.json",
						AccessStrategy: []model.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				VersionInput: &model.VersionInput{
					Value: "1.1.0",
				},
			},
		},
		Tombstones: []*model.TombstoneInput{
			{
				OrdID:       api2ORDID,
				RemovalDate: "2020-12-02T14:12:59Z",
			},
		},
		Vendors: []*model.VendorInput{
			{
				OrdID:  vendorORDID,
				Title:  "SAP",
				Type:   "sap",
				Labels: json.RawMessage(labels),
			},
		},
	}
}

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\n", ""), "\t", "")
}
