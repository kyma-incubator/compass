package ord_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	absoluteDocURL              = "http://config.com/open-resource-discovery/v1/documents/example1"
	ordDocURI                   = "/open-resource-discovery/v1/documents/example1"
	proxyURL                    = "http://proxy.com:8080"
	baseURL                     = "http://test.com:8080"
	baseURL2                    = "http://second.com"
	customWebhookConfigURL      = "http://custom.com/config/endpoint"
	packageORDID                = "ns:package:PACKAGE_ID:v1"
	productORDID                = "sap:product:id:"
	globalProductORDID          = "sap:product:SAPCloudPlatform:"
	product2ORDID               = "ns:product:id2:"
	bundleORDID                 = "ns:consumptionBundle:BUNDLE_ID:v1"
	secondBundleORDID           = "ns:consumptionBundle:BUNDLE_ID:v2"
	vendorORDID                 = "sap:vendor:SAP:"
	vendor2ORDID                = "partner:vendor:SAP:"
	api1ORDID                   = "ns:apiResource:API_ID:v2"
	api2ORDID                   = "ns:apiResource:API_ID2:v1"
	event1ORDID                 = "ns:eventResource:EVENT_ID:v1"
	event2ORDID                 = "ns2:eventResource:EVENT_ID:v1"
	entityType1ORDID            = "ns:entityType:ENTITYTYPE_ID:v1"
	entityType2ORDID            = "ns2:entityType:ENTITYTYPE_ID:v1"
	capability1ORDID            = "sap.foo.bar:capability:fieldExtensibility:v1"
	capability2ORDID            = "sap2.foo.bar:capability:fieldExtensibility:v1"
	integrationDependency1ORDID = "sap.foo.bar:integrationDependency:CustomerOrder:v1"
	integrationDependency2ORDID = "sap2.foo.bar:integrationDependency:CustomerOrder:v1"

	whID                     = "testWh"
	tenantID                 = "testTenant"
	externalTenantID         = "externalTestTenant"
	packageID                = "testPkg"
	vendorID                 = "testVendor"
	vendorID2                = "testVendor2"
	productID                = "testProduct"
	bundleID                 = "testBndl"
	api1ID                   = "testAPI1"
	api2ID                   = "testAPI2"
	event1ID                 = "testEvent1"
	event2ID                 = "testEvent2"
	capability1ID            = "testCapability1"
	capability2ID            = "testCapability2"
	integrationDependency1ID = "testIntegrationDependency1"
	integrationDependency2ID = "testIntegrationDependency2"
	tombstoneID              = "testTs"
	localTenantID            = "localTenantID"
	webhookID                = "webhookID"

	api1spec1ID       = "api1spec1ID"
	api1spec2ID       = "api1spec2ID"
	api1spec3ID       = "api1spec3ID"
	api2spec1ID       = "api2spec1ID"
	api2spec2ID       = "api2spec2ID"
	event1specID      = "event1specID"
	event2specID      = "event2specID"
	capability1SpecID = "capability1SpecID"
	capability2SpecID = "capability2SpecID"

	cursor                    = "cursor"
	apiImplementationStandard = "cff:open-service-broker:v2"
	correlationIDs            = `["foo.bar.baz:foo:123456","foo.bar.baz:bar:654321"]`
	partners                  = `["microsoft:vendor:Microsoft:"]`

	externalClientCertSecretName = "resource-name1"
	extSvcClientCertSecretName   = "resource-name2"

	appTemplateVersionValue = "2303"
	appTemplateName         = "appTemplateName"

	applicationTypeLabelValue = "customType"

	entityTypeID        = "entity-type-id"
	entityTypeMappingID = "entity-type-mapping-id"
	ordID               = "com.compass.v1"
	level               = "aggregate"
	title               = "BusinessPartner"
	publicVisibility    = "public"
	products            = `["sap:product:S4HANA_OD:"]`
	releaseStatus       = "active"
)

var (
	appID                = "testApp"
	appTemplateVersionID = "testAppTemplateVersionID"
	appTemplateID        = "testAppTemplate"
	policyLevel          = "sap:core:v1"
	customPolicyLevel    = "sap:core:v1"
	uidSvc               = uid.NewService()
	packageLinksFormat   = removeWhitespace(`[
        {
          "type": "terms-of-service",
          "url": "https://example.com/en/legal/terms-of-use.html"
        },
        {
          "type": "client-registration",
          "url": "%s/ui/public/showRegisterForm"
        }
      ]`)

	linksFormat = removeWhitespace(`[
        {
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 1",
          "url": "https://example.com/2018/04/11/testing/"
        },
		{
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 2",
          "url": "%s/testing/relative"
        }
      ]`)

	packageLabels = removeWhitespace(`{
        "label-key-1": [
          "label-val"
        ],
		"pkg-label": [
          "label-val"
        ]
      }`)

	labels = removeWhitespace(`{
        "label-key-1": [
          "label-value-1",
          "label-value-2"
        ]
      }`)

	mergedLabels = removeWhitespace(`{
        "label-key-1": [
          "label-val",
		  "label-value-1",
          "label-value-2"
        ],
		"pkg-label": [
          "label-val"
        ]
      }`)

	tags = removeWhitespace(`[
        "testTag"
      ]`)

	documentLabels = removeWhitespace(`{
        "Some Aspect": ["Markdown Documentation [with links](#)", "With multiple values"]
      }`)

	supportedUseCases = removeWhitespace(`[
        "mass-extraction"
      ]`)

	credentialExchangeStrategiesWithCustomTypeFormat = removeWhitespace(`[
		{
		  "callbackUrl": "http://example.com/credentials",
          "customType": "%s",
		  "type": "custom",
		  "customDescription": "description"
        }
      ]`)

	credentialExchangeStrategiesWithMultipleSameTypesFormat = removeWhitespace(`[
		{
		  "callbackUrl": "http://example.com/credentials-fake",
          "customType": "%s",
		  "type": "custom",
		  "customDescription": "description"
        },
        {
		  "callbackUrl": "http://example.com/credentials",
          "customType": "%s",
		  "type": "custom",
		  "customDescription": "description"
        }
      ]`)

	credentialExchangeStrategiesFormat = removeWhitespace(`[
        {
		  "callbackUrl": "%s/credentials/relative",
          "customType": "ns:credential-exchange:v1",
		  "type": "custom"
        },
		{
		  "callbackUrl": "http://example.com/credentials",
          "customType": "ns:credential-exchange2:v3",
		  "type": "custom"
        },
		{
		  "callbackUrl": "http://example.com/credentials",
          "customType": "%s",
		  "type": "custom"
        }
      ]`)

	credentialExchangeStrategiesBasic = removeWhitespace(`[
		{
		  "callbackUrl": "http://example.com/credentials",
          "customType": "ns:credential-exchange2:v3",
		  "type": "custom"
        }
      ]`)

	resourceLinksFormat = removeWhitespace(`[
        {
          "type": "console",
          "url": "https://example.com/shell/discover"
        },
		{
          "type": "console",
          "url": "%s/shell/discover/relative"
        }
      ]`)

	changeLogEntries = removeWhitespace(`[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`)

	apiAPIModelSelectors = removeWhitespace(`[
		{
		  "type": "json-pointer",
		  "jsonPointer": "#/objects/schemas/WorkForcePersonRead"
		},
		{
		  "type": "json-pointer",
		  "jsonPointer": "#/objects/schemas/WorkForcePersonUpdate"
		},
		{
		  "type": "json-pointer",
		  "jsonPointer": "#/objects/schemas/WorkForcePersonCreate"
		}
	  ]`)

	apiEntityTypeTargets = removeWhitespace(`[
		{
		  "ordId": "sap.odm:entityType:WorkforcePerson:v1"
		},
		{
		  "correlationId": "sap.s4:csnEntity:WorkForcePersonView_v1"
		}
	  ]`)

	eventAPIModelSelectors = removeWhitespace(`[
		{
		  "type": "json-pointer",
		  "jsonPointer": "#/components/messages/sap_odm_finance_costobject_CostCenter_Created_v1/payload"
		}
	  ]`)

	eventEntityTypeTargets = removeWhitespace(`[
		{
		  "ordId": "sap.odm:entityType:CostCenter:v1"
		},
		{
		  "correlationId": "sap.s4:csnEntity:CostCenter_v1"
		}
	  ]`)

	boolPtr = true

	apisFromDB = map[string]*model.APIDefinition{
		api1ORDID: fixAPIsWithHash()[0],
		api2ORDID: fixAPIsWithHash()[1],
	}

	eventsFromDB = map[string]*model.EventDefinition{
		event1ORDID: fixEventsWithHash()[0],
		event2ORDID: fixEventsWithHash()[1],
	}

	capabilitiesFromDB = map[string]*model.Capability{
		capability1ORDID: fixCapabilitiesWithHash()[0],
		capability2ORDID: fixCapabilitiesWithHash()[1],
	}

	integrationDependenciesFromDB = map[string]*model.IntegrationDependency{
		integrationDependency1ORDID: fixIntegrationDependenciesWithHash()[0],
		integrationDependency2ORDID: fixIntegrationDependenciesWithHash()[1],
	}

	pkgsFromDB = map[string]*model.Package{
		packageORDID: fixPackagesWithHash()[0],
	}

	bndlsFromDB = map[string]*model.Bundle{
		bundleORDID: fixBundlesWithHash()[0],
	}

	resourceHashes = fixResourceHashes()

	credentialExchangeStrategyType           = "sap.ucl:tenant-mapping:v1"
	credentialExchangeStrategyVersion        = "v1"
	credentialExchangeStrategyTenantMappings = map[string]ord.CredentialExchangeStrategyTenantMapping{
		credentialExchangeStrategyType: {
			Mode:    model.WebhookModeSync,
			Version: credentialExchangeStrategyVersion,
		},
	}

	fixedTimestamp         = time.Now()
	shortDescription       = "A business partner is a person, an organization, or a group of persons or organizations in which a company has a business interest."
	description            = "A workforce person is a natural person with a work agreement or relationship in form of a work assignment; it can be an employee or a contingent worker.\n"
	systemInstanceAware    = false
	sunsetDate             = "2022-01-08T15:47:04+00:00"
	successors             = `["sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1"]`
	extensible             = `{"supported":"automatic","description":"Please find the extensibility documentation"}`
	resourceHash           = "123456"
	versionValue           = "v1.1"
	versionDeprecated      = false
	versionDeprecatedSince = "v1.0"
	versionForRemoval      = false
	mandatoryTrue          = true

	documentationLabels = removeWhitespace(`{
        "Some Aspect": ["Markdown Documentation [with links](#)", "With multiple values"]
      }`)
)

