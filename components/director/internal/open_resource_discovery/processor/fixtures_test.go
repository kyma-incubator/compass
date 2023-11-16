package processor_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	ordID = "com.compass.v1"

	apiID        = "api-id"
	eventID      = "event-id"
	capabilityID = "capability-id"
	vendorID     = "vendor-id"
	entityTypeID = "entity-type-id"
	packageID    = "package-id"

	tenantID         = "testTenant"
	externalTenantID = "externalTestTenant"

	vendorORDID      = "sap:vendor:SAP:"
	packageORDID     = "ns:package:PACKAGE_ID:v1"
	productORDID     = "sap:product:id:"
	bundleORDID      = "ns:consumptionBundle:BUNDLE_ID:v1"
	apiORDID         = "ns:apiResource:API_ID:v1"
	apiORDID2        = "ns:apiResource:API_ID:v2"
	eventORDID       = "ns:eventResource:EVENT_ID:v1"
	eventORDID2      = "ns:eventResource:EVENT_ID:v2"
	capabilityORDID  = "sap.foo.bar:capability:fieldExtensibility:v1"
	capabilityORDID2 = "sap.foo.bar:capability:fieldExtensibility:v2"

	publicVisibility = "public"
	products         = `["sap:product:S4HANA_OD:"]`
	releaseStatus    = "active"
	custom           = "custom"
	localTenantID    = "BusinessPartner"
	correlationIDs   = `["sap.s4:sot:BusinessPartner", "sap.s4:sot:CostCenter", "sap.s4:sot:WorkforcePerson"]`
	level            = "aggregate"
	title            = "BusinessPartner"
	baseURL          = "http://test.com:8080"
)