func fixResourceHashes() map[string]uint64 {
	return fixResourceHashesForDocument(fixORDDocument())
}

func fixResourceHashesForDocument(doc *ord.Document) map[string]uint64 {
	result := map[string]uint64{}
	for _, resource := range doc.APIResources {
		hash, _ := ord.HashObject(resource)
		result[*resource.OrdID] = hash
	}
	for _, resource := range doc.EventResources {
		hash, _ := ord.HashObject(resource)
		result[*resource.OrdID] = hash
	}
	for _, resource := range doc.Capabilities {
		hash, _ := ord.HashObject(resource)
		result[*resource.OrdID] = hash
	}
	for _, resource := range doc.IntegrationDependencies {
		hash, _ := ord.HashObject(resource)
		result[*resource.OrdID] = hash
	}
	for _, resource := range doc.Packages {
		hash, _ := ord.HashObject(resource)
		result[resource.OrdID] = hash
	}
	for _, resource := range doc.EntityTypes {
		hash, _ := ord.HashObject(resource)
		result[resource.OrdID] = hash
	}
	for _, resource := range doc.ConsumptionBundles {
		hash, _ := ord.HashObject(resource)
		result[*resource.OrdID] = hash
	}

	return result
}

func fixWellKnownConfig() *ord.WellKnownConfig {
	return &ord.WellKnownConfig{
		Schema:  "../spec/v1/generated/Configuration.schema.json",
		BaseURL: baseURL,
		OpenResourceDiscoveryV1: ord.OpenResourceDiscoveryV1{
			Documents: []ord.DocumentDetails{
				{
					URL:                 ordDocURI,
					SystemInstanceAware: true,
					AccessStrategies: []accessstrategy.AccessStrategy{
						{
							Type: accessstrategy.OpenAccessStrategy,
						},
					},
				},
			},
		},
	}
}

func fixORDDocument() *ord.Document {
	return fixORDDocumentWithBaseURL("")
}

func fixORDDocumentWithoutCredentialExchanges() *ord.Document {
	doc := fixORDDocumentWithBaseURL("")
	doc.ConsumptionBundles[0].CredentialExchangeStrategies = nil
	return doc
}

func fixORDStaticDocument() *ord.Document {
	doc := fixORDDocumentWithBaseURL("")
	doc.DescribedSystemInstance = nil
	doc.DescribedSystemVersion = fixAppTemplateVersionInput()
	doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(credentialExchangeStrategiesBasic)

	return doc
}

func fixSanitizedORDDocument() *ord.Document {
	sanitizedDoc := fixORDDocumentWithBaseURL(baseURL)
	sanitizeResources(sanitizedDoc)
	return sanitizedDoc
}

func fixSanitizedORDDocumentForProxyURL() *ord.Document {
	sanitizedDoc := fixORDDocumentWithBaseURL(proxyURL)
	sanitizedDoc.ConsumptionBundles[0].CredentialExchangeStrategies = nil
	sanitizeResources(sanitizedDoc)
	return sanitizedDoc
}

func fixSanitizedStaticORDDocument() *ord.Document {
	sanitizedDoc := fixORDStaticDocumentWithBaseURL(baseURL)
	sanitizeResources(sanitizedDoc)
	return sanitizedDoc
}

func sanitizeResources(doc *ord.Document) {
	doc.Packages[0].PolicyLevel = str.Ptr(policyLevel)

	doc.APIResources[0].PolicyLevel = str.Ptr(policyLevel)
	doc.APIResources[0].Tags = json.RawMessage(`["testTag","apiTestTag"]`)
	doc.APIResources[0].Countries = json.RawMessage(`["BG","EN","US"]`)
	doc.APIResources[0].LineOfBusiness = json.RawMessage(`["Finance","Sales"]`)
	doc.APIResources[0].Industry = json.RawMessage(`["Automotive","Banking","Chemicals"]`)
	doc.APIResources[0].Labels = json.RawMessage(mergedLabels)

	doc.APIResources[1].PolicyLevel = str.Ptr(policyLevel)
	doc.APIResources[1].Tags = json.RawMessage(`["testTag","ZGWSAMPLE"]`)
	doc.APIResources[1].Countries = json.RawMessage(`["BG","EN","BR"]`)
	doc.APIResources[1].LineOfBusiness = json.RawMessage(`["Finance","Sales"]`)
	doc.APIResources[1].Industry = json.RawMessage(`["Automotive","Banking","Chemicals"]`)
	doc.APIResources[1].Labels = json.RawMessage(mergedLabels)

	doc.EventResources[0].PolicyLevel = str.Ptr(policyLevel)
	doc.EventResources[0].Tags = json.RawMessage(`["testTag","eventTestTag"]`)
	doc.EventResources[0].Countries = json.RawMessage(`["BG","EN","US"]`)
	doc.EventResources[0].LineOfBusiness = json.RawMessage(`["Finance","Sales"]`)
	doc.EventResources[0].Industry = json.RawMessage(`["Automotive","Banking","Chemicals"]`)
	doc.EventResources[0].Labels = json.RawMessage(mergedLabels)

	doc.EventResources[1].PolicyLevel = str.Ptr(policyLevel)
	doc.EventResources[1].Tags = json.RawMessage(`["testTag","eventTestTag2"]`)
	doc.EventResources[1].Countries = json.RawMessage(`["BG","EN","BR"]`)
	doc.EventResources[1].LineOfBusiness = json.RawMessage(`["Finance","Sales"]`)
	doc.EventResources[1].Industry = json.RawMessage(`["Automotive","Banking","Chemicals"]`)
	doc.EventResources[1].Labels = json.RawMessage(mergedLabels)

	doc.EntityTypes[0].PolicyLevel = str.Ptr(policyLevel)
	doc.EntityTypes[0].Tags = json.RawMessage(`["testTag","eventTestTag"]`)
	doc.EntityTypes[0].Labels = json.RawMessage(mergedLabels)

	doc.EntityTypes[1].PolicyLevel = str.Ptr(policyLevel)
	doc.EntityTypes[1].Tags = json.RawMessage(`["testTag","eventTestTag"]`)
	doc.EntityTypes[1].Labels = json.RawMessage(mergedLabels)

	doc.Capabilities[0].Tags = json.RawMessage(`["testTag","capabilityTestTag"]`)
	doc.Capabilities[0].Labels = json.RawMessage(mergedLabels)
	doc.Capabilities[1].Tags = json.RawMessage(`["testTag","capabilityTestTag"]`)
	doc.Capabilities[1].Labels = json.RawMessage(mergedLabels)

	doc.IntegrationDependencies[0].Tags = json.RawMessage(`["testTag","integrationDependencyTestTag"]`)
	doc.IntegrationDependencies[0].Labels = json.RawMessage(mergedLabels)
	doc.IntegrationDependencies[1].Tags = json.RawMessage(`["testTag","integrationDependencyTestTag"]`)
	doc.IntegrationDependencies[1].Labels = json.RawMessage(mergedLabels)
}

func fixORDDocumentWithBaseURL(providedBaseURL string) *ord.Document {
	return &ord.Document{
		Schema:                "./spec/v1/generated/Document.schema.json",
		OpenResourceDiscovery: "1.0",
		Description:           "Test Document",
		Perspective:           ord.SystemInstancePerspective,
		DescribedSystemInstance: &model.Application{
			BaseURL:             str.Ptr(baseURL),
			OrdLabels:           json.RawMessage(labels),
			Tags:                json.RawMessage(tags),
			DocumentationLabels: json.RawMessage(documentLabels),
		},
		PolicyLevel: str.Ptr(policyLevel),
		Packages: []*model.PackageInput{
			{
				OrdID:               packageORDID,
				Vendor:              str.Ptr(vendorORDID),
				Title:               "PACKAGE 1 TITLE",
				ShortDescription:    "short desc",
				Description:         "longer desc",
				Version:             "1.1.2",
				PackageLinks:        json.RawMessage(fmt.Sprintf(packageLinksFormat, providedBaseURL)),
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				LicenseType:         str.Ptr("licence"),
				SupportInfo:         str.Ptr("support-info"),
				Tags:                json.RawMessage(tags),
				Countries:           json.RawMessage(`["BG","EN"]`),
				Labels:              json.RawMessage(packageLabels),
				DocumentationLabels: json.RawMessage(documentLabels),
				PolicyLevel:         nil,
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:      json.RawMessage(`["Finance","Sales"]`),
				Industry:            json.RawMessage(`["Automotive","Banking","Chemicals"]`),
			},
		},
		ConsumptionBundles: []*model.BundleCreateInput{
			{
				Name:                         "BUNDLE TITLE",
				Description:                  str.Ptr("lorem ipsum dolor nsq sme"),
				Version:                      str.Ptr("1.1.2"),
				OrdID:                        str.Ptr(bundleORDID),
				LocalTenantID:                str.Ptr(localTenantID),
				ShortDescription:             str.Ptr("lorem ipsum"),
				Links:                        json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				Tags:                         json.RawMessage(tags),
				Labels:                       json.RawMessage(labels),
				DocumentationLabels:          json.RawMessage(documentLabels),
				CredentialExchangeStrategies: json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesFormat, providedBaseURL, credentialExchangeStrategyType)),
				CorrelationIDs:               json.RawMessage(correlationIDs),
			},
		},
		Products: []*model.ProductInput{
			{
				OrdID:               productORDID,
				Title:               "PRODUCT TITLE",
				ShortDescription:    "short desc",
				Description:         str.Ptr("long desc"),
				Vendor:              vendorORDID,
				Parent:              str.Ptr(product2ORDID),
				CorrelationIDs:      json.RawMessage(correlationIDs),
				Tags:                json.RawMessage(tags),
				Labels:              json.RawMessage(labels),
				DocumentationLabels: json.RawMessage(documentLabels),
			},
		},
		APIResources: []*model.APIDefinitionInput{
			{
				OrdID:                                   str.Ptr(api1ORDID),
				LocalTenantID:                           str.Ptr(localTenantID),
				OrdPackageID:                            str.Ptr(packageORDID),
				Name:                                    "API TITLE",
				Description:                             str.Ptr("long desc"),
				TargetURLs:                              json.RawMessage(`["https://exmaple.com/test/v1","https://exmaple.com/test/v2"]`),
				ShortDescription:                        str.Ptr("short desc"),
				SystemInstanceAware:                     &boolPtr,
				APIProtocol:                             str.Ptr("odata-v2"),
				Tags:                                    json.RawMessage(`["apiTestTag"]`),
				Countries:                               json.RawMessage(`["BG","US"]`),
				Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				APIResourceLinks:                        json.RawMessage(fmt.Sprintf(resourceLinksFormat, providedBaseURL)),
				ReleaseStatus:                           str.Ptr("active"),
				SunsetDate:                              nil,
				Successors:                              nil,
				ChangeLogEntries:                        json.RawMessage(changeLogEntries),
				Labels:                                  json.RawMessage(labels),
				SupportedUseCases:                       json.RawMessage(supportedUseCases),
				DocumentationLabels:                     json.RawMessage(documentLabels),
				Visibility:                              str.Ptr("public"),
				Disabled:                                &boolPtr,
				PartOfProducts:                          json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:                          json.RawMessage(`["Finance","Sales"]`),
				Industry:                                json.RawMessage(`["Automotive","Banking","Chemicals"]`),
				ImplementationStandard:                  str.Ptr(apiImplementationStandard),
				CustomImplementationStandard:            nil,
				CustomImplementationStandardDescription: nil,
				Extensible:                              json.RawMessage(`{"supported":"automatic","description":"Please find the extensibility documentation"}`),
				LastUpdate:                              str.Ptr("2023-01-26T15:47:04+00:00"),
				ResourceDefinitions: []*model.APIResourceDefinition{
					{
						Type:      "openapi-v3",
						MediaType: "application/json",
						URL:       fmt.Sprintf("%s/external-api/unsecured/spec/flapping", providedBaseURL),
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
					{
						Type:      "openapi-v3",
						MediaType: "text/yaml",
						URL:       "https://test.com/odata/1.0/catalog",
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
					{
						Type:      "edmx",
						MediaType: "application/xml",
						URL:       "https://TEST:443//odata/$metadata",
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				PartOfConsumptionBundles: []*model.ConsumptionBundleReference{
					{
						BundleOrdID:      bundleORDID,
						DefaultTargetURL: "https://exmaple.com/test/v1",
					},
				},
				EntityTypeMappings: []*model.EntityTypeMappingInput{
					{
						APIModelSelectors: json.RawMessage(apiAPIModelSelectors),
						EntityTypeTargets: json.RawMessage(apiEntityTypeTargets),
					},
				},
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
				Direction: str.Ptr("mixed"),
			},
			{
				Extensible:                              json.RawMessage(`{"supported":"automatic","description":"Please find the extensibility documentation"}`),
				OrdID:                                   str.Ptr(api2ORDID),
				LocalTenantID:                           str.Ptr(localTenantID),
				OrdPackageID:                            str.Ptr(packageORDID),
				Name:                                    "Gateway Sample Service",
				Description:                             str.Ptr("long desc"),
				TargetURLs:                              json.RawMessage(`["http://localhost:8080/some-api/v1"]`),
				ShortDescription:                        str.Ptr("short desc"),
				SystemInstanceAware:                     &boolPtr,
				APIProtocol:                             str.Ptr("odata-v2"),
				Tags:                                    json.RawMessage(`["ZGWSAMPLE"]`),
				Countries:                               json.RawMessage(`["BR"]`),
				Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				APIResourceLinks:                        json.RawMessage(fmt.Sprintf(resourceLinksFormat, providedBaseURL)),
				ReleaseStatus:                           str.Ptr("deprecated"),
				SunsetDate:                              str.Ptr("2020-12-08T15:47:04+0000"),
				Successors:                              json.RawMessage(fmt.Sprintf(`["%s"]`, api1ORDID)),
				ChangeLogEntries:                        json.RawMessage(changeLogEntries),
				Labels:                                  json.RawMessage(labels),
				SupportedUseCases:                       json.RawMessage(supportedUseCases),
				DocumentationLabels:                     json.RawMessage(documentLabels),
				Visibility:                              str.Ptr("public"),
				Disabled:                                nil,
				PartOfProducts:                          json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:                          json.RawMessage(`["Finance","Sales"]`),
				Industry:                                json.RawMessage(`["Automotive","Banking","Chemicals"]`),
				ImplementationStandard:                  str.Ptr(apiImplementationStandard),
				CustomImplementationStandard:            nil,
				CustomImplementationStandardDescription: nil,
				LastUpdate:                              str.Ptr("2022-01-26T15:47:04+00:00"),
				DeprecationDate:                         str.Ptr("2020-11-08T15:47:04+0000"),
				ResourceDefinitions: []*model.APIResourceDefinition{
					{
						Type:      "edmx",
						MediaType: "application/xml",
						URL:       "https://TEST:443//odata/$metadata",
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
					{
						Type:      "openapi-v3",
						MediaType: "application/json",
						URL:       fmt.Sprintf("%s/odata/1.0/catalog.svc/$value?type=json", providedBaseURL),
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				PartOfConsumptionBundles: []*model.ConsumptionBundleReference{
					{
						BundleOrdID: bundleORDID,
					},
				},
				VersionInput: &model.VersionInput{
					Value: "1.1.0",
				},
			},
		},
		EventResources: []*model.EventDefinitionInput{
			{
				OrdID:                                   str.Ptr(event1ORDID),
				LocalTenantID:                           str.Ptr(localTenantID),
				OrdPackageID:                            str.Ptr(packageORDID),
				Name:                                    "EVENT TITLE",
				Description:                             str.Ptr("long desc"),
				ShortDescription:                        str.Ptr("short desc"),
				SystemInstanceAware:                     &boolPtr,
				ChangeLogEntries:                        json.RawMessage(changeLogEntries),
				Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				EventResourceLinks:                      json.RawMessage(fmt.Sprintf(resourceLinksFormat, providedBaseURL)),
				Tags:                                    json.RawMessage(`["eventTestTag"]`),
				Countries:                               json.RawMessage(`["BG","US"]`),
				ReleaseStatus:                           str.Ptr("active"),
				SunsetDate:                              nil,
				Successors:                              nil,
				Labels:                                  json.RawMessage(labels),
				DocumentationLabels:                     json.RawMessage(documentLabels),
				Visibility:                              str.Ptr("public"),
				Disabled:                                &boolPtr,
				PartOfProducts:                          json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:                          json.RawMessage(`["Finance","Sales"]`),
				Industry:                                json.RawMessage(`["Automotive","Banking","Chemicals"]`),
				Extensible:                              json.RawMessage(`{"supported":"automatic","description":"Please find the extensibility documentation"}`),
				ImplementationStandard:                  str.Ptr(custom),
				CustomImplementationStandard:            str.Ptr("sap.foo.bar:some-event-contract:v1"),
				CustomImplementationStandardDescription: str.Ptr("description"),
				LastUpdate:                              str.Ptr("2023-01-26T15:47:04+00:00"),
				ResourceDefinitions: []*model.EventResourceDefinition{
					{
						Type:      "asyncapi-v2",
						MediaType: "application/json",
						URL:       "http://localhost:8080/asyncApi2.json",
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				PartOfConsumptionBundles: []*model.ConsumptionBundleReference{
					{
						BundleOrdID: bundleORDID,
					},
				},
				EntityTypeMappings: []*model.EntityTypeMappingInput{
					{
						APIModelSelectors: json.RawMessage(eventAPIModelSelectors),
						EntityTypeTargets: json.RawMessage(eventEntityTypeTargets),
					},
				},
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
			},
			{
				OrdID:               str.Ptr(event2ORDID),
				LocalTenantID:       str.Ptr(localTenantID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "EVENT TITLE 2",
				Description:         str.Ptr("long desc"),
				ShortDescription:    str.Ptr("short desc"),
				SystemInstanceAware: &boolPtr,
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				EventResourceLinks:  json.RawMessage(fmt.Sprintf(resourceLinksFormat, providedBaseURL)),
				Tags:                json.RawMessage(`["eventTestTag2"]`),
				Countries:           json.RawMessage(`["BR"]`),
				ReleaseStatus:       str.Ptr("deprecated"),
				SunsetDate:          str.Ptr("2020-12-08T15:47:04+0000"),
				Successors:          json.RawMessage(fmt.Sprintf(`["%s"]`, event2ORDID)),
				Labels:              json.RawMessage(labels),
				DocumentationLabels: json.RawMessage(documentLabels),
				Visibility:          str.Ptr("public"),
				Disabled:            nil,
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				LineOfBusiness:      json.RawMessage(`["Finance","Sales"]`),
				Industry:            json.RawMessage(`["Automotive","Banking","Chemicals"]`),
				Extensible:          json.RawMessage(`{"supported":"automatic","description":"Please find the extensibility documentation"}`),
				LastUpdate:          str.Ptr("2022-01-26T15:47:04+00:00"),
				DeprecationDate:     str.Ptr("2020-11-08T15:47:04+0000"),
				ResourceDefinitions: []*model.EventResourceDefinition{
					{
						Type:      "asyncapi-v2",
						MediaType: "application/json",
						URL:       fmt.Sprintf("%s/api/eventCatalog.json", providedBaseURL),
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				PartOfConsumptionBundles: []*model.ConsumptionBundleReference{
					{
						BundleOrdID: bundleORDID,
					},
				},
				VersionInput: &model.VersionInput{
					Value: "1.1.0",
				},
			},
		},
		EntityTypes: []*model.EntityTypeInput{
			{
				OrdID:               entityType1ORDID,
				LocalTenantID:       localTenantID,
				Level:               "aggregate",
				Title:               "Business Partner",
				ShortDescription:    str.Ptr("short desc"),
				Description:         str.Ptr("long desc"),
				SystemInstanceAware: &boolPtr,
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				OrdPackageID:        packageORDID,
				Visibility:          "public",
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				PolicyLevel:         str.Ptr(policyLevel),
				ReleaseStatus:       "active",
				SunsetDate:          nil,
				DeprecationDate:     nil,
				Successors:          nil,
				Extensible:          json.RawMessage(`{"supported":"automatic","description":"Please find the extensibility documentation"}`),
				Tags:                json.RawMessage(`["eventTestTag"]`),
				Labels:              json.RawMessage(labels),
				DocumentationLabels: json.RawMessage(documentLabels),
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
			},
			{
				OrdID:               entityType2ORDID,
				LocalTenantID:       localTenantID,
				Level:               "aggregate",
				Title:               "Workforce Person",
				ShortDescription:    str.Ptr("short desc"),
				Description:         str.Ptr("long desc"),
				SystemInstanceAware: &boolPtr,
				ChangeLogEntries:    json.RawMessage(changeLogEntries),
				OrdPackageID:        packageORDID,
				Visibility:          "public",
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
				PolicyLevel:         str.Ptr(policyLevel),
				ReleaseStatus:       "active",
				SunsetDate:          nil,
				DeprecationDate:     nil,
				Successors:          nil,
				Extensible:          json.RawMessage(`{"supported":"automatic","description":"Please find the extensibility documentation"}`),
				Tags:                json.RawMessage(`["eventTestTag"]`),
				Labels:              json.RawMessage(labels),
				DocumentationLabels: json.RawMessage(documentLabels),
				VersionInput: &model.VersionInput{
					Value: "1.1.0",
				},
			},
		},
		Capabilities: []*model.CapabilityInput{
			{
				OrdID:               str.Ptr(capability1ORDID),
				LocalTenantID:       str.Ptr(localTenantID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "Capability Title",
				Description:         str.Ptr("Capability Description"),
				Type:                "sap.mdo:mdi-capability:v1",
				CustomType:          nil,
				ShortDescription:    str.Ptr("Capability short description"),
				SystemInstanceAware: &boolPtr,
				Tags:                json.RawMessage(`["capabilityTestTag"]`),
				RelatedEntityTypes:  json.RawMessage(`["ns:entityType:ENTITYTYPE_ID:v1"]`),
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				ReleaseStatus:       str.Ptr("active"),
				Labels:              json.RawMessage(labels),
				Visibility:          str.Ptr("public"),
				CapabilityDefinitions: []*model.CapabilityDefinition{
					{
						Type:      "sap.mdo:mdi-capability-definition:v1",
						MediaType: "application/json",
						URL:       "http://localhost:8080/Capability.json",
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				DocumentationLabels: json.RawMessage(documentLabels),
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
				LastUpdate: str.Ptr("2023-01-26T15:47:04+00:00"),
			},
			{
				OrdID:               str.Ptr(capability2ORDID),
				LocalTenantID:       str.Ptr(localTenantID),
				OrdPackageID:        str.Ptr(packageORDID),
				Name:                "Capability Title 2",
				Description:         str.Ptr("Capability Description"),
				Type:                "sap.mdo:mdi-capability:v1",
				CustomType:          nil,
				ShortDescription:    str.Ptr("Capability short description"),
				SystemInstanceAware: &boolPtr,
				Tags:                json.RawMessage(`["capabilityTestTag"]`),
				RelatedEntityTypes:  json.RawMessage(`["ns:entityType:ENTITYTYPE_ID:v1"]`),
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				ReleaseStatus:       str.Ptr("active"),
				Labels:              json.RawMessage(labels),
				Visibility:          str.Ptr("public"),
				LastUpdate:          str.Ptr("2022-01-26T15:47:04+00:00"),
				CapabilityDefinitions: []*model.CapabilityDefinition{
					{
						Type:      "sap.mdo:mdi-capability-definition:v1",
						MediaType: "application/json",
						URL:       "http://localhost:8080/Capability.json",
						AccessStrategy: []accessstrategy.AccessStrategy{
							{
								Type: "open",
							},
						},
					},
				},
				DocumentationLabels: json.RawMessage(documentLabels),
				VersionInput: &model.VersionInput{
					Value: "1.1.1",
				},
			},
		},
		IntegrationDependencies: []*model.IntegrationDependencyInput{
			{
				OrdID:            str.Ptr(integrationDependency1ORDID),
				LocalTenantID:    str.Ptr(localTenantID),
				OrdPackageID:     str.Ptr(packageORDID),
				Title:            "Integration Dependency Title",
				Description:      str.Ptr("Integration Dependency Description"),
				ShortDescription: str.Ptr("Integration Dependency short description"),
				Tags:             json.RawMessage(`["integrationDependencyTestTag"]`),
				Links:            json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				Mandatory:        &mandatoryTrue,
				Aspects: []*model.AspectInput{
					{
						Title:                    "Aspect Title",
						Description:              str.Ptr("Description of Aspect"),
						Mandatory:                &mandatoryTrue,
						SupportMultipleProviders: &mandatoryTrue,
						APIResources:             json.RawMessage("[]"),
						EventResources:           json.RawMessage("[]"),
					},
				},
				ReleaseStatus:       str.Ptr("active"),
				Labels:              json.RawMessage(labels),
				Visibility:          "public",
				DocumentationLabels: json.RawMessage(documentLabels),
				VersionInput: &model.VersionInput{
					Value: "2.1.2",
				},
				LastUpdate: str.Ptr("2023-01-26T15:47:04+00:00"),
			},
			{
				OrdID:               str.Ptr(integrationDependency2ORDID),
				LocalTenantID:       str.Ptr(localTenantID),
				OrdPackageID:        str.Ptr(packageORDID),
				Title:               "Integration Dependency Title 2",
				Description:         str.Ptr("Integration Dependency Description"),
				ShortDescription:    str.Ptr("Integration Dependency short description"),
				Tags:                json.RawMessage(`["integrationDependencyTestTag"]`),
				Links:               json.RawMessage(fmt.Sprintf(linksFormat, providedBaseURL)),
				Mandatory:           &mandatoryTrue,
				ReleaseStatus:       str.Ptr("active"),
				Labels:              json.RawMessage(labels),
				Visibility:          "public",
				LastUpdate:          str.Ptr("2022-01-26T15:47:04+00:00"),
				DocumentationLabels: json.RawMessage(documentLabels),
				VersionInput: &model.VersionInput{
					Value: "1.1.1",
				},
			},
		},
		Tombstones: []*model.TombstoneInput{
			{
				OrdID:       api2ORDID,
				RemovalDate: "2020-12-02T14:12:59Z",
				Description: str.Ptr("long description"),
			},
		},
		Vendors: []*model.VendorInput{
			{
				OrdID:               vendorORDID,
				Title:               "SAP",
				Partners:            json.RawMessage(partners),
				Tags:                json.RawMessage(tags),
				Labels:              json.RawMessage(labels),
				DocumentationLabels: json.RawMessage(documentLabels),
			},
			{
				OrdID:               vendor2ORDID,
				Title:               "SAP",
				Partners:            json.RawMessage(partners),
				Tags:                json.RawMessage(tags),
				Labels:              json.RawMessage(labels),
				DocumentationLabels: json.RawMessage(documentLabels),
			},
		},
	}
}

func fixORDStaticDocumentWithBaseURL(providedBaseURL string) *ord.Document {
	doc := fixORDDocumentWithBaseURL(providedBaseURL)
	doc.DescribedSystemInstance = nil
	doc.DescribedSystemVersion = fixAppTemplateVersionInput()
	doc.ConsumptionBundles[0].CredentialExchangeStrategies = json.RawMessage(credentialExchangeStrategiesBasic)

	return doc
}

func fixApplicationPage() *model.ApplicationPage {
	return &model.ApplicationPage{
		Data: []*model.Application{
			{
				Name: "testApp",
				BaseEntity: &model.BaseEntity{
					ID:    appID,
					Ready: true,
				},
				Type:                  testApplicationType,
				ApplicationTemplateID: str.Ptr(appTemplateID),
			},
		},
		PageInfo: &pagination.Page{
			StartCursor: cursor,
			EndCursor:   cursor,
			HasNextPage: false,
		},
		TotalCount: 1,
	}
}

func fixAppTemplate() *model.ApplicationTemplate {
	return &model.ApplicationTemplate{
		ID:   appTemplateID,
		Name: appTemplateName,
	}
}

func fixAppTemplateVersions() []*model.ApplicationTemplateVersion {
	return []*model.ApplicationTemplateVersion{
		fixAppTemplateVersion(),
	}
}

func fixAppTemplateVersion() *model.ApplicationTemplateVersion {
	return &model.ApplicationTemplateVersion{
		ID:                    appTemplateVersionID,
		Version:               appTemplateVersionValue,
		CorrelationIDs:        json.RawMessage(correlationIDs),
		ApplicationTemplateID: appTemplateID,
	}
}

func fixAppTemplateVersionInput() *model.ApplicationTemplateVersionInput {
	return &model.ApplicationTemplateVersionInput{
		Version:        appTemplateVersionValue,
		Title:          str.Ptr("Title"),
		ReleaseDate:    str.Ptr("2020-12-08T15:47:04+0000"),
		CorrelationIDs: json.RawMessage(correlationIDs),
	}
}

func fixApplications() []*model.Application {
	return []*model.Application{
		{
			Name: "testApp",
			BaseEntity: &model.BaseEntity{
				ID:    appID,
				Ready: true,
			},
			Type:                  testApplicationType,
			ApplicationTemplateID: str.Ptr(appTemplateID),
		},
	}
}

func fixApplicationsWithBaseURL() []*model.Application {
	return []*model.Application{
		{
			Name: "testApp",
			BaseEntity: &model.BaseEntity{
				ID:    appID,
				Ready: true,
			},
			BaseURL:               str.Ptr(baseURL),
			Type:                  testApplicationType,
			ApplicationTemplateID: str.Ptr(appTemplateID),
		},
	}
}

func fixTenantMappingWebhookGraphQLInput() *graphql.WebhookInput {
	syncMode := graphql.WebhookModeSync
	return &graphql.WebhookInput{
		URL: str.Ptr("http://example.com/credentials"),
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		},
		Mode:    &syncMode,
		Version: str.Ptr(credentialExchangeStrategyVersion),
	}
}

func fixTenantMappingWebhookModelInput() *model.WebhookInput {
	syncMode := model.WebhookModeSync
	return &model.WebhookInput{
		URL: str.Ptr("http://example.com/credentials"),
		Auth: &model.AuthInput{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		},
		Mode: &syncMode,
	}
}
func fixWebhookForApplicationWithProxyURL() *model.Webhook {
	return &model.Webhook{
		ID:             whID,
		ObjectID:       appID,
		ObjectType:     model.ApplicationWebhookReference,
		Type:           model.WebhookTypeOpenResourceDiscovery,
		URL:            str.Ptr(baseURL),
		ProxyURL:       str.Ptr(customWebhookConfigURL),
		HeaderTemplate: str.Ptr(`{"target_host": ["{{.Application.BaseURL}}"] }`),
	}
}
func fixWebhooksForApplication() []*model.Webhook {
	return []*model.Webhook{
		{
			ID:         whID,
			ObjectID:   appID,
			ObjectType: model.ApplicationWebhookReference,
			Type:       model.WebhookTypeOpenResourceDiscovery,
			URL:        str.Ptr(baseURL),
		},
	}
}
func fixOrdWebhooksForAppTemplate() []*model.Webhook {
	return []*model.Webhook{
		{
			ID:         whID,
			ObjectID:   appTemplateID,
			ObjectType: model.ApplicationTemplateWebhookReference,
			Type:       model.WebhookTypeOpenResourceDiscovery,
			URL:        str.Ptr(baseURL),
		},
	}
}
func fixStaticOrdWebhooksForAppTemplate() []*model.Webhook {
	return []*model.Webhook{
		{
			ID:         whID,
			ObjectID:   appTemplateID,
			ObjectType: model.ApplicationTemplateWebhookReference,
			Type:       model.WebhookTypeOpenResourceDiscoveryStatic,
			URL:        str.Ptr(baseURL),
		},
	}
}
func fixTenantMappingWebhooksForApplication() []*model.Webhook {
	syncMode := model.WebhookModeSync
	return []*model.Webhook{{
		ID:  webhookID,
		URL: str.Ptr("http://example.com/credentials"),
		Auth: &model.Auth{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		},
		Mode:       &syncMode,
		ObjectType: model.ApplicationWebhookReference,
		ObjectID:   appID,
	}}
}

func fixVendors() []*model.Vendor {
	return []*model.Vendor{
		{
			ID:                  vendorID,
			OrdID:               vendorORDID,
			ApplicationID:       str.Ptr(appID),
			Title:               "SAP",
			Partners:            json.RawMessage(partners),
			Labels:              json.RawMessage(labels),
			DocumentationLabels: json.RawMessage(documentLabels),
		},
		{
			ID:                  vendorID2,
			OrdID:               vendor2ORDID,
			Title:               "SAP",
			Partners:            json.RawMessage(partners),
			Labels:              json.RawMessage(labels),
			DocumentationLabels: json.RawMessage(documentLabels),
		},
	}
}

func fixGlobalVendors() []*model.Vendor {
	return []*model.Vendor{
		{
			ID:    vendorID,
			OrdID: vendorORDID,
			Title: "SAP SE",
		},
	}
}

func fixProducts() []*model.Product {
	return []*model.Product{
		{
			ID:                  productID,
			OrdID:               productORDID,
			ApplicationID:       str.Ptr(appID),
			Title:               "PRODUCT TITLE",
			Description:         str.Ptr("long description"),
			ShortDescription:    "short description",
			Vendor:              vendorORDID,
			Parent:              str.Ptr(product2ORDID),
			CorrelationIDs:      json.RawMessage(`["foo.bar.baz:123456"]`),
			Labels:              json.RawMessage(labels),
			DocumentationLabels: json.RawMessage(documentLabels),
		},
	}
}

func fixGlobalProducts() []*model.Product {
	return []*model.Product{
		{
			ID:               productID,
			OrdID:            globalProductORDID,
			Title:            "SAP Business Technology Platform",
			ShortDescription: "Enhance business results",
			Description:      str.Ptr("Accelerate business outcomes with integration, data to value, and extensibility."),
			Vendor:           vendorORDID,
		},
	}
}

func fixPackages() []*model.Package {
	return []*model.Package{
		{
			ID:                  packageID,
			ApplicationID:       &appID,
			OrdID:               packageORDID,
			Vendor:              str.Ptr(vendorORDID),
			Title:               "PACKAGE 1 TITLE",
			ShortDescription:    "lorem ipsum",
			Description:         "lorem ipsum dolor set",
			Version:             "1.1.2",
			PackageLinks:        json.RawMessage(fmt.Sprintf(packageLinksFormat, baseURL)),
			Links:               json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			LicenseType:         str.Ptr("licence"),
			SupportInfo:         str.Ptr("support-info"),
			Tags:                json.RawMessage(`["testTag"]`),
			Countries:           json.RawMessage(`["BG","EN"]`),
			Labels:              json.RawMessage(packageLabels),
			DocumentationLabels: json.RawMessage(documentLabels),
			PolicyLevel:         str.Ptr(policyLevel),
			PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
			LineOfBusiness:      json.RawMessage(`["Finance","Sales"]`),
			Industry:            json.RawMessage(`["Automotive","Banking","Chemicals"]`),
		},
	}
}

func fixBundles() []*model.Bundle {
	return []*model.Bundle{
		{
			ApplicationID:                &appID,
			Name:                         "BUNDLE TITLE",
			Description:                  str.Ptr("lorem ipsum dolor nsq sme"),
			Version:                      str.Ptr("1.1.2"),
			OrdID:                        str.Ptr(bundleORDID),
			LocalTenantID:                str.Ptr(localTenantID),
			ShortDescription:             str.Ptr("lorem ipsum"),
			Links:                        json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			Labels:                       json.RawMessage(labels),
			DocumentationLabels:          json.RawMessage(documentLabels),
			CredentialExchangeStrategies: json.RawMessage(fmt.Sprintf(credentialExchangeStrategiesFormat, baseURL)),
			CorrelationIDs:               json.RawMessage(correlationIDs),
			BaseEntity: &model.BaseEntity{
				ID:    bundleID,
				Ready: true,
			},
		},
	}
}

func fixBundlesWithCredentialExchangeStrategies() []*model.Bundle {
	bundles := fixBundles()
	bundles[0].CredentialExchangeStrategies = json.RawMessage(credentialExchangeStrategiesBasic)
	return bundles
}

func fixBundleCreateInput() []*model.BundleCreateInput {
	return []*model.BundleCreateInput{
		{
			Name:                "BUNDLE TITLE",
			Description:         str.Ptr("lorem ipsum dolor nsq sme"),
			Version:             str.Ptr("1.1.2"),
			OrdID:               str.Ptr(bundleORDID),
			LocalTenantID:       str.Ptr(localTenantID),
			ShortDescription:    str.Ptr("lorem ipsum"),
			Labels:              json.RawMessage(labels),
			DocumentationLabels: json.RawMessage(documentLabels),
			CorrelationIDs:      json.RawMessage(correlationIDs),
		},
		{
			Name:                "BUNDLE TITLE 2 ",
			Description:         str.Ptr("foo bar"),
			Version:             str.Ptr("1.1.2"),
			OrdID:               str.Ptr(secondBundleORDID),
			LocalTenantID:       str.Ptr(localTenantID),
			ShortDescription:    str.Ptr("bar foo"),
			Labels:              json.RawMessage(labels),
			DocumentationLabels: json.RawMessage(documentLabels),
			CorrelationIDs:      json.RawMessage(correlationIDs),
		},
	}
}

func fixAPIsWithHash() []*model.APIDefinition {
	apis := fixAPIs()

	for idx, api := range apis {
		ordID := str.PtrStrToStr(api.OrdID)
		hash := str.Ptr(strconv.FormatUint(resourceHashes[ordID], 10))
		api.ResourceHash = hash
		api.Version.Value = fixORDDocument().APIResources[idx].VersionInput.Value
	}

	return apis
}

func fixEventsWithHash() []*model.EventDefinition {
	events := fixEvents()

	for idx, event := range events {
		ordID := str.PtrStrToStr(event.OrdID)
		hash := str.Ptr(strconv.FormatUint(resourceHashes[ordID], 10))
		event.ResourceHash = hash
		event.Version.Value = fixORDDocument().EventResources[idx].VersionInput.Value
	}

	return events
}

func fixCapabilitiesWithHash() []*model.Capability {
	capabilities := fixCapabilities()

	for idx, capability := range capabilities {
		ordID := str.PtrStrToStr(capability.OrdID)
		hash := str.Ptr(strconv.FormatUint(resourceHashes[ordID], 10))
		capability.ResourceHash = hash
		capability.Version.Value = fixORDDocument().Capabilities[idx].VersionInput.Value
	}

	return capabilities
}

func fixIntegrationDependenciesWithHash() []*model.IntegrationDependency {
	integrationDependencies := fixIntegrationDependencies()

	for idx, integrationDependency := range integrationDependencies {
		ordID := str.PtrStrToStr(integrationDependency.OrdID)
		hash := str.Ptr(strconv.FormatUint(resourceHashes[ordID], 10))
		integrationDependency.ResourceHash = hash
		integrationDependency.Version.Value = fixORDDocument().IntegrationDependencies[idx].VersionInput.Value
	}
	return integrationDependencies
}

func fixPackagesWithHash() []*model.Package {
	pkgs := fixPackages()

	for idx, pkg := range pkgs {
		hash := str.Ptr(strconv.FormatUint(resourceHashes[pkg.OrdID], 10))
		pkg.ResourceHash = hash
		pkg.Version = fixORDDocument().Packages[idx].Version
	}

	return pkgs
}

func fixBundlesWithHash() []*model.Bundle {
	bndls := fixBundles()

	for idx, bndl := range bndls {
		hash := str.Ptr(strconv.FormatUint(resourceHashes[str.PtrStrToStr(bndl.OrdID)], 10))
		bndl.ResourceHash = hash
		bndl.Version = fixORDDocument().ConsumptionBundles[idx].Version
	}

	return bndls
}

func fixAPIs() []*model.APIDefinition {
	return []*model.APIDefinition{
		{
			ApplicationID:                           &appID,
			PackageID:                               str.Ptr(packageORDID),
			Name:                                    "API TITLE",
			Description:                             str.Ptr("lorem ipsum dolor sit amet"),
			TargetURLs:                              json.RawMessage(`["/test/v1"]`),
			OrdID:                                   str.Ptr(api1ORDID),
			ShortDescription:                        str.Ptr("lorem ipsum"),
			APIProtocol:                             str.Ptr("odata-v2"),
			Tags:                                    json.RawMessage(`["testTag","apiTestTag"]`),
			Countries:                               json.RawMessage(`["BG","EN","US"]`),
			Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			APIResourceLinks:                        json.RawMessage(fmt.Sprintf(resourceLinksFormat, baseURL)),
			ReleaseStatus:                           str.Ptr("active"),
			ChangeLogEntries:                        json.RawMessage(changeLogEntries),
			Labels:                                  json.RawMessage(mergedLabels),
			Visibility:                              str.Ptr("public"),
			Disabled:                                &boolPtr,
			PartOfProducts:                          json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
			LineOfBusiness:                          json.RawMessage(`["Finance","Sales"]`),
			Industry:                                json.RawMessage(`["Automotive","Banking","Chemicals"]`),
			ImplementationStandard:                  str.Ptr(apiImplementationStandard),
			CustomImplementationStandard:            nil,
			CustomImplementationStandardDescription: nil,
			LastUpdate:                              str.Ptr("2023-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "2.1.3",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    api1ID,
				Ready: true,
			},
		},
		{
			ApplicationID:                           &appID,
			PackageID:                               str.Ptr(packageORDID),
			Name:                                    "Gateway Sample Service",
			Description:                             str.Ptr("lorem ipsum dolor sit amet"),
			TargetURLs:                              json.RawMessage(`["/some-api/v1"]`),
			OrdID:                                   str.Ptr(api2ORDID),
			ShortDescription:                        str.Ptr("lorem ipsum"),
			APIProtocol:                             str.Ptr("odata-v2"),
			Tags:                                    json.RawMessage(`["testTag","ZGWSAMPLE"]`),
			Countries:                               json.RawMessage(`["BG","EN","BR"]`),
			Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			APIResourceLinks:                        json.RawMessage(fmt.Sprintf(resourceLinksFormat, baseURL)),
			ReleaseStatus:                           str.Ptr("deprecated"),
			SunsetDate:                              str.Ptr("2020-12-08T15:47:04+0000"),
			Successors:                              json.RawMessage(fmt.Sprintf(`["%s"]`, api1ORDID)),
			ChangeLogEntries:                        json.RawMessage(changeLogEntries),
			Labels:                                  json.RawMessage(mergedLabels),
			Visibility:                              str.Ptr("public"),
			PartOfProducts:                          json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
			LineOfBusiness:                          json.RawMessage(`["Finance","Sales"]`),
			Industry:                                json.RawMessage(`["Automotive","Banking","Chemicals"]`),
			ImplementationStandard:                  str.Ptr(apiImplementationStandard),
			CustomImplementationStandard:            nil,
			CustomImplementationStandardDescription: nil,
			LastUpdate:                              str.Ptr("2022-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "1.1.1",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    api2ID,
				Ready: true,
			},
		},
	}
}

func fixAPIsNoNewerLastUpdate() []*model.APIDefinition {
	apis := fixAPIs()
	doc := fixORDDocument()
	for i, api := range apis {
		api.LastUpdate = doc.APIResources[i].LastUpdate
	}
	return apis
}

func fixAPIPartOfConsumptionBundles() []*model.ConsumptionBundleReference {
	return []*model.ConsumptionBundleReference{
		{
			BundleOrdID:      bundleORDID,
			DefaultTargetURL: "https://exmaple.com/test/v1",
		},
		{
			BundleOrdID:      secondBundleORDID,
			DefaultTargetURL: "https://exmaple.com/test/v2",
		},
	}
}

func fixEventPartOfConsumptionBundles() []*model.ConsumptionBundleReference {
	return []*model.ConsumptionBundleReference{
		{
			BundleOrdID: bundleORDID,
		},
		{
			BundleOrdID: secondBundleORDID,
		},
	}
}

func fixEvents() []*model.EventDefinition {
	return []*model.EventDefinition{
		{
			ApplicationID:       &appID,
			PackageID:           str.Ptr(packageORDID),
			Name:                "EVENT TITLE",
			Description:         str.Ptr("lorem ipsum dolor sit amet"),
			OrdID:               str.Ptr(event1ORDID),
			ShortDescription:    str.Ptr("lorem ipsum"),
			ChangeLogEntries:    json.RawMessage(changeLogEntries),
			Links:               json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			Tags:                json.RawMessage(`["testTag","eventTestTag"]`),
			Countries:           json.RawMessage(`["BG","EN","US"]`),
			ReleaseStatus:       str.Ptr("active"),
			Labels:              json.RawMessage(mergedLabels),
			DocumentationLabels: json.RawMessage(documentLabels),
			Visibility:          str.Ptr("public"),
			Disabled:            &boolPtr,
			PartOfProducts:      json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
			LineOfBusiness:      json.RawMessage(`["Finance","Sales"]`),
			Industry:            json.RawMessage(`["Automotive","Banking","Chemicals"]`),
			LastUpdate:          str.Ptr("2023-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "2.1.3",
			},
			BaseEntity: &model.BaseEntity{
				ID:    event1ID,
				Ready: true,
			},
		},
		{
			ApplicationID:    &appID,
			PackageID:        str.Ptr(packageORDID),
			Name:             "EVENT TITLE 2",
			Description:      str.Ptr("lorem ipsum dolor sit amet"),
			OrdID:            str.Ptr(event2ORDID),
			ShortDescription: str.Ptr("lorem ipsum"),
			ChangeLogEntries: json.RawMessage(changeLogEntries),
			Links:            json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			Tags:             json.RawMessage(`["testTag","eventTestTag2"]`),
			Countries:        json.RawMessage(`["BG","EN","BR"]`),
			ReleaseStatus:    str.Ptr("deprecated"),
			SunsetDate:       str.Ptr("2020-12-08T15:47:04+0000"),
			Successors:       json.RawMessage(fmt.Sprintf(`["%s"]`, event2ORDID)),
			Labels:           json.RawMessage(mergedLabels),
			Visibility:       str.Ptr("public"),
			PartOfProducts:   json.RawMessage(fmt.Sprintf(`["%s"]`, productORDID)),
			LineOfBusiness:   json.RawMessage(`["Finance","Sales"]`),
			Industry:         json.RawMessage(`["Automotive","Banking","Chemicals"]`),
			LastUpdate:       str.Ptr("2022-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "1.1.1",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    event2ID,
				Ready: true,
			},
		},
	}
}

func fixEventsNoNewerLastUpdate() []*model.EventDefinition {
	events := fixEvents()
	doc := fixORDDocument()
	for i, event := range events {
		event.LastUpdate = doc.EventResources[i].LastUpdate
	}
	return events
}

func fixCapabilities() []*model.Capability {
	return []*model.Capability{
		{
			ApplicationID:       &appID,
			PackageID:           str.Ptr(packageORDID),
			Name:                "Capability Title",
			Description:         str.Ptr("Capability Description"),
			OrdID:               str.Ptr(capability1ORDID),
			Type:                "sap.mdo:mdi-capability:v1",
			CustomType:          nil,
			LocalTenantID:       nil,
			ShortDescription:    str.Ptr("Capability short description"),
			SystemInstanceAware: nil,
			Tags:                json.RawMessage(`["testTag","capabilityTestTag"]`),
			RelatedEntityTypes:  json.RawMessage(`["ns:entityType:ENTITYTYPE_ID:v1"]`),
			Links:               json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			ReleaseStatus:       str.Ptr("active"),
			Labels:              json.RawMessage(mergedLabels),
			Visibility:          str.Ptr("public"),
			LastUpdate:          str.Ptr("2023-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "2.1.3",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    capability1ID,
				Ready: true,
			},
		},
		{
			ApplicationID:       &appID,
			PackageID:           str.Ptr(packageORDID),
			Name:                "Capability Title 2",
			Description:         str.Ptr("Capability Description"),
			OrdID:               str.Ptr(capability2ORDID),
			Type:                "sap.mdo:mdi-capability:v1",
			CustomType:          nil,
			LocalTenantID:       nil,
			ShortDescription:    str.Ptr("Capability short description"),
			SystemInstanceAware: nil,
			Tags:                json.RawMessage(`["testTag","capabilityTestTag"]`),
			RelatedEntityTypes:  json.RawMessage(`["ns:entityType:ENTITYTYPE_ID:v1"]`),
			Links:               json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			ReleaseStatus:       str.Ptr("active"),
			Labels:              json.RawMessage(mergedLabels),
			Visibility:          str.Ptr("public"),
			LastUpdate:          str.Ptr("2022-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "1.1.0",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    capability2ID,
				Ready: true,
			},
		},
	}
}

func fixCapabilitiesNoNewerLastUpdate() []*model.Capability {
	capabilities := fixCapabilities()
	doc := fixORDDocument()
	for i, capability := range capabilities {
		capability.LastUpdate = doc.Capabilities[i].LastUpdate
	}
	return capabilities
}

func fixAPI1SpecInputs(url string) []*model.SpecInput {
	openAPIType := model.APISpecTypeOpenAPIV3
	edmxAPIType := model.APISpecTypeEDMX
	return []*model.SpecInput{
		{
			Format:     "application/json",
			APIType:    &openAPIType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  url + "/external-api/unsecured/spec/flapping",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
		{
			Format:     "text/yaml",
			APIType:    &openAPIType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  "https://test.com/odata/1.0/catalog",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
		{
			Format:     "application/xml",
			APIType:    &edmxAPIType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  "https://TEST:443//odata/$metadata",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
	}
}

func fixAPI1IDs() []string {
	return []string{api1spec1ID, api1spec2ID, api1spec3ID}
}

func fixAPI2SpecInputs(url string) []*model.SpecInput {
	edmxAPIType := model.APISpecTypeEDMX
	openAPIType := model.APISpecTypeOpenAPIV3
	return []*model.SpecInput{
		{
			Format:     "application/xml",
			APIType:    &edmxAPIType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  "https://TEST:443//odata/$metadata",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
		{
			Format:     "application/json",
			APIType:    &openAPIType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  url + "/odata/1.0/catalog.svc/$value?type=json",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
	}
}

func fixAPI2IDs() []string {
	return []string{api2spec1ID, api2spec2ID}
}

func fixEvent1SpecInputs() []*model.SpecInput {
	eventType := model.EventSpecTypeAsyncAPIV2
	return []*model.SpecInput{
		{
			Format:     "application/json",
			EventType:  &eventType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  "http://localhost:8080/asyncApi2.json",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
	}
}

func fixEvent1IDs() []string {
	return []string{event1specID}
}

func fixEvent2SpecInputs(url string) []*model.SpecInput {
	eventType := model.EventSpecTypeAsyncAPIV2
	return []*model.SpecInput{
		{
			Format:     "application/json",
			EventType:  &eventType,
			CustomType: str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  url + "/api/eventCatalog.json",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
	}
}

func fixEvent2IDs() []string {
	return []string{event2specID}
}

func fixCapability1IDs() []string {
	return []string{capability1SpecID}
}

func fixCapability2IDs() []string {
	return []string{capability2SpecID}
}

func fixCapabilitySpecInputs() []*model.SpecInput {
	capabilityType := model.CapabilitySpecTypeMDICapabilityDefinitionV1
	return []*model.SpecInput{
		{
			Format:         "application/json",
			CapabilityType: &capabilityType,
			CustomType:     str.Ptr(""),
			FetchRequest: &model.FetchRequestInput{
				URL:  "http://localhost:8080/Capability.json",
				Auth: &model.AuthInput{AccessStrategy: str.Ptr("open")},
			},
		},
	}
}

func fixTombstones() []*model.Tombstone {
	return []*model.Tombstone{
		{
			ID:            tombstoneID,
			OrdID:         api2ORDID,
			ApplicationID: &appID,
			RemovalDate:   "2020-12-02T14:12:59Z",
			Description:   str.Ptr("description"),
		},
	}
}

func fixIntegrationDependencies() []*model.IntegrationDependency {
	return []*model.IntegrationDependency{
		{
			ApplicationID:    &appID,
			PackageID:        str.Ptr(packageORDID),
			Title:            "Integration Dependency Title",
			Description:      str.Ptr("Integration Dependency Description"),
			OrdID:            str.Ptr(integrationDependency1ORDID),
			LocalTenantID:    nil,
			ShortDescription: str.Ptr("Integration Dependency short description"),
			Tags:             json.RawMessage(`["testTag","integrationDependencyTestTag"]`),
			Links:            json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			Mandatory:        &mandatoryTrue,
			ReleaseStatus:    str.Ptr("active"),
			Labels:           json.RawMessage(mergedLabels),
			Visibility:       "public",
			LastUpdate:       str.Ptr("2023-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "2.1.3",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    integrationDependency1ID,
				Ready: true,
			},
		},
		{
			ApplicationID:    &appID,
			PackageID:        str.Ptr(packageORDID),
			Title:            "Integration Dependency Title 2",
			Description:      str.Ptr("Integration Dependency Description"),
			OrdID:            str.Ptr(integrationDependency2ORDID),
			LocalTenantID:    nil,
			ShortDescription: str.Ptr("Integration Dependency short description"),
			Tags:             json.RawMessage(`["testTag","integrationDependencyTestTag"]`),
			Links:            json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			Mandatory:        &mandatoryTrue,
			ReleaseStatus:    str.Ptr("active"),
			Labels:           json.RawMessage(mergedLabels),
			Visibility:       "public",
			LastUpdate:       str.Ptr("2022-01-25T15:47:04+00:00"),
			Version: &model.Version{
				Value: "1.1.0",
			},
			DocumentationLabels: json.RawMessage(documentLabels),
			BaseEntity: &model.BaseEntity{
				ID:    integrationDependency2ID,
				Ready: true,
			},
		},
	}
}

func fixEntityTypes() []*model.EntityType {
	return []*model.EntityType{
		{
			BaseEntity: &model.BaseEntity{
				ID:        entityTypeID,
				Ready:     true,
				CreatedAt: &fixedTimestamp,
				UpdatedAt: &time.Time{},
				DeletedAt: &time.Time{},
				Error:     nil,
			},
			ApplicationID:                &appID,
			ApplicationTemplateVersionID: &appTemplateVersionID,
			OrdID:                        entityType1ORDID,
			LocalTenantID:                localTenantID,
			CorrelationIDs:               json.RawMessage(correlationIDs),
			Level:                        level,
			Title:                        title,
			ShortDescription:             &shortDescription,
			Description:                  &description,
			SystemInstanceAware:          &systemInstanceAware,
			ChangeLogEntries:             json.RawMessage(changeLogEntries),
			PackageID:                    packageID,
			Visibility:                   publicVisibility,
			Links:                        json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			PartOfProducts:               json.RawMessage(products),
			PolicyLevel:                  &policyLevel,
			CustomPolicyLevel:            &customPolicyLevel,
			ReleaseStatus:                releaseStatus,
			SunsetDate:                   &sunsetDate,
			Successors:                   json.RawMessage(successors),
			Extensible:                   json.RawMessage(extensible),
			Tags:                         json.RawMessage(tags),
			Labels:                       json.RawMessage(labels),
			DocumentationLabels:          json.RawMessage(documentationLabels),
			Version:                      fixVersionModel(versionValue, versionDeprecated, versionDeprecatedSince, versionForRemoval),
			ResourceHash:                 &resourceHash,
		},
	}
}

func fixEntityTypeMappingInput(apiID, eventID string) *model.EntityTypeMappingInput {
	entityTypeMapping := &model.EntityTypeMappingInput{}
	if apiID != "" {
		entityTypeMapping.APIModelSelectors = json.RawMessage(apiAPIModelSelectors)
		entityTypeMapping.EntityTypeTargets = json.RawMessage(apiEntityTypeTargets)
	} else if eventID != "" {
		entityTypeMapping.APIModelSelectors = json.RawMessage(eventAPIModelSelectors)
		entityTypeMapping.EntityTypeTargets = json.RawMessage(eventEntityTypeTargets)
	}
	return entityTypeMapping
}

func fixEntityTypeMappingInputsEmpty() []*model.EntityTypeMappingInput {
	return []*model.EntityTypeMappingInput{}
}

func fixEntityTypeMapping(apiID, eventID string) *model.EntityTypeMapping {
	entityTypeMapping := &model.EntityTypeMapping{
		BaseEntity: &model.BaseEntity{
			ID:        entityTypeMappingID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
	}
	if apiID != "" {
		entityTypeMapping.APIDefinitionID = &apiID
		entityTypeMapping.APIModelSelectors = json.RawMessage(apiAPIModelSelectors)
		entityTypeMapping.EntityTypeTargets = json.RawMessage(apiEntityTypeTargets)
	} else if eventID != "" {
		entityTypeMapping.EventDefinitionID = &eventID
		entityTypeMapping.APIModelSelectors = json.RawMessage(eventAPIModelSelectors)
		entityTypeMapping.EntityTypeTargets = json.RawMessage(eventEntityTypeTargets)
	}
	return entityTypeMapping
}

func fixEntityTypeMappings(apiID, eventID string) []*model.EntityTypeMapping {
	return []*model.EntityTypeMapping{fixEntityTypeMapping(apiID, eventID)}
}

func fixEntityTypeMappingsEmpty() []*model.EntityTypeMapping {
	return []*model.EntityTypeMapping{}
}

func fixVersionModel(value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.Version {
	return &model.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixSuccessfulFetchRequest() *model.FetchRequest {
	return &model.FetchRequest{
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionSucceeded,
		},
	}
}

func fixFailedFetchRequest() *model.FetchRequest {
	return &model.FetchRequest{
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionFailed,
		},
	}
}

func fixFetchRequestFromFetchRequestInput(fr *model.FetchRequestInput, objectType model.FetchRequestReferenceObjectType, specID string) *model.FetchRequest {
	id := uidSvc.Generate()
	return fr.ToFetchRequest(time.Now(), id, objectType, specID)
}

func bundleUpdateInputFromCreateInput(in model.BundleCreateInput) model.BundleUpdateInput {
	return model.BundleUpdateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: in.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            in.DefaultInstanceAuth,
		OrdID:                          in.OrdID,
		ShortDescription:               in.ShortDescription,
		Links:                          in.Links,
		Labels:                         in.Labels,
		DocumentationLabels:            in.DocumentationLabels,
		CredentialExchangeStrategies:   in.CredentialExchangeStrategies,
		CorrelationIDs:                 in.CorrelationIDs,
	}
}

func fixGlobalRegistryORDDocument() *ord.Document {
	return &ord.Document{
		Schema:                "./spec/v1/generated/Document.schema.json",
		OpenResourceDiscovery: "1.0",
		Products: []*model.ProductInput{
			{
				OrdID:            globalProductORDID,
				Title:            "SAP Business Technology Platform",
				ShortDescription: "Enhance business results.",
				Description:      str.Ptr("Accelerate business outcomes with integration, data to value, and extensibility."),
				Vendor:           vendorORDID,
			},
		},
		Vendors: []*model.VendorInput{
			{
				OrdID: vendorORDID,
				Title: "SAP SE",
			},
		},
	}
}

func fixApplicationTypeLabel() *model.Label {
	return &model.Label{
		Key:   application.ApplicationTypeLabelKey,
		Value: applicationTypeLabelValue,
	}
}

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\n", ""), "\t", "")
}