var (
	fixedTimestamp         = time.Now()
	appID                  = "appID"
	appTemplateVersionID   = "appTemplateVersionID"
	shortDescription       = "A business partner is a person, an organization, or a group of persons or organizations in which a company has a business interest."
	description            = "A workforce person is a natural person with a work agreement or relationship in form of a work assignment; it can be an employee or a contingent worker.\n"
	systemInstanceAware    = false
	policyLevel            = "custom"
	customPolicyLevel      = "sap:core:v1"
	sunsetDate             = "2022-01-08T15:47:04+00:00"
	successors             = `["sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1"]`
	extensible             = `{"supported":"automatic","description":"Please find the extensibility documentation"}`
	tags                   = `["storage","high-availability"]`
	resourceHash           = "123456"
	uint64ResourceHash     = uint64(123456)
	versionValue           = "v1.1"
	versionDeprecated      = false
	versionDeprecatedSince = "v1.0"
	versionForRemoval      = false
	mandatoryTrue          = true
	boolPtr                = true
	nilString              *string
	nilSpecInput           *model.SpecInput
	nilSpecInputSlice      []*model.SpecInput
	emptyHash              uint64

	changeLogEntries = removeWhitespace(`[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`)
	links = removeWhitespace(`[
        {
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 1",
          "url": "https://example.com/2018/04/11/testing/"
        },
		{
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 2",
          "url": "http://example.com/testing/"
        }
      ]`)
	labels = removeWhitespace(`{
        "label-key-1": [
          "label-value-1",
          "label-value-2"
        ]
      }`)
	packageLabels = removeWhitespace(`{
        "label-key-1": [
          "label-val"
        ],
		"pkg-label": [
          "label-val"
        ]
      }`)
	packageLinksFormat = removeWhitespace(`[
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
	documentationLabels = removeWhitespace(`{
        "Some Aspect": ["Markdown Documentation [with links](#)", "With multiple values"]
      }`)
	apiModelSelectors = removeWhitespace(`[
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

	entityTypeTargets = removeWhitespace(`[
		{
		  "ordId": "sap.odm:entityType:WorkforcePerson:v1"
		},
		{
		  "correlationId": "sap.s4:csnEntity:WorkForcePersonView_v1"
		}
	  ]`)

	supportedUseCases = removeWhitespace(`[
        "mass-extraction"
      ]`)

	documentLabels = removeWhitespace(`{
        "Some Aspect": ["Markdown Documentation [with links](#)", "With multiple values"]
      }`)

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

	apiImplementationStandard = "cff:open-service-broker:v2"

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

	testErr = errors.New("test error")
)

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\n", ""), "\t", "")
}

func fixVersionModel(value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.Version {
	return &model.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixEntityTypeModel(entityTypeID string) *model.EntityType {
	return &model.EntityType{
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
		OrdID:                        ordID,
		LocalTenantID:                localTenantID,
		CorrelationIDs:               json.RawMessage(correlationIDs),
		Level:                        level,
		Title:                        title,
		ShortDescription:             &shortDescription,
		Description:                  &description,
		SystemInstanceAware:          &systemInstanceAware,
		ChangeLogEntries:             json.RawMessage(changeLogEntries),
		PackageID:                    packageORDID,
		Visibility:                   publicVisibility,
		Links:                        json.RawMessage(links),
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
	}
}

func fixEntityTypeInputModel() *model.EntityTypeInput {
	return &model.EntityTypeInput{
		OrdID:               ordID,
		LocalTenantID:       localTenantID,
		CorrelationIDs:      json.RawMessage(correlationIDs),
		Level:               level,
		Title:               title,
		ShortDescription:    &shortDescription,
		Description:         &description,
		SystemInstanceAware: &systemInstanceAware,
		ChangeLogEntries:    json.RawMessage(changeLogEntries),
		OrdPackageID:        packageORDID,
		Visibility:          publicVisibility,
		Links:               json.RawMessage(links),
		PartOfProducts:      json.RawMessage(products),
		PolicyLevel:         &policyLevel,
		CustomPolicyLevel:   &customPolicyLevel,
		ReleaseStatus:       releaseStatus,
		SunsetDate:          &sunsetDate,
		Successors:          json.RawMessage(successors),
		Extensible:          json.RawMessage(extensible),
		Tags:                json.RawMessage(tags),
		Labels:              json.RawMessage(labels),
		DocumentationLabels: json.RawMessage(documentationLabels),
	}
}

func fixIntegrationDependencyModel(integrationDependencyID, integrationDependencyORDID string) *model.IntegrationDependency {
	return &model.IntegrationDependency{
		BaseEntity: &model.BaseEntity{
			ID:    integrationDependencyID,
			Ready: true,
		},
		OrdID:                        str.Ptr(integrationDependencyORDID),
		ApplicationID:                &appID,
		ApplicationTemplateVersionID: &appTemplateVersionID,
		PackageID:                    str.Ptr(packageORDID),
	}
}

func fixIntegrationDependencyInputModel(integrationDependencyORDID string) *model.IntegrationDependencyInput {
	return &model.IntegrationDependencyInput{
		OrdID:        str.Ptr(integrationDependencyORDID),
		OrdPackageID: str.Ptr(packageORDID),
		Aspects: []*model.AspectInput{
			{
				Title:     "Test integration aspect name",
				Mandatory: &mandatoryTrue,
			},
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
			DocumentationLabels: json.RawMessage(documentationLabels),
			PolicyLevel:         str.Ptr(policyLevel),
			PartOfProducts:      json.RawMessage(products),
			LineOfBusiness:      json.RawMessage(`["Finance","Sales"]`),
			Industry:            json.RawMessage(`["Automotive","Banking","Chemicals"]`),
		},
	}
}

func fixEntityTypeMappingModel(entityTypeMappingID string) *model.EntityTypeMapping {
	return &model.EntityTypeMapping{
		BaseEntity: &model.BaseEntity{
			ID:        entityTypeMappingID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		APIModelSelectors: json.RawMessage(apiModelSelectors),
		EntityTypeTargets: json.RawMessage(entityTypeTargets),
	}
}

func fixEntityTypeMappingInputModel() *model.EntityTypeMappingInput {
	return &model.EntityTypeMappingInput{
		APIModelSelectors: json.RawMessage(apiModelSelectors),
		EntityTypeTargets: json.RawMessage(entityTypeTargets),
	}
}

func fixAPI(id string, ordID *string) *model.APIDefinition {
	return &model.APIDefinition{
		ApplicationID:                           &appID,
		PackageID:                               str.Ptr(packageORDID),
		Name:                                    "API TITLE",
		Description:                             str.Ptr("lorem ipsum dolor sit amet"),
		TargetURLs:                              json.RawMessage(`["/test/v1"]`),
		OrdID:                                   ordID,
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
			ID:    id,
			Ready: true,
		},
	}
}

func fixAPIInput() *model.APIDefinitionInput {
	return &model.APIDefinitionInput{
		OrdID:                                   str.Ptr(apiORDID),
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
		Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
		APIResourceLinks:                        json.RawMessage(fmt.Sprintf(resourceLinksFormat, baseURL)),
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
				URL:       fmt.Sprintf("%s/external-api/unsecured/spec/flapping", baseURL),
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
	}
}

func fixEvent(id string, ordID *string) *model.EventDefinition {
	return &model.EventDefinition{

		ApplicationID:       &appID,
		PackageID:           str.Ptr(packageORDID),
		Name:                "EVENT TITLE",
		Description:         str.Ptr("lorem ipsum dolor sit amet"),
		OrdID:               ordID,
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
			ID:    id,
			Ready: true,
		},
	}
}

func fixCapability(id string, ordID *string) *model.Capability {
	return &model.Capability{
		ApplicationID:       &appID,
		PackageID:           str.Ptr(packageORDID),
		Name:                "Capability Title",
		Description:         str.Ptr("Capability Description"),
		OrdID:               ordID,
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
			ID:    id,
			Ready: true,
		},
	}
}

func fixCapabilityInput() *model.CapabilityInput {
	return &model.CapabilityInput{
		OrdID:               str.Ptr(capabilityORDID),
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
		Links:               json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
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
	}
}

func fixAPIsNoNewerLastUpdate() []*model.APIDefinition {
	api := fixAPI(apiID, str.Ptr(apiORDID))
	api.LastUpdate = fixAPIInput().LastUpdate
	return []*model.APIDefinition{
		api,
	}
}

func fixEventsNoNewerLastUpdate() []*model.EventDefinition {
	event := fixEvent(eventID, str.Ptr(eventORDID))
	event.LastUpdate = fixEventInput().LastUpdate
	return []*model.EventDefinition{
		event,
	}
}

func fixCapabilitiesNoNewerLastUpdate() []*model.Capability {
	capability := fixCapability(capabilityID, str.Ptr(capabilityORDID))
	capability.LastUpdate = fixCapabilityInput().LastUpdate
	return []*model.Capability{
		capability,
	}
}

func fixEventInput() *model.EventDefinitionInput {
	return &model.EventDefinitionInput{
		OrdID:                                   str.Ptr(eventORDID),
		LocalTenantID:                           str.Ptr(localTenantID),
		OrdPackageID:                            str.Ptr(packageORDID),
		Name:                                    "EVENT TITLE",
		Description:                             str.Ptr("long desc"),
		ShortDescription:                        str.Ptr("short desc"),
		SystemInstanceAware:                     &boolPtr,
		ChangeLogEntries:                        json.RawMessage(changeLogEntries),
		Links:                                   json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
		EventResourceLinks:                      json.RawMessage(fmt.Sprintf(resourceLinksFormat, baseURL)),
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
	}
}

func fixSuccessfulFetchRequest() *model.FetchRequest {
	return &model.FetchRequest{
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionSucceeded,
		},
	}
}

func fixEntityTypeMappingsEmpty() []*model.EntityTypeMapping {
	return []*model.EntityTypeMapping{}
}

func fixEmptyPackages() []*model.Package {
	return []*model.Package{}
}

func fixEmptyBundles() []*model.Bundle {
	return []*model.Bundle{}
}
