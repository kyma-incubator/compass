package ord

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/common"

	"github.com/mitchellh/hashstructure/v2"

	"golang.org/x/mod/semver"

	"github.com/google/go-cmp/cmp"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Disclaimer: All regexes below are provided by the ORD spec itself.
const (
	// VendorOrdIDRegex represents the valid structure of the ordID of the Vendor
	VendorOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(vendor):([a-zA-Z0-9._\\-]+):()$"
	// ProductOrdIDRegex represents the valid structure of the ordID of the Product
	ProductOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(product):([a-zA-Z0-9._\\-]+):()$"
	// BundleOrdIDRegex represents the valid structure of the ordID of the ConsumptionBundle
	BundleOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(consumptionBundle):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// EntityTypeOrdIDRegex represents the valid structure of the ordID of the EntityType
	EntityTypeOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(entityType):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// TombstoneOrdIDRegex represents the valid structure of the ordID of the Tombstone
	TombstoneOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(package|consumptionBundle|product|vendor|apiResource|eventResource|entityType|capability|integrationDependency|dataProduct):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*|)?$"
	// SystemInstanceBaseURLRegex represents the valid structure of the field
	SystemInstanceBaseURLRegex = "^http[s]?:\\/\\/[^:\\/\\s]+\\.[^:\\/\\s\\.]+(:\\d+)?(\\/[a-zA-Z0-9-\\._~]+)*$"
	// ConfigBaseURLRegex represents the valid structure of the field
	ConfigBaseURLRegex = "^http[s]?:\\/\\/[^:\\/\\s]+\\.[^:\\/\\s\\.]+(:\\d+)?(\\/[a-zA-Z0-9-\\._~]+)*$"
	// StringArrayElementRegex represents the valid structure of the field
	StringArrayElementRegex = "^[a-zA-Z0-9-_.\\/ ]*$"
	// CountryRegex represents the valid structure of the field
	CountryRegex = "^[A-Z]{2}$"
	// APIOrdIDRegex represents the valid structure of the ordID of the API
	APIOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(apiResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// EventOrdIDRegex represents the valid structure of the ordID of the Event
	EventOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(eventResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// CapabilityOrdIDRegex represents the valid structure of the ordID of the Capability
	CapabilityOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(capability):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// DataProductOrdIDRegex represents the valid structure of the ordID of the Data Product
	DataProductOrdIDRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(dataProduct):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// DataProductOutputPortsRegex represents the valid structure of the output ports of the Data Product
	DataProductOutputPortsRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(apiResource|eventResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// CorrelationIDsRegex represents the valid structure of the field
	CorrelationIDsRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):([a-zA-Z0-9._\\-\\/]+):([a-zA-Z0-9._\\-\\/]+)$"
	// LabelsKeyRegex represents the valid structure of the field
	LabelsKeyRegex = "^[a-zA-Z0-9-_.]*$"
	// NoNewLineRegex represents the valid structure of the field
	NoNewLineRegex = "^[^\\n]*$"
	// CustomImplementationStandardRegex represents the valid structure of the field
	CustomImplementationStandardRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	// VendorPartnersRegex represents the valid structure of the field
	VendorPartnersRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(vendor):([a-zA-Z0-9._\\-]+):()$"
	// CustomPolicyLevelRegex represents the valid structure of the field
	CustomPolicyLevelRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// CustomTypeCredentialExchangeStrategyRegex represents the valid structure of the field
	CustomTypeCredentialExchangeStrategyRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	// SAPProductOrdIDNamespaceRegex represents the valid structure of a SAP Product OrdID Namespace part
	SAPProductOrdIDNamespaceRegex = "^(sap)((\\.)([a-z0-9-]+(?:[.][a-z0-9-]+)*))*$"
	// CapabilityCustomTypeRegex represents the valid structure of a Capability custom type
	CapabilityCustomTypeRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	// ShortDescriptionSapCorePolicyRegex represents the valid structure of a short description field due to sap core policy
	ShortDescriptionSapCorePolicyRegex = "^([a-zA-Z0-9 _\\-.(),']*(S/4HANA|country/region|G/L)*[a-zA-Z0-9 _\\-.(),']*)$"

	// APISuccessorsRegex represents the valid structure of the API successors array items
	APISuccessorsRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(apiResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// EventSuccessorsRegex represents the valid structure of the Event successors array items
	EventSuccessorsRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(eventResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// IntegrationDependencySuccessorsRegex represents the valid structure of the Integration Dependency successors array items
	IntegrationDependencySuccessorsRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(integrationDependency):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// DataProductSuccessorsRegex represents the valid structure of the Integration Dependency successors array items
	DataProductSuccessorsRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):(dataProduct):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"

	// ResponsibleRegex represents the valid structure of the `responsible` field for API, Event and Data Product
	ResponsibleRegex = "^([a-z0-9]+(?:[.][a-z0-9]+)*):([a-zA-Z0-9._\\-\\/]+):([a-zA-Z0-9._\\-\\/]+)$"

	// MinDescriptionLength represents the minimal accepted length of the Description field
	MinDescriptionLength = 1
	// MaxDescriptionLength represents the maximal accepted length of the Description field
	MaxDescriptionLength = 5000
	// MaxDescriptionLengthEntityType represents the maximal accepted length of the Description field for Entity Type when policy level is sap:core:v1
	MaxDescriptionLengthEntityType = 1000

	// MinShortDescriptionLength represents the minimal length of ShortDescription field
	MinShortDescriptionLength = 1
	// MaxShortDescriptionRuneLength represents the maximal length of ShortDescription field
	MaxShortDescriptionRuneLength = 256
	// MaxShortDescriptionLengthSapCorePolicy represents the maximal length of ShortDescription field for sap:core:v1 policy
	MaxShortDescriptionLengthSapCorePolicy = 180
	// MinLocalTenantIDLength represents the minimal accepted length of the LocalID field
	MinLocalTenantIDLength = 1
	// MaxLocalTenantIDLength represents the maximal accepted length of the LocalID field
	MaxLocalTenantIDLength = 255
	// MinSystemVersionTitleLength represents the minimal accepted length of the LocalID field
	MinSystemVersionTitleLength = 1
	// MaxSystemVersionTitleLength represents the maximal accepted length of the LocalID field
	MaxSystemVersionTitleLength = 255
	// MinTitleLength represents the minimal accepted length of the Title field
	MinTitleLength = 1
	// MaxTitleLength represents the maximal accepted length of the Title field
	MaxTitleLength = 255
	// MinLevelLength represents the minimal accepted length of the Level field
	MinLevelLength = 1
	// MaxLevelLength represents the maximal accepted length of the Level field
	MaxLevelLength = 255
	// MaxTitleLengthSAPCorePolicy represents the maximal accepted length of the Title field due to sap core policy
	MaxTitleLengthSAPCorePolicy = 120
	// MinOrdIDLength represents the minimal accepted length of the OrdID field
	MinOrdIDLength = 1
	// MaxOrdIDLength represents the maximal accepted length of the OrdID field
	MaxOrdIDLength = 255
	// MinOrdPackageIDLength represents the minimal accepted length of the ordPackageID field
	MinOrdPackageIDLength = 1
	// MaxOrdPackageIDLength represents the maximal accepted length of the ordPackageID field
	MaxOrdPackageIDLength = 255
	// MinResourceLinkCustomTypeLength represents the minimal accepted length of the custom type field in a resource link
	MinResourceLinkCustomTypeLength = 1
	// MaxResourceLinkCustomTypeLength represents the maximal accepted length of the custom type field in a resource link
	MaxResourceLinkCustomTypeLength = 255
	// MinCorrelationIDLength represents the minimal accepted length of the Correlation ID field
	MinCorrelationIDLength = 1
	// MaxCorrelationIDLength represents the maximal accepted length of the Correlation ID field
	MaxCorrelationIDLength = 255
	// MinResponsibleLength represents the minimal accepted length of the Correlation ID field
	MinResponsibleLength = 1
	// MaxResponsibleLength represents the maximal accepted length of the Correlation ID field
	MaxResponsibleLength = 255

	// IntegrationDependencyMsg represents the resource name for Integration Dependency used in error message
	IntegrationDependencyMsg string = "integration dependency"
)

const (
	custom   string = "custom"
	none     string = "none"
	public   string = "public"
	private  string = "private"
	internal string = "internal"

	// PolicyLevelSap is one of the available policy options
	PolicyLevelSap string = "sap:core:v1"
	// PolicyLevelCustom is one of the available policy options
	PolicyLevelCustom = custom
	// PolicyLevelNone is one of the available policy options
	PolicyLevelNone string = none

	// APIProtocolODataV2 is one of the available api protocol options
	APIProtocolODataV2 string = "odata-v2"
	// APIProtocolODataV4 is one of the available api protocol options
	APIProtocolODataV4 string = "odata-v4"
	// APIProtocolSoapInbound is one of the available api protocol options
	APIProtocolSoapInbound string = "soap-inbound"
	// APIProtocolSoapOutbound is one of the available api protocol options
	APIProtocolSoapOutbound string = "soap-outbound"
	// APIProtocolRest is one of the available api protocol options
	APIProtocolRest string = "rest"
	// APIProtocolSapRfc is one of the available api protocol options
	APIProtocolSapRfc string = "sap-rfc"
	// APIProtocolWebsocket is one of the available api protocol options
	APIProtocolWebsocket string = "websocket"
	// APIProtocolSAPSQLAPIV1 is one of the available api protocol options
	APIProtocolSAPSQLAPIV1 string = "sap-sql-api-v1"
	// APIProtocolGraphql is one of the available api protocol options
	APIProtocolGraphql string = "graphql"
	// APIProtocolDeltaSharing is one of the available api protocol options
	APIProtocolDeltaSharing string = "delta-sharing"
	// APIProtocolSapInaAPIV1 is one of the available api protocol options
	APIProtocolSapInaAPIV1 string = "sap-ina-api-v1"

	// APIVisibilityPublic is one of the available api visibility options
	APIVisibilityPublic = public
	// APIVisibilityPrivate is one of the available api visibility options
	APIVisibilityPrivate = private
	// APIVisibilityInternal is one of the available api visibility options
	APIVisibilityInternal = internal

	// EventVisibilityPublic is one of the available event visibility options
	EventVisibilityPublic = public
	// EventVisibilityPrivate is one of the available event visibility options
	EventVisibilityPrivate = private
	// EventVisibilityInternal is one of the available event visibility options
	EventVisibilityInternal = internal

	// CapabilityVisibilityPublic is one of the available Capability visibility options
	CapabilityVisibilityPublic = public
	// CapabilityVisibilityPrivate is one of the available Capability visibility options
	CapabilityVisibilityPrivate = private
	// CapabilityVisibilityInternal is one of the available Capability visibility options
	CapabilityVisibilityInternal = internal

	// IntegrationDependencyVisibilityPublic is one of the available Integration Dependency visibility options
	IntegrationDependencyVisibilityPublic = public
	// IntegrationDependencyVisibilityPrivate is one of the available Integration Dependency visibility options
	IntegrationDependencyVisibilityPrivate = private
	// IntegrationDependencyVisibilityInternal is one of the available Integration Dependency visibility options
	IntegrationDependencyVisibilityInternal = internal

	// DataProductVisibilityPublic is one of the available Data Product visibility options
	DataProductVisibilityPublic = public
	// DataProductVisibilityPrivate is one of the available Data Product visibility options
	DataProductVisibilityPrivate = private
	// DataProductVisibilityInternal is one of the available Data Product visibility options
	DataProductVisibilityInternal = internal

	// APIImplementationStandardDocumentAPI is one of the available api implementation standard options
	APIImplementationStandardDocumentAPI string = "sap:ord-document-api:v1"
	// APIImplementationStandardServiceBroker is one of the available api implementation standard options
	APIImplementationStandardServiceBroker string = "cff:open-service-broker:v2"
	// APIImplementationStandardCsnExposure is one of the available api implementation standard options
	APIImplementationStandardCsnExposure string = "sap:csn-exposure:v1"
	// APIImplementationStandardApeAPI is one of the available api implementation standard options
	APIImplementationStandardApeAPI string = "sap:ape-api:v1"
	// APIImplementationStandardCdiAPI is one of the available api implementation standard options
	APIImplementationStandardCdiAPI string = "sap:cdi-api:v1"
	// APIImplementationStandardHdlfDeltaSharing is one of the available api implementation standard options
	APIImplementationStandardHdlfDeltaSharing string = "sap:hdlf-delta-sharing:v1"
	// APIImplementationStandardHanaCloudSQL is one of the available api implementation standard options
	APIImplementationStandardHanaCloudSQL string = "sap:hana-cloud-sql:v1"
	// APIImplementationStandardCustom is one of the available api implementation standard options
	APIImplementationStandardCustom = custom

	// EventImplementationStandardCustom is one of the available event implementation standard options
	EventImplementationStandardCustom = custom

	// APIDirectionInbound is one of the available direction options
	APIDirectionInbound = "inbound"
	// APIDirectionMixed is one of the available direction options
	APIDirectionMixed = "mixed"
	// APIDirectionOutbound is one of the available direction options
	APIDirectionOutbound = "outbound"

	// SapVendor is a valid Vendor ordID
	SapVendor = "sap:vendor:SAP:"
	// PartnerVendor is a valid partner Vendor ordID
	PartnerVendor = "partner:vendor:SAP:"

	// CapabilityTypeCustom is one of the available Capability type options
	CapabilityTypeCustom = custom
	// CapabilityTypeMDICapabilityV1 is the MDI Capability V1 Specification
	CapabilityTypeMDICapabilityV1 string = "sap.mdo:mdi-capability:v1"

	// APIModelSelectorTypeODATA for odata selector type.
	APIModelSelectorTypeODATA = "odata"
	// APIModelSelectorTypeJSONPointer for json pointer selector type.
	APIModelSelectorTypeJSONPointer = "json-pointer"

	// PackageRuntimeRestriction is one of the available RuntimeRestriction options for Package
	PackageRuntimeRestriction = "sap.datasphere"

	// APIUsageExternal is one of the available Usage options for API
	APIUsageExternal = "external"
	// APIUsageLocal is one of the available Usage options for API
	APIUsageLocal = "local"

	// DataProductTypeBase is one of the available Type options for Data Product
	DataProductTypeBase = "base"
	// DataProductTypeDerived is one of the available Type options for Data Product
	DataProductTypeDerived = "derived"

	// DataProductCategoryBusinessObject is one of the available Category options for Data Product
	DataProductCategoryBusinessObject = "business-object"
	// DataProductCategoryAnalytical is one of the available Category options for Data Product
	DataProductCategoryAnalytical = "analytical"
	// DataProductCategoryOther is one of the available Category options for Data Product
	DataProductCategoryOther = "other"
)

var (
	// LineOfBusinesses contain all valid values for this field from the spec
	LineOfBusinesses = map[string]bool{
		"Asset Management":                 true,
		"Commerce":                         true,
		"Finance":                          true,
		"Human Resources":                  true,
		"Manufacturing":                    true,
		"Marketing":                        true,
		"R&D Engineering":                  true,
		"Sales":                            true,
		"Service":                          true,
		"Sourcing and Procurement":         true,
		"Supply Chain":                     true,
		"Sustainability":                   true,
		"Metering":                         true,
		"Grid Operations and Maintenance":  true,
		"Plant Operations and Maintenance": true,
		"Maintenance and Engineering":      true,
	}
	// Industries contain all valid values for this field from the spec
	Industries = map[string]bool{
		"Aerospace and Defense": true,
		"Automotive":            true,
		"Banking":               true,
		"Chemicals":             true,
		"Consumer Products":     true,
		"Defense and Security":  true,
		"Engineering Construction and Operations": true,
		"Healthcare":                          true,
		"Higher Education and Research":       true,
		"High Tech":                           true,
		"Industrial Machinery and Components": true,
		"Insurance":                           true,
		"Life Sciences":                       true,
		"Media":                               true,
		"Mill Products":                       true,
		"Mining":                              true,
		"Oil and Gas":                         true,
		"Professional Services":               true,
		"Public Sector":                       true,
		"Retail":                              true,
		"Sports and Entertainment":            true,
		"Telecommunications":                  true,
		"Travel and Transportation":           true,
		"Utilities":                           true,
		"Wholesale Distribution":              true,
	}
	// SupportedUseCases contain all valid values for this field from the spec
	SupportedUseCases = map[string]bool{
		"data-federation": true,
		"snapshot":        true,
		"incremental":     true,
		"streaming":       true,
	}

	forbiddenTermsInTitle = []string{
		"deprecated",
		"decommissioned",
		"create",
		"read",
		"delete",
		"update",
	}

	apiEventResourceLinkTypes = []interface{}{
		"api-documentation", "authentication", "client-registration", "console", "payment", "service-level-agreement", "support", "custom",
	}

	dataProductResourceLinkTypes = []interface{}{
		"payment", "service-level-agreement", "support", "custom",
	}
)

func titleRules(docPolicyLevel, resourcePolicyLevel *string) []validation.Rule {
	return []validation.Rule{
		validation.Required, validation.NewStringRule(common.NoNewLines, "title should not contain line breaks"),
		validation.When(checkResourcePolicyLevel(docPolicyLevel, resourcePolicyLevel, PolicyLevelSap), validation.Length(MinTitleLength, MaxTitleLengthSAPCorePolicy), validation.By(validateTitleDoesNotContainsTerms)),
		validation.Length(MinTitleLength, MaxTitleLength),
	}
}

func descriptionRules(docPolicyLevel, resourcePolicyLevel, resourceShortDescription *string) []validation.Rule {
	return []validation.Rule{
		validation.Required,
		validation.When(checkResourcePolicyLevel(docPolicyLevel, resourcePolicyLevel, PolicyLevelSap) && resourceShortDescription != nil, validation.By(validateDescriptionDoesNotContainShortDescription(resourceShortDescription))),
		validation.Length(MinDescriptionLength, MaxDescriptionLength),
	}
}

func optionalDescriptionRules(docPolicyLevel, resourceShortDescription *string) []validation.Rule {
	return []validation.Rule{
		validation.NilOrNotEmpty,
		validation.When(checkResourcePolicyLevel(docPolicyLevel, nil, PolicyLevelSap) && resourceShortDescription != nil, validation.By(validateDescriptionDoesNotContainShortDescription(resourceShortDescription))),
		validation.Length(MinDescriptionLength, MaxDescriptionLength),
	}
}

func shortDescriptionRules(docPolicyLevel, resourcePolicyLevel *string, resourceName string) []validation.Rule {
	return []validation.Rule{
		validation.Required, validation.NewStringRule(common.NoNewLines, "short description should not contain line breaks"),
		validation.When(checkResourcePolicyLevel(docPolicyLevel, resourcePolicyLevel, PolicyLevelSap), validation.Match(regexp.MustCompile(ShortDescriptionSapCorePolicyRegex)), validation.Length(MinShortDescriptionLength, MaxShortDescriptionLengthSapCorePolicy), validation.By(validateShortDescriptionDoesNotStartWithResourceName(resourceName))),
		validation.RuneLength(1, 256),
	}
}

func optionalShortDescriptionRules(docPolicyLevel, resourcePolicyLevel *string, resourceName string) []validation.Rule {
	return []validation.Rule{
		validation.NilOrNotEmpty, validation.NewStringRule(common.NoNewLines, "short description should not contain line breaks"),
		validation.When(checkResourcePolicyLevel(docPolicyLevel, resourcePolicyLevel, PolicyLevelSap), validation.Match(regexp.MustCompile(ShortDescriptionSapCorePolicyRegex)), validation.Length(MinShortDescriptionLength, MaxShortDescriptionLengthSapCorePolicy), validation.By(validateShortDescriptionDoesNotStartWithResourceName(resourceName))),
		validation.RuneLength(1, 256),
	}
}

func correlationIdsRules() []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CorrelationIDsRegex))
		}),
	}
}

func partOfProductsRules() []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ProductOrdIDRegex))
		}),
	}
}

func tagsRules() []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		}),
	}
}

func countriesRules() []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CountryRegex))
		}),
	}
}

func lineOfBusinessRules(docPolicyLevel, resourcePolicyLevel *string) []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			return validateWhenPolicyLevelIsSAP(docPolicyLevel, resourcePolicyLevel, func() error {
				return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
			})
		}),
		validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		}),
	}
}

func industryRules(docPolicyLevel, resourcePolicyLevel *string) []validation.Rule {
	return []validation.Rule{
		validation.By(func(value interface{}) error {
			return validateWhenPolicyLevelIsSAP(docPolicyLevel, resourcePolicyLevel, func() error {
				return validateJSONArrayOfStringsContainsInMap(value, Industries)
			})
		}),
		validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		}),
	}
}

// ORDDocumentValidationError contains the validation errors when aggregating ord documents
type ORDDocumentValidationError struct {
	Err error
}

func (e *ORDDocumentValidationError) Error() string {
	return e.Err.Error()
}

// ValidateSystemInstanceInput validates the given SystemInstance
func ValidateSystemInstanceInput(app *model.Application) error {
	return validation.ValidateStruct(app,
		validation.Field(&app.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&app.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&app.BaseURL, is.RequestURI, validation.Match(regexp.MustCompile(SystemInstanceBaseURLRegex))),
		validation.Field(&app.OrdLabels, validation.By(validateORDLabels)),
		validation.Field(&app.Tags, tagsRules()...),
		validation.Field(&app.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

// ValidateSystemVersionInput validates the given SystemVersion
func ValidateSystemVersionInput(appTemplateVersion *model.ApplicationTemplateVersionInput) error {
	return validation.ValidateStruct(appTemplateVersion,
		validation.Field(&appTemplateVersion.Title, validation.NilOrNotEmpty, validation.Length(MinSystemVersionTitleLength, MaxSystemVersionTitleLength), validation.Match(regexp.MustCompile(NoNewLineRegex))),
		validation.Field(&appTemplateVersion.ReleaseDate, validation.Required),
	)
}

func validateDocumentInput(doc *Document) error {
	return validation.ValidateStruct(doc, validation.Field(&doc.OpenResourceDiscovery, validation.Required, validation.Match(regexp.MustCompile(`^1\.\d$`))),
		validation.Field(&doc.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelCustom, PolicyLevelNone), validation.When(doc.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&doc.CustomPolicyLevel, validation.When(doc.PolicyLevel != nil && *doc.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
	)
}

func validatePackageInput(pkg *model.PackageInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&pkg.Title, titleRules(docPolicyLevel, pkg.PolicyLevel)...),
		validation.Field(&pkg.ShortDescription, shortDescriptionRules(docPolicyLevel, pkg.PolicyLevel, pkg.Title)...),
		validation.Field(&pkg.Description, descriptionRules(docPolicyLevel, pkg.PolicyLevel, &pkg.ShortDescription)...),
		validation.Field(&pkg.SupportInfo, validation.NilOrNotEmpty),
		validation.Field(&pkg.Version, validation.Required, validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&pkg.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelCustom, PolicyLevelNone), validation.When(pkg.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&pkg.CustomPolicyLevel, validation.When(pkg.PolicyLevel != nil && *pkg.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&pkg.PackageLinks, validation.By(validatePackageLinks)),
		validation.Field(&pkg.Links, validation.By(validateORDLinks)),
		validation.Field(&pkg.Vendor, validation.Required,
			validation.When(checkResourcePolicyLevel(docPolicyLevel, pkg.PolicyLevel, PolicyLevelSap), validation.In(SapVendor)),
			validation.Match(regexp.MustCompile(VendorOrdIDRegex)), validation.Length(1, 256)),
		validation.Field(&pkg.PartOfProducts, partOfProductsRules()...),
		validation.Field(&pkg.Tags, tagsRules()...),
		validation.Field(&pkg.RuntimeRestriction, validation.NilOrNotEmpty, validation.In(PackageRuntimeRestriction)),
		validation.Field(&pkg.Labels, validation.By(validateORDLabels)),
		validation.Field(&pkg.Countries, countriesRules()...),
		validation.Field(&pkg.LineOfBusiness, lineOfBusinessRules(docPolicyLevel, pkg.PolicyLevel)...),
		validation.Field(&pkg.Industry, industryRules(docPolicyLevel, pkg.PolicyLevel)...),
		validation.Field(&pkg.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func checkResourcePolicyLevel(docPolicyLevel *string, resourcePolicyLevel *string, policyLevelValue string) bool {
	policyLevel := str.PtrStrToStr(docPolicyLevel)
	if resourcePolicyLevel != nil {
		policyLevel = str.PtrStrToStr(resourcePolicyLevel)
	}

	return policyLevel == policyLevelValue
}

func validateTitleDoesNotContainsTerms(value interface{}) error {
	title, ok := value.(string)
	if !ok {
		return errors.New("title should be a string")
	}

	titleLower := strings.ToLower(title)

	var existingTermsInTitle []string

	for _, term := range forbiddenTermsInTitle {
		if strings.Contains(titleLower, term) {
			existingTermsInTitle = append(existingTermsInTitle, term)
		}
	}

	if len(existingTermsInTitle) != 0 {
		return errors.New(fmt.Sprintf("title must not contain the terms %q", existingTermsInTitle))
	}

	return nil
}

func validateShortDescriptionDoesNotStartWithResourceName(name string) func(value interface{}) error {
	return func(value interface{}) error {
		var shortDescription string

		// this is need, because in the package the value is string, but in apis and events the value is *string
		switch v := value.(type) {
		case *string:
			if v == nil {
				return nil
			}
			shortDescription = *v
		case string:
			shortDescription = v
		default:
			return errors.New("short description should be a string or string pointer")
		}

		if strings.HasPrefix(shortDescription, name) {
			return errors.New("short description must not start with resource name")
		}

		return nil
	}
}

func validateDescriptionDoesNotContainShortDescription(shortDescription *string) func(value interface{}) error {
	return func(value interface{}) error {
		var description string

		// this is need, because in the package the value is string, but in apis and events the value is *string
		switch v := value.(type) {
		case *string:
			if v == nil {
				return nil
			}
			description = *v
		case string:
			description = v
		default:
			return errors.New("description should be a string or string pointer")
		}

		descriptionLower := strings.ToLower(description)
		shortDescriptionLower := strings.ToLower(*shortDescription)
		if strings.Contains(descriptionLower, shortDescriptionLower) {
			return errors.New("description must not contain short description")
		}

		return nil
	}
}

func validatePackageInputWithSuppressedErrors(pkg *model.PackageInput, packagesFromDB map[string]*model.Package, resourceHashes map[string]uint64) error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.Version, validation.By(func(value interface{}) error {
			return validatePackageVersionInput(value, *pkg, packagesFromDB, resourceHashes)
		})))
}

func validateBundleInput(bndl *model.BundleCreateInput, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping, docPolicyLevel *string) error {
	return validation.ValidateStruct(bndl,
		validation.Field(&bndl.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(BundleOrdIDRegex))),
		validation.Field(&bndl.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&bndl.Name, titleRules(docPolicyLevel, nil)...),
		validation.Field(&bndl.ShortDescription, optionalShortDescriptionRules(docPolicyLevel, nil, bndl.Name)...),
		validation.Field(&bndl.Description, optionalDescriptionRules(docPolicyLevel, bndl.ShortDescription)...),
		validation.Field(&bndl.Version, validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&bndl.Links, validation.By(validateORDLinks)),
		validation.Field(&bndl.Labels, validation.By(validateORDLabels)),
		validation.Field(&bndl.CredentialExchangeStrategies, validation.By(func(value interface{}) error {
			return common.ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
				"type": {
					validation.Required,
					validation.In(custom),
				},
				"callbackUrl": {
					is.RequestURI,
				},
			}, validateCustomType(credentialExchangeStrategyTenantMappings), validateCustomDescription)
		})),
		validation.Field(&bndl.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&bndl.Tags, tagsRules()...),
		validation.Field(&bndl.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func validateBundleInputWithSuppressedErrors(bndl *model.BundleCreateInput, bundlesFromDB map[string]*model.Bundle, resourceHashes map[string]uint64) error {
	return validation.ValidateStruct(bndl,
		validation.Field(&bndl.Version, validation.By(func(value interface{}) error {
			return validateBundleVersionInput(value, *bndl, bundlesFromDB, resourceHashes)
		})))
}

func validateAPIInput(api *model.APIDefinitionInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(api,
		validation.Field(&api.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(APIOrdIDRegex))),
		validation.Field(&api.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&api.Name, titleRules(docPolicyLevel, api.PolicyLevel)...),
		validation.Field(&api.ShortDescription, shortDescriptionRules(docPolicyLevel, api.PolicyLevel, api.Name)...),
		validation.Field(&api.Description, descriptionRules(docPolicyLevel, api.PolicyLevel, api.ShortDescription)...),
		validation.Field(&api.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelCustom, PolicyLevelNone), validation.When(api.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&api.CustomPolicyLevel, validation.When(api.PolicyLevel != nil && *api.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&api.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&api.OrdPackageID, validation.Required, validation.Length(MinOrdPackageIDLength, MaxOrdPackageIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&api.APIProtocol, validation.Required, validation.In(APIProtocolODataV2, APIProtocolODataV4, APIProtocolSoapInbound, APIProtocolSoapOutbound, APIProtocolRest, APIProtocolSapRfc, APIProtocolWebsocket, APIProtocolSAPSQLAPIV1, APIProtocolGraphql, APIProtocolDeltaSharing, APIProtocolSapInaAPIV1)),
		validation.Field(&api.Visibility, validation.Required, validation.In(APIVisibilityPublic, APIVisibilityInternal, APIVisibilityPrivate)),
		validation.Field(&api.PartOfProducts, partOfProductsRules()...),
		validation.Field(&api.SupportedUseCases,
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsContainsInMap(value, SupportedUseCases)
			}),
		),
		validation.Field(&api.Tags, tagsRules()...),
		validation.Field(&api.Countries, countriesRules()...),
		validation.Field(&api.LineOfBusiness, lineOfBusinessRules(docPolicyLevel, api.PolicyLevel)...),
		validation.Field(&api.Industry, industryRules(docPolicyLevel, api.PolicyLevel)...),
		validation.Field(&api.ResourceDefinitions, validation.By(func(value interface{}) error {
			return validateAPIResourceDefinitions(value, *api, docPolicyLevel)
		})),
		validation.Field(&api.APIResourceLinks, validation.By(func(value interface{}) error {
			return validateResourceLinks(value, apiEventResourceLinkTypes)
		})),
		validation.Field(&api.Links, validation.By(validateORDLinks)),
		validation.Field(&api.ReleaseStatus, validation.Required, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&api.SunsetDate, validation.When(*api.ReleaseStatus == common.ReleaseStatusDeprecated, validation.Required), validation.When(api.SunsetDate != nil, validation.By(isValidDate))),
		validation.Field(&api.Successors, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(APISuccessorsRegex))
		})),
		validation.Field(&api.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&api.TargetURLs, validation.By(validateEntryPoints), validation.When(api.TargetURLs == nil, validation.By(notPartOfConsumptionBundles(api.PartOfConsumptionBundles)))),
		validation.Field(&api.Labels, validation.By(validateORDLabels)),
		validation.Field(&api.ImplementationStandard, validation.In(APIImplementationStandardDocumentAPI, APIImplementationStandardServiceBroker, APIImplementationStandardCsnExposure, APIImplementationStandardApeAPI, APIImplementationStandardCdiAPI, APIImplementationStandardHdlfDeltaSharing, APIImplementationStandardHanaCloudSQL, APIImplementationStandardCustom)),
		validation.Field(&api.CustomImplementationStandard, validation.When(api.ImplementationStandard != nil && *api.ImplementationStandard == APIImplementationStandardCustom, validation.Required, validation.Match(regexp.MustCompile(CustomImplementationStandardRegex))).Else(validation.Empty)),
		validation.Field(&api.CustomImplementationStandardDescription, validation.When(api.ImplementationStandard != nil && *api.ImplementationStandard == APIImplementationStandardCustom, validation.Required).Else(validation.Empty)),
		validation.Field(&api.PartOfConsumptionBundles, validation.By(func(value interface{}) error {
			return validateAPIPartOfConsumptionBundles(value, api.TargetURLs, regexp.MustCompile(BundleOrdIDRegex))
		})),
		validation.Field(&api.DefaultConsumptionBundle, validation.Match(regexp.MustCompile(BundleOrdIDRegex)), validation.By(func(value interface{}) error {
			return validateDefaultConsumptionBundle(value, api.PartOfConsumptionBundles)
		})),
		validation.Field(&api.Extensible, validation.By(func(value interface{}) error {
			return validateExtensibleField(value, docPolicyLevel, true)
		})),
		validation.Field(&api.EntityTypeMappings, validation.By(validateEntityTypeMappings)),
		validation.Field(&api.DocumentationLabels, validation.By(validateDocumentationLabels)),
		validation.Field(&api.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&api.Direction, validation.In(APIDirectionInbound, APIDirectionMixed, APIDirectionOutbound)),
		validation.Field(&api.LastUpdate, validation.When(api.LastUpdate != nil, validation.By(isValidDate))),
		validation.Field(&api.DeprecationDate, validation.NilOrNotEmpty, validation.When(api.DeprecationDate != nil, validation.By(isValidDate))),
		validation.Field(&api.Responsible, validation.NilOrNotEmpty, validation.Length(MinResponsibleLength, MaxResponsibleLength), validation.Match(regexp.MustCompile(ResponsibleRegex))),
		validation.Field(&api.Usage, validation.NilOrNotEmpty, validation.In(APIUsageExternal, APIUsageLocal)),
	)
}

// fields with validation errors will lead to persisting of the API resource
func validateAPIInputWithSuppressedErrors(api *model.APIDefinitionInput, apisFromDB map[string]*model.APIDefinition, apiHashes map[string]uint64) error {
	return validation.ValidateStruct(api,
		validation.Field(&api.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateAPIDefinitionVersionInput(value, *api, apisFromDB, apiHashes)
		})))
}

func validateEventInput(event *model.EventDefinitionInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(event,
		validation.Field(&event.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(EventOrdIDRegex))),
		validation.Field(&event.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&event.Name, titleRules(docPolicyLevel, event.PolicyLevel)...),
		validation.Field(&event.ShortDescription, shortDescriptionRules(docPolicyLevel, event.PolicyLevel, event.Name)...),
		validation.Field(&event.Description, descriptionRules(docPolicyLevel, event.PolicyLevel, event.ShortDescription)...),
		validation.Field(&event.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelCustom, PolicyLevelNone), validation.When(event.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&event.CustomPolicyLevel, validation.When(event.PolicyLevel != nil && *event.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&event.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&event.OrdPackageID, validation.Required, validation.Length(MinOrdPackageIDLength, MaxOrdPackageIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&event.Visibility, validation.Required, validation.In(EventVisibilityPublic, EventVisibilityInternal, EventVisibilityPrivate)),
		validation.Field(&event.PartOfProducts, partOfProductsRules()...),
		validation.Field(&event.Tags, tagsRules()...),
		validation.Field(&event.Countries, countriesRules()...),
		validation.Field(&event.LineOfBusiness, lineOfBusinessRules(docPolicyLevel, event.PolicyLevel)...),
		validation.Field(&event.Industry, industryRules(docPolicyLevel, event.PolicyLevel)...),
		validation.Field(&event.ResourceDefinitions, validation.By(func(value interface{}) error {
			return validateEventResourceDefinition(value, *event, docPolicyLevel)
		})),
		validation.Field(&event.Links, validation.By(validateORDLinks)),
		validation.Field(&event.EventResourceLinks, validation.By(func(value interface{}) error {
			return validateResourceLinks(value, apiEventResourceLinkTypes)
		})),
		validation.Field(&event.ReleaseStatus, validation.Required, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&event.SunsetDate, validation.When(*event.ReleaseStatus == common.ReleaseStatusDeprecated, validation.Required), validation.When(event.SunsetDate != nil, validation.By(isValidDate))),
		validation.Field(&event.Successors, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(EventSuccessorsRegex))
		})),
		validation.Field(&event.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&event.Labels, validation.By(validateORDLabels)),
		validation.Field(&event.PartOfConsumptionBundles, validation.By(func(value interface{}) error {
			return validateEventPartOfConsumptionBundles(value, regexp.MustCompile(BundleOrdIDRegex))
		})),
		validation.Field(&event.DefaultConsumptionBundle, validation.Match(regexp.MustCompile(BundleOrdIDRegex)), validation.By(func(value interface{}) error {
			return validateDefaultConsumptionBundle(value, event.PartOfConsumptionBundles)
		})),
		validation.Field(&event.ImplementationStandard, validation.In(EventImplementationStandardCustom)),
		validation.Field(&event.CustomImplementationStandard, validation.When(event.ImplementationStandard != nil && *event.ImplementationStandard == EventImplementationStandardCustom, validation.Required, validation.Match(regexp.MustCompile(CustomImplementationStandardRegex))).Else(validation.Empty)),
		validation.Field(&event.CustomImplementationStandardDescription, validation.When(event.ImplementationStandard != nil && *event.ImplementationStandard == EventImplementationStandardCustom, validation.Required).Else(validation.Empty)),
		validation.Field(&event.Extensible, validation.By(func(value interface{}) error {
			return validateExtensibleField(value, docPolicyLevel, true)
		})),
		validation.Field(&event.EntityTypeMappings, validation.By(validateEntityTypeMappings)),
		validation.Field(&event.DocumentationLabels, validation.By(validateDocumentationLabels)),
		validation.Field(&event.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&event.LastUpdate, validation.When(event.LastUpdate != nil, validation.By(isValidDate))),
		validation.Field(&event.DeprecationDate, validation.NilOrNotEmpty, validation.When(event.DeprecationDate != nil, validation.By(isValidDate))),
		validation.Field(&event.Responsible, validation.NilOrNotEmpty, validation.Length(MinResponsibleLength, MaxResponsibleLength), validation.Match(regexp.MustCompile(ResponsibleRegex))),
	)
}

func validateEventInputWithSuppressedErrors(event *model.EventDefinitionInput, eventsFromDB map[string]*model.EventDefinition, eventHashes map[string]uint64) error {
	return validation.ValidateStruct(event,
		validation.Field(&event.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateEventDefinitionVersionInput(value, *event, eventsFromDB, eventHashes)
		})))
}

func validateEntityTypeInputWithSuppressedErrors(entityType *model.EntityTypeInput, entityTypesFromDB map[string]*model.EntityType, entityTypeHashes map[string]uint64) error {
	return validation.ValidateStruct(entityType,
		validation.Field(&entityType.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateEntityTypeVersionInput(value, *entityType, entityTypesFromDB, entityTypeHashes)
		})))
}

func validateEntityTypeInput(entityType *model.EntityTypeInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(entityType,
		validation.Field(&entityType.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(EntityTypeOrdIDRegex))),
		validation.Field(&entityType.LocalTenantID, validation.Required, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&entityType.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&entityType.Level, validation.Required, validation.Length(MinLevelLength, MaxLevelLength)),
		validation.Field(&entityType.Title, titleRules(docPolicyLevel, entityType.PolicyLevel)...),
		validation.Field(&entityType.ShortDescription, optionalShortDescriptionRules(docPolicyLevel, entityType.PolicyLevel, entityType.Title)...),
		validation.Field(&entityType.Description, validation.NilOrNotEmpty,
			validation.When(checkResourcePolicyLevel(docPolicyLevel, entityType.PolicyLevel, PolicyLevelSap) && entityType.ShortDescription != nil, validation.By(validateDescriptionDoesNotContainShortDescription(entityType.ShortDescription)), validation.Length(MinDescriptionLength, MaxDescriptionLengthEntityType)),
			validation.Length(MinDescriptionLength, MaxDescriptionLength)),
		validation.Field(&entityType.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&entityType.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&entityType.OrdPackageID, validation.Required, validation.Length(MinOrdPackageIDLength, MaxOrdPackageIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&entityType.Visibility, validation.Required, validation.In(APIVisibilityPublic, APIVisibilityInternal, APIVisibilityPrivate)),
		validation.Field(&entityType.Links, validation.By(validateORDLinks)),
		validation.Field(&entityType.PartOfProducts, partOfProductsRules()...),
		validation.Field(&entityType.LastUpdate, validation.When(entityType.LastUpdate != nil, validation.By(isValidDate))),
		validation.Field(&entityType.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelCustom, PolicyLevelNone), validation.When(entityType.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&entityType.CustomPolicyLevel, validation.When(entityType.PolicyLevel != nil && *entityType.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&entityType.ReleaseStatus, validation.Required, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&entityType.SunsetDate, validation.When(entityType.ReleaseStatus == common.ReleaseStatusDeprecated, validation.Required), validation.When(entityType.SunsetDate != nil, validation.By(isValidDate))),
		validation.Field(&entityType.DeprecationDate, validation.NilOrNotEmpty, validation.When(entityType.DeprecationDate != nil, validation.By(isValidDate))),
		validation.Field(&entityType.Successors, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(EntityTypeOrdIDRegex))
		})),
		validation.Field(&entityType.Extensible, validation.By(func(value interface{}) error {
			return validateExtensibleField(value, docPolicyLevel, false)
		})),
		validation.Field(&entityType.Tags, tagsRules()...),
		validation.Field(&entityType.Labels, validation.By(validateORDLabels)),
		validation.Field(&entityType.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func validateCapabilityInput(capability *model.CapabilityInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(capability,
		validation.Field(&capability.OrdPackageID, validation.Required, validation.Length(MinOrdPackageIDLength, MaxOrdPackageIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&capability.Name, titleRules(docPolicyLevel, nil)...),
		validation.Field(&capability.Description, optionalDescriptionRules(docPolicyLevel, capability.ShortDescription)...),
		validation.Field(&capability.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(CapabilityOrdIDRegex))),
		validation.Field(&capability.Type, validation.Required, validation.In(CapabilityTypeCustom, CapabilityTypeMDICapabilityV1), validation.When(capability.CustomType != nil, validation.In(CapabilityTypeCustom))),
		validation.Field(&capability.CustomType, validation.When(capability.Type != CapabilityTypeCustom, validation.Empty), validation.Match(regexp.MustCompile(CapabilityCustomTypeRegex))),
		validation.Field(&capability.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&capability.ShortDescription, optionalShortDescriptionRules(docPolicyLevel, nil, capability.Name)...),
		validation.Field(&capability.Tags, tagsRules()...),
		validation.Field(&capability.RelatedEntityTypes, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(EntityTypeOrdIDRegex))
		})),
		validation.Field(&capability.Links, validation.By(validateORDLinks)),
		validation.Field(&capability.ReleaseStatus, validation.Required, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&capability.Labels, validation.By(validateORDLabels)),
		validation.Field(&capability.Visibility, validation.Required, validation.In(CapabilityVisibilityPublic, CapabilityVisibilityInternal, CapabilityVisibilityPrivate)),
		validation.Field(&capability.CapabilityDefinitions, validation.By(func(value interface{}) error {
			return validateCapabilityDefinitions(value, *capability)
		})),
		validation.Field(&capability.DocumentationLabels, validation.By(validateDocumentationLabels)),
		validation.Field(&capability.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&capability.LastUpdate, validation.When(capability.LastUpdate != nil, validation.By(isValidDate))),
		validation.Field(&capability.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(common.SemVerRegex))),
	)
}

// fields with validation errors will lead to persisting of the Capability resource
func validateCapabilityInputWithSuppressedErrors(capability *model.CapabilityInput, capabilitiesFromDB map[string]*model.Capability, capabilityHashes map[string]uint64) error {
	return validation.ValidateStruct(capability,
		validation.Field(&capability.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateCapabilityVersionInput(value, *capability, capabilitiesFromDB, capabilityHashes)
		})))
}

func validateIntegrationDependencyInput(integrationDependency *model.IntegrationDependencyInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(integrationDependency,
		validation.Field(&integrationDependency.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(common.IntegrationDependencyOrdIDRegex))),
		validation.Field(&integrationDependency.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&integrationDependency.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&integrationDependency.Title, titleRules(docPolicyLevel, nil)...),
		validation.Field(&integrationDependency.ShortDescription, optionalShortDescriptionRules(docPolicyLevel, nil, integrationDependency.Title)...),
		validation.Field(&integrationDependency.Description, optionalDescriptionRules(docPolicyLevel, integrationDependency.ShortDescription)...),
		validation.Field(&integrationDependency.OrdPackageID, validation.Required, validation.Length(MinOrdPackageIDLength, MaxOrdPackageIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&integrationDependency.LastUpdate, validation.When(integrationDependency.LastUpdate != nil, validation.By(isValidDate))),
		validation.Field(&integrationDependency.Visibility, validation.Required, validation.In(IntegrationDependencyVisibilityPublic, IntegrationDependencyVisibilityInternal, IntegrationDependencyVisibilityPrivate)),
		validation.Field(&integrationDependency.ReleaseStatus, validation.Required, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&integrationDependency.SunsetDate, validation.When(*integrationDependency.ReleaseStatus == common.ReleaseStatusDeprecated, validation.Required), validation.When(integrationDependency.SunsetDate != nil, validation.By(isValidDate))),
		validation.Field(&integrationDependency.Successors, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(IntegrationDependencySuccessorsRegex))
		})),
		validation.Field(&integrationDependency.Mandatory, validation.By(func(value interface{}) error {
			return common.ValidateFieldMandatory(value, IntegrationDependencyMsg)
		})),
		validation.Field(&integrationDependency.Aspects),
		validation.Field(&integrationDependency.RelatedIntegrationDependencies, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(common.IntegrationDependencyOrdIDRegex))
		})),
		validation.Field(&integrationDependency.Links, validation.By(validateORDLinks)),
		validation.Field(&integrationDependency.Tags, tagsRules()...),
		validation.Field(&integrationDependency.Labels, validation.By(validateORDLabels)),
		validation.Field(&integrationDependency.DocumentationLabels, validation.By(validateDocumentationLabels)))
}

// fields with validation errors will lead to persisting of the IntegrationDependency resource
func validateIntegrationDependencyInputWithSuppressedErrors(integrationDependency *model.IntegrationDependencyInput, integrationDependenciesFromDB map[string]*model.IntegrationDependency, integrationDependencyHashes map[string]uint64) error {
	return validation.ValidateStruct(integrationDependency,
		validation.Field(&integrationDependency.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateIntegrationDependencyVersionInput(value, *integrationDependency, integrationDependenciesFromDB, integrationDependencyHashes)
		})))
}

func validateDataProductInput(dataProduct *model.DataProductInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(dataProduct,
		validation.Field(&dataProduct.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(DataProductOrdIDRegex))),
		validation.Field(&dataProduct.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&dataProduct.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&dataProduct.Title, titleRules(docPolicyLevel, dataProduct.PolicyLevel)...),
		validation.Field(&dataProduct.ShortDescription, optionalShortDescriptionRules(docPolicyLevel, dataProduct.PolicyLevel, dataProduct.Title)...),
		validation.Field(&dataProduct.Description, optionalDescriptionRules(docPolicyLevel, dataProduct.ShortDescription)...),
		validation.Field(&dataProduct.OrdPackageID, validation.Required, validation.Length(MinOrdPackageIDLength, MaxOrdPackageIDLength), validation.Match(regexp.MustCompile(common.PackageOrdIDRegex))),
		validation.Field(&dataProduct.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(common.SemVerRegex))),
		validation.Field(&dataProduct.LastUpdate, validation.When(dataProduct.LastUpdate != nil, validation.By(isValidDate))),
		validation.Field(&dataProduct.Visibility, validation.Required, validation.In(DataProductVisibilityPublic, DataProductVisibilityInternal, DataProductVisibilityPrivate)),
		validation.Field(&dataProduct.ReleaseStatus, validation.Required, validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated)),
		validation.Field(&dataProduct.DeprecationDate, validation.NilOrNotEmpty, validation.When(*dataProduct.ReleaseStatus == common.ReleaseStatusDeprecated, validation.Required), validation.When(dataProduct.DeprecationDate != nil, validation.By(isValidDate))),
		validation.Field(&dataProduct.SunsetDate, validation.NilOrNotEmpty, validation.When(*dataProduct.ReleaseStatus == common.ReleaseStatusDeprecated, validation.Required), validation.When(dataProduct.SunsetDate != nil, validation.By(isValidDate))),
		validation.Field(&dataProduct.Successors, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(DataProductSuccessorsRegex))
		})),
		validation.Field(&dataProduct.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&dataProduct.Type, validation.Required, validation.In(DataProductTypeBase, DataProductTypeDerived)),
		validation.Field(&dataProduct.Category, validation.Required, validation.In(DataProductCategoryBusinessObject, DataProductCategoryAnalytical, DataProductCategoryOther)),
		validation.Field(&dataProduct.EntityTypes, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(EntityTypeOrdIDRegex))
		})),
		validation.Field(&dataProduct.InputPorts, validation.By(validateDataProductInputPorts)),
		validation.Field(&dataProduct.OutputPorts, validation.Required, validation.By(validateDataProductOutputPorts)),
		validation.Field(&dataProduct.Responsible, validation.Required, validation.Length(MinResponsibleLength, MaxResponsibleLength), validation.Match(regexp.MustCompile(ResponsibleRegex))),
		validation.Field(&dataProduct.DataProductLinks, validation.By(func(value interface{}) error {
			return validateResourceLinks(value, dataProductResourceLinkTypes)
		})),
		validation.Field(&dataProduct.Links, validation.By(validateORDLinks)),
		validation.Field(&dataProduct.Industry, industryRules(docPolicyLevel, dataProduct.PolicyLevel)...),
		validation.Field(&dataProduct.LineOfBusiness, lineOfBusinessRules(docPolicyLevel, dataProduct.PolicyLevel)...),
		validation.Field(&dataProduct.Tags, tagsRules()...),
		validation.Field(&dataProduct.Labels, validation.By(validateORDLabels)),
		validation.Field(&dataProduct.DocumentationLabels, validation.By(validateDocumentationLabels)),
		validation.Field(&dataProduct.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelCustom, PolicyLevelNone), validation.When(dataProduct.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&dataProduct.CustomPolicyLevel, validation.When(dataProduct.PolicyLevel != nil && *dataProduct.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
	)
}

// fields with validation errors will lead to persisting of the DataProduct resource
func validateDataProductInputWithSuppressedErrors(dataProduct *model.DataProductInput, dataProductsFromDB map[string]*model.DataProduct, dataProductHashes map[string]uint64) error {
	return validation.ValidateStruct(dataProduct,
		validation.Field(&dataProduct.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateDataProductVersionInput(value, *dataProduct, dataProductsFromDB, dataProductHashes)
		})))
}

func validateProductInput(product *model.ProductInput, docPolicyLevel *string) error {
	productOrdIDNamespace := strings.Split(product.OrdID, ":")[0]

	return validation.ValidateStruct(product,
		validation.Field(&product.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(ProductOrdIDRegex))),
		validation.Field(&product.Title, titleRules(docPolicyLevel, nil)...),
		validation.Field(&product.ShortDescription, shortDescriptionRules(docPolicyLevel, nil, product.Title)...),
		validation.Field(&product.Description, optionalDescriptionRules(docPolicyLevel, &product.ShortDescription)...),
		validation.Field(&product.Vendor, validation.Required,
			validation.Match(regexp.MustCompile(VendorOrdIDRegex)),
			validation.When(regexp.MustCompile(SAPProductOrdIDNamespaceRegex).MatchString(productOrdIDNamespace), validation.In(SapVendor)).Else(validation.NotIn(SapVendor)),
			validation.Length(1, 256),
		),
		validation.Field(&product.Parent, validation.When(product.Parent != nil, validation.Match(regexp.MustCompile(ProductOrdIDRegex)))),
		validation.Field(&product.CorrelationIDs, correlationIdsRules()...),
		validation.Field(&product.Labels, validation.By(validateORDLabels)),
		validation.Field(&product.Tags, tagsRules()...),
		validation.Field(&product.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func validateVendorInput(vendor *model.VendorInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(vendor,
		validation.Field(&vendor.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(VendorOrdIDRegex))),
		validation.Field(&vendor.Title, titleRules(docPolicyLevel, nil)...),
		validation.Field(&vendor.Labels, validation.By(validateORDLabels)),
		validation.Field(&vendor.Partners, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(VendorPartnersRegex))
		})),
		validation.Field(&vendor.Tags, tagsRules()...),
		validation.Field(&vendor.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func validateTombstoneInput(tombstone *model.TombstoneInput) error {
	return validation.ValidateStruct(tombstone,
		validation.Field(&tombstone.OrdID, validation.Required, validation.Length(MinOrdIDLength, MaxOrdIDLength), validation.Match(regexp.MustCompile(TombstoneOrdIDRegex))),
		validation.Field(&tombstone.RemovalDate, validation.Required, validation.By(isValidDate)),
		validation.Field(&tombstone.Description, validation.NilOrNotEmpty, validation.Length(MinDescriptionLength, MaxDescriptionLength)))
}

func validateORDLabels(val interface{}) error {
	return validateLabels(val, LabelsKeyRegex)
}

func validateDocumentationLabels(val interface{}) error {
	return validateLabels(val, NoNewLineRegex)
}

func validateLabels(val interface{}, regex string) error {
	if val == nil {
		return nil
	}

	labels, ok := val.(json.RawMessage)
	if !ok {
		return errors.New("labels should be json")
	}

	if len(labels) == 0 {
		return nil
	}

	if !gjson.ValidBytes(labels) {
		return errors.New("labels should be valid json")
	}

	parsedLabels := gjson.ParseBytes(labels)
	if !parsedLabels.IsObject() {
		return errors.New("labels should be json object")
	}

	var err error
	parsedLabels.ForEach(func(key, value gjson.Result) bool {
		if err = validation.Validate(key.String(), validation.Match(regexp.MustCompile(regex))); err != nil {
			return false
		}
		if !value.IsArray() {
			err = errors.New("label value should be array")
			return false
		}
		for _, el := range value.Array() {
			if el.Type != gjson.String {
				err = errors.New("label value should be array of strings")
				return false
			}
		}
		return true
	})
	return err
}

func validateEntryPoints(val interface{}) error {
	if val == nil {
		return nil
	}

	entryPoints, ok := val.(json.RawMessage)
	if !ok {
		return errors.New("entryPoints should be json")
	}

	if len(entryPoints) == 0 {
		return nil
	}

	if !gjson.ValidBytes(entryPoints) {
		return errors.New("entryPoints should be valid json")
	}

	parsedArr := gjson.ParseBytes(entryPoints)
	if !parsedArr.IsArray() {
		return errors.New("should be json array")
	}

	if len(parsedArr.Array()) == 0 {
		return errors.New("entryPoints should not be empty if present")
	}

	if areThereEntryPointDuplicates(parsedArr.Array()) {
		return errors.New("entryPoints should not contain duplicates")
	}

	for _, el := range parsedArr.Array() {
		if el.Type != gjson.String {
			return errors.New("should be array of strings")
		}

		err := validation.Validate(el.String(), is.RequestURI)
		if err != nil {
			return errors.New("entryPoint should be a valid URI")
		}
	}
	return nil
}

func validateLinks(arr interface{}) error {
	if arr == nil {
		return nil
	}

	links, ok := arr.(json.RawMessage)
	if !ok {
		return errors.New("links should be json")
	}

	if len(links) == 0 {
		return nil
	}

	if !gjson.ValidBytes(links) {
		return errors.New("links should be valid json")
	}

	parsedArr := gjson.ParseBytes(links)
	if !parsedArr.IsArray() {
		return errors.New("should be json array")
	}

	if len(parsedArr.Array()) == 0 {
		return nil
	}

	for _, el := range parsedArr.Array() {
		if el.Type != gjson.JSON {
			return errors.New("should be array of json objects")
		}
	}

	if areThereLinkTitleDuplicates(parsedArr.Array()) {
		return errors.New("links should not contain duplicates")
	}

	return nil
}

func validateORDChangeLogEntries(value interface{}) error {
	return common.ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"version": {
			validation.Required,
			validation.Match(regexp.MustCompile(common.SemVerRegex)),
		},
		"releaseStatus": {
			validation.Required,
			validation.In(common.ReleaseStatusBeta, common.ReleaseStatusActive, common.ReleaseStatusDeprecated),
		},
		"date": {
			validation.Required,
			validation.By(isValidDate),
		},
		"description": {
			validation.NilOrNotEmpty,
			validation.Length(MinDescriptionLength, MaxDescriptionLength),
		},
		"url": {
			is.RequestURI,
		},
	})
}

func validateORDLinks(value interface{}) error {
	elementFieldRules := map[string][]validation.Rule{
		"title": {
			validation.Length(MinTitleLength, MaxTitleLength),
			validation.Required,
		},
		"url": {
			validation.Required,
			is.RequestURI,
		},
		"description": {
			validation.NilOrNotEmpty,
			validation.Length(MinDescriptionLength, MaxDescriptionLength),
		},
	}
	if err := common.ValidateJSONArrayOfObjects(value, elementFieldRules); err != nil {
		return err
	}
	if err := validateLinks(value); err != nil {
		return err
	}
	return nil
}

func validatePackageLinks(value interface{}) error {
	return common.ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"type": {
			validation.Required,
			validation.In("terms-of-service", "license", "client-registration", "payment", "sandbox", "service-level-agreement", "support", "custom"),
		},
		"url": {
			validation.Required,
			is.RequestURI,
		},
	}, func(el gjson.Result) error {
		if el.Get("customType").Exists() {
			if el.Get("type").String() != custom {
				return errors.New("if customType is provided, type should be set to 'custom'")
			} else {
				return validation.Validate(el.Get("customType").String(), validation.Match(regexp.MustCompile(CustomImplementationStandardRegex)))
			}
		}
		return nil
	})
}

func validateResourceLinks(value interface{}, resourceTypes []interface{}) error {
	return common.ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"type": {
			validation.Required,
			validation.In(resourceTypes...),
		},
		"url": {
			validation.Required,
			is.RequestURI,
		},
	}, func(el gjson.Result) error {
		if el.Get("customType").Exists() {
			if el.Get("type").String() != custom {
				return errors.New("if customType is provided, type should be set to 'custom'")
			} else {
				return validation.Validate(el.Get("customType").String(), validation.Length(MinResourceLinkCustomTypeLength, MaxResourceLinkCustomTypeLength), validation.Match(regexp.MustCompile(CustomImplementationStandardRegex)))
			}
		}
		return nil
	})
}

func validateDataProductInputPorts(value interface{}) error {
	return common.ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"ordId": {
			validation.Required,
			validation.Length(MinOrdIDLength, MaxOrdIDLength),
			validation.Match(regexp.MustCompile(common.IntegrationDependencyOrdIDRegex)),
		},
	})
}

func validateDataProductOutputPorts(value interface{}) error {
	return common.ValidateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"ordId": {
			validation.Required,
			validation.Length(MinOrdIDLength, MaxOrdIDLength),
			validation.Match(regexp.MustCompile(DataProductOutputPortsRegex)),
		},
	})
}

func validateAPIResourceDefinitions(value interface{}, api model.APIDefinitionInput, docPolicyLevel *string) error {
	if value == nil {
		return nil
	}

	policyLevel := str.PtrStrToStr(docPolicyLevel)
	apiVisibility := str.PtrStrToStr(api.Visibility)
	apiProtocol := str.PtrStrToStr(api.APIProtocol)
	resourceDefinitions := api.ResourceDefinitions

	isResourceDefinitionMandatory := !(policyLevel == PolicyLevelSap && apiVisibility == APIVisibilityPrivate)
	if len(resourceDefinitions) == 0 && isResourceDefinitionMandatory {
		return errors.New("when api resource visibility is public or internal, resource definitions must be provided")
	}

	if len(resourceDefinitions) == 0 && !isResourceDefinitionMandatory {
		return nil
	}

	resourceDefinitionTypes := make(map[model.APISpecType]bool)

	for _, rd := range resourceDefinitions {
		resourceDefinitionType := rd.Type
		resourceDefinitionTypes[resourceDefinitionType] = true
	}

	isPolicyCore := policyLevel == PolicyLevelSap
	wsdlTypeExists := resourceDefinitionTypes[model.APISpecTypeWsdlV1] || resourceDefinitionTypes[model.APISpecTypeWsdlV2]
	if isPolicyCore && (apiProtocol == APIProtocolSoapInbound || apiProtocol == APIProtocolSoapOutbound) && !wsdlTypeExists {
		return errors.New("for APIResources of policyLevel='sap' and with apiProtocol='soap-inbound' or 'soap-outbound' it is mandatory to provide either WSDL V2 or WSDL V1 definitions")
	}

	edmxTypeExists := resourceDefinitionTypes[model.APISpecTypeEDMX]
	openAPITypeExists := resourceDefinitionTypes[model.APISpecTypeOpenAPIV2] || resourceDefinitionTypes[model.APISpecTypeOpenAPIV3]
	if isPolicyCore && (apiProtocol == APIProtocolODataV2 || apiProtocol == APIProtocolODataV4) && !(edmxTypeExists && openAPITypeExists) {
		return errors.New("for APIResources of policyLevel='sap' and with apiProtocol='odata-v2' or 'odata-v4' it is mandatory to not only provide edmx definitions, but also OpenAPI definitions")
	}

	if isPolicyCore && apiProtocol == APIProtocolRest && !openAPITypeExists {
		return errors.New("for APIResources of policyLevel='sap' and with apiProtocol='rest' it is mandatory to provide either OpenAPI 3 or OpenAPI 2 definitions")
	}

	rfcMetadataTypeExists := resourceDefinitionTypes[model.APISpecTypeRfcMetadata]
	if isPolicyCore && apiProtocol == APIProtocolSapRfc && !rfcMetadataTypeExists {
		return errors.New("for APIResources of policyLevel='sap' and with apiProtocol='sap-rfc' it is mandatory to provide SAP RFC definitions")
	}

	graphqlSDLTypeExists := resourceDefinitionTypes[model.APISpecTypeGraphqlSDL]
	if isPolicyCore && apiProtocol == APIProtocolGraphql && !graphqlSDLTypeExists {
		return errors.New("for APIResources of policyLevel='sap' and with apiProtocol='graphql' it is mandatory to provide Graphql definitions")
	}

	if apiProtocol == APIProtocolWebsocket && (api.ImplementationStandard == nil || !resourceDefinitionTypes[model.APISpecTypeCustom]) {
		return errors.New("for APIResources with apiProtocol='websocket' it is mandatory to provide implementationStandard definition and type to be set to custom")
	}

	if apiProtocol == APIProtocolSAPSQLAPIV1 && !(resourceDefinitionTypes[model.APISpecTypeCustom] || resourceDefinitionTypes[model.APISpecTypeSQLAPIDefinitionV1]) {
		return errors.New("for APIResources with apiProtocol='sap-sql-api-v1' it is mandatory type to be set either to sap-sql-api-definition-v1 or custom")
	}

	return nil
}

func validateEventResourceDefinition(value interface{}, event model.EventDefinitionInput, docPolicyLevel *string) error {
	if value == nil {
		return nil
	}

	policyLevel := str.PtrStrToStr(docPolicyLevel)
	eventVisibility := str.PtrStrToStr(event.Visibility)

	if policyLevel == PolicyLevelSap && eventVisibility == EventVisibilityPrivate {
		return nil
	}

	eventResourceDef, ok := value.([]*model.EventResourceDefinition)
	if !ok {
		return errors.New("error while casting to EventResourceDefinition")
	}

	if len(eventResourceDef) == 0 {
		return errors.New("when event resource visibility is public or internal, resource definitions must be provided")
	}

	return nil
}

func validateCapabilityDefinitions(value interface{}, capability model.CapabilityInput) error {
	if value == nil {
		return nil
	}

	capabilityDefinitions := capability.CapabilityDefinitions

	capabilityDefinitionTypes := make(map[model.CapabilitySpecType]bool)

	for _, cd := range capabilityDefinitions {
		capabilityDefinitionType := cd.Type
		capabilityDefinitionTypes[capabilityDefinitionType] = true
	}

	mdiCapabilitySpecTypeExists := capabilityDefinitionTypes[model.CapabilitySpecTypeMDICapabilityDefinitionV1]
	if capability.Type != CapabilityTypeMDICapabilityV1 && mdiCapabilitySpecTypeExists {
		return errors.New("when capability definition type is `sap.mdo:mdi-capability-definition:v1`, capability type should be `sap.mdo:mdi-capability:v1`")
	}

	return nil
}

func validatePackageVersionInput(value interface{}, pkg model.PackageInput, pkgsFromDB map[string]*model.Package, resourceHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(pkgsFromDB) == 0 {
		return nil
	}

	pkgFromDB, ok := pkgsFromDB[pkg.OrdID]
	if !ok || isResourceHashMissing(pkgFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(pkgFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(resourceHashes[pkg.OrdID], 10)

	return checkHashEquality(pkgFromDB.Version, pkg.Version, hashDB, hashDoc)
}

func validateBundleVersionInput(value interface{}, bndl model.BundleCreateInput, bndlsFromDB map[string]*model.Bundle, resourceHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(bndlsFromDB) == 0 {
		return nil
	}

	bndlOrdID := str.PtrStrToStr(bndl.OrdID)
	bndlFromDB, ok := bndlsFromDB[bndlOrdID]
	if !ok || isResourceHashMissing(bndlFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(bndlFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(resourceHashes[str.PtrStrToStr(bndl.OrdID)], 10)

	if bndlFromDB.Version != nil && bndl.Version != nil {
		return checkHashEquality(*bndlFromDB.Version, *bndl.Version, hashDB, hashDoc)
	}
	if bndlFromDB.Version != nil && bndl.Version == nil {
		return errors.New("bundle version is present in the DB, but is missing from the document")
	}
	return nil
}

func validateEventDefinitionVersionInput(value interface{}, event model.EventDefinitionInput, eventsFromDB map[string]*model.EventDefinition, eventHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(eventsFromDB) == 0 {
		return nil
	}

	eventFromDB, ok := eventsFromDB[str.PtrStrToStr(event.OrdID)]
	if !ok || isResourceHashMissing(eventFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(eventFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(eventHashes[str.PtrStrToStr(event.OrdID)], 10)

	return checkHashEquality(eventFromDB.Version.Value, event.VersionInput.Value, hashDB, hashDoc)
}

func validateEntityTypeVersionInput(value interface{}, entityType model.EntityTypeInput, entityTypesFromDB map[string]*model.EntityType, entityTypeHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(entityTypesFromDB) == 0 {
		return nil
	}

	eventFromDB, ok := entityTypesFromDB[entityType.OrdID]
	if !ok || isResourceHashMissing(eventFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(eventFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(entityTypeHashes[entityType.OrdID], 10)

	return checkHashEquality(eventFromDB.Version.Value, entityType.VersionInput.Value, hashDB, hashDoc)
}

func validateAPIDefinitionVersionInput(value interface{}, api model.APIDefinitionInput, apisFromDB map[string]*model.APIDefinition, apiHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(apisFromDB) == 0 {
		return nil
	}

	apiFromDB, ok := apisFromDB[str.PtrStrToStr(api.OrdID)]
	if !ok || isResourceHashMissing(apiFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(apiFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(apiHashes[str.PtrStrToStr(api.OrdID)], 10)

	return checkHashEquality(apiFromDB.Version.Value, api.VersionInput.Value, hashDB, hashDoc)
}

func validateCapabilityVersionInput(value interface{}, capability model.CapabilityInput, capabilitiesFromDB map[string]*model.Capability, capabilityHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(capabilitiesFromDB) == 0 {
		return nil
	}

	capabilityFromDB, ok := capabilitiesFromDB[str.PtrStrToStr(capability.OrdID)]
	if !ok || isResourceHashMissing(capabilityFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(capabilityFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(capabilityHashes[str.PtrStrToStr(capability.OrdID)], 10)

	return checkHashEquality(capabilityFromDB.Version.Value, capability.VersionInput.Value, hashDB, hashDoc)
}

func validateIntegrationDependencyVersionInput(value interface{}, integrationDependency model.IntegrationDependencyInput, integrationDependenciesFromDB map[string]*model.IntegrationDependency, integrationDependencyHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(integrationDependenciesFromDB) == 0 {
		return nil
	}

	integrationDependencyFromDB, ok := integrationDependenciesFromDB[str.PtrStrToStr(integrationDependency.OrdID)]
	if !ok || isResourceHashMissing(integrationDependencyFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(integrationDependencyFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(integrationDependencyHashes[str.PtrStrToStr(integrationDependency.OrdID)], 10)

	return checkHashEquality(integrationDependencyFromDB.Version.Value, integrationDependency.VersionInput.Value, hashDB, hashDoc)
}

func validateDataProductVersionInput(value interface{}, dataProduct model.DataProductInput, dataProductsFromDB map[string]*model.DataProduct, dataProductHashes map[string]uint64) error {
	if value == nil {
		return nil
	}

	if len(dataProductsFromDB) == 0 {
		return nil
	}

	dataProductFromDB, ok := dataProductsFromDB[str.PtrStrToStr(dataProduct.OrdID)]
	if !ok || isResourceHashMissing(dataProductFromDB.ResourceHash) {
		return nil
	}

	hashDB := str.PtrStrToStr(dataProductFromDB.ResourceHash)
	hashDoc := strconv.FormatUint(dataProductHashes[str.PtrStrToStr(dataProduct.OrdID)], 10)

	return checkHashEquality(dataProductFromDB.Version.Value, dataProduct.VersionInput.Value, hashDB, hashDoc)
}

func normalizeAPIDefinition(api *model.APIDefinitionInput) (model.APIDefinitionInput, error) {
	bytes, err := json.Marshal(api)
	if err != nil {
		return model.APIDefinitionInput{}, errors.Wrapf(err, "error while marshalling api definition with ID %s", str.PtrStrToStr(api.OrdID))
	}

	var normalizedAPIDefinition model.APIDefinitionInput
	if err := json.Unmarshal(bytes, &normalizedAPIDefinition); err != nil {
		return model.APIDefinitionInput{}, errors.Wrapf(err, "error while unmarshalling api definition with ID %s", str.PtrStrToStr(api.OrdID))
	}

	return normalizedAPIDefinition, nil
}

func normalizeEventDefinition(event *model.EventDefinitionInput) (model.EventDefinitionInput, error) {
	bytes, err := json.Marshal(event)
	if err != nil {
		return model.EventDefinitionInput{}, errors.Wrapf(err, "error while marshalling event definition with ID %s", str.PtrStrToStr(event.OrdID))
	}

	var normalizedEventDefinition model.EventDefinitionInput
	if err := json.Unmarshal(bytes, &normalizedEventDefinition); err != nil {
		return model.EventDefinitionInput{}, errors.Wrapf(err, "error while unmarshalling event definition with ID %s", str.PtrStrToStr(event.OrdID))
	}

	return normalizedEventDefinition, nil
}

func normalizeEntityType(entityType *model.EntityTypeInput) (model.EntityTypeInput, error) {
	bytes, err := json.Marshal(entityType)
	if err != nil {
		return model.EntityTypeInput{}, errors.Wrapf(err, "error while marshalling entity type with ID %s", entityType.OrdID)
	}

	var normalizedEntityType model.EntityTypeInput
	if err := json.Unmarshal(bytes, &normalizedEntityType); err != nil {
		return model.EntityTypeInput{}, errors.Wrapf(err, "error while unmarshalling entity type with ID %s", entityType.OrdID)
	}

	return normalizedEntityType, nil
}

func normalizeCapability(capability *model.CapabilityInput) (model.CapabilityInput, error) {
	bytes, err := json.Marshal(capability)
	if err != nil {
		return model.CapabilityInput{}, errors.Wrapf(err, "error while marshalling capability with ID %s", str.PtrStrToStr(capability.OrdID))
	}

	var normalizedCapability model.CapabilityInput
	if err := json.Unmarshal(bytes, &normalizedCapability); err != nil {
		return model.CapabilityInput{}, errors.Wrapf(err, "error while unmarshalling capability with ID %s", str.PtrStrToStr(capability.OrdID))
	}

	return normalizedCapability, nil
}

func normalizeIntegrationDependency(integrationDependency *model.IntegrationDependencyInput) (model.IntegrationDependencyInput, error) {
	bytes, err := json.Marshal(integrationDependency)
	if err != nil {
		return model.IntegrationDependencyInput{}, errors.Wrapf(err, "error while marshalling integration dependency with ID %s", str.PtrStrToStr(integrationDependency.OrdID))
	}

	var normalizedIntegrationDependency model.IntegrationDependencyInput
	if err := json.Unmarshal(bytes, &normalizedIntegrationDependency); err != nil {
		return model.IntegrationDependencyInput{}, errors.Wrapf(err, "error while unmarshalling integration dependency with ID %s", str.PtrStrToStr(integrationDependency.OrdID))
	}

	return normalizedIntegrationDependency, nil
}

func normalizePackage(pkg *model.PackageInput) (model.PackageInput, error) {
	bytes, err := json.Marshal(pkg)
	if err != nil {
		return model.PackageInput{}, errors.Wrapf(err, "error while marshalling package definition with ID %s", pkg.OrdID)
	}

	var normalizedPkgDefinition model.PackageInput
	if err := json.Unmarshal(bytes, &normalizedPkgDefinition); err != nil {
		return model.PackageInput{}, errors.Wrapf(err, "error while unmarshalling package definition with ID %s", pkg.OrdID)
	}

	return normalizedPkgDefinition, nil
}

func normalizeBundle(bndl *model.BundleCreateInput) (model.BundleCreateInput, error) {
	bytes, err := json.Marshal(bndl)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrapf(err, "error while marshalling bundle definition with ID %v", bndl.OrdID)
	}

	var normalizedBndlDefinition model.BundleCreateInput
	if err := json.Unmarshal(bytes, &normalizedBndlDefinition); err != nil {
		return model.BundleCreateInput{}, errors.Wrapf(err, "error while unmarshalling bundle definition with ID %v", bndl.OrdID)
	}

	return normalizedBndlDefinition, nil
}

func isResourceHashMissing(hash *string) bool {
	hashStr := str.PtrStrToStr(hash)
	return hashStr == ""
}

func validateWhenPolicyLevelIsSAP(docPolicyLevel *string, resourcePolicyLevel *string, validationFunc func() error) error {
	policyLevel := str.PtrStrToStr(docPolicyLevel)
	if resourcePolicyLevel != nil {
		policyLevel = str.PtrStrToStr(resourcePolicyLevel)
	}

	if policyLevel != PolicyLevelSap {
		return nil
	}

	return validationFunc()
}

func validateJSONArrayOfStringsContainsInMap(arr interface{}, validValues map[string]bool) error {
	if arr == nil {
		return nil
	}

	jsonArr, ok := arr.(json.RawMessage)
	if !ok {
		return errors.New("should be json")
	}

	if len(jsonArr) == 0 {
		return nil
	}

	if !gjson.ValidBytes(jsonArr) {
		return errors.New("should be valid json")
	}

	parsedArr := gjson.ParseBytes(jsonArr)
	if !parsedArr.IsArray() {
		return errors.New("should be json array")
	}

	for _, el := range parsedArr.Array() {
		if el.Type != gjson.String {
			return errors.New("should be array of strings")
		}

		exists, ok := validValues[el.String()]

		if !exists || !ok {
			return errors.New("array element is not in the list of valid values")
		}
	}

	return nil
}

func validateJSONArrayOfStringsMatchPattern(arr interface{}, regexPattern *regexp.Regexp) error {
	if arr == nil {
		return nil
	}

	jsonArr, ok := arr.(json.RawMessage)
	if !ok {
		return errors.New("should be json")
	}

	if len(jsonArr) == 0 {
		return nil
	}

	if !gjson.ValidBytes(jsonArr) {
		return errors.New("should be valid json")
	}
	parsedArr := gjson.ParseBytes(jsonArr)

	if !parsedArr.IsArray() {
		return errors.New("should be json array")
	}

	if len(parsedArr.Array()) == 0 {
		return nil
	}

	for _, el := range parsedArr.Array() {
		if el.Type != gjson.String {
			return errors.New("should be array of strings")
		}
		if !regexPattern.MatchString(el.String()) {
			return errors.Errorf("elements should match %q", regexPattern.String())
		}
	}
	return nil
}

func validateJSONObjects(obj interface{}, elementFieldRules map[string][]validation.Rule, crossFieldRules ...func(gjson.Result) error) error {
	if obj == nil {
		return nil
	}

	jsonObj, ok := obj.(json.RawMessage)
	if !ok {
		return errors.New("should be json")
	}

	if len(jsonObj) == 0 {
		return nil
	}

	if !gjson.ValidBytes(jsonObj) {
		return errors.New("should be valid json")
	}

	parsedObj := gjson.ParseBytes(jsonObj)
	if !parsedObj.IsObject() {
		return errors.New("should be json object")
	}

	for field, rules := range elementFieldRules {
		if err := validation.Validate(parsedObj.Get(field).Value(), rules...); err != nil {
			return errors.Wrapf(err, "error validating field %s", field)
		}
		for _, f := range crossFieldRules {
			if err := f(parsedObj); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateCustomType(credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) func(el gjson.Result) error {
	return func(el gjson.Result) error {
		if el.Get("customType").Exists() && el.Get("type").String() != custom {
			return errors.New("if customType is provided, type should be set to 'custom'")
		}

		customType := el.Get("customType").String()
		if _, ok := credentialExchangeStrategyTenantMappings[customType]; strings.Contains(customType, TenantMappingCustomTypeIdentifier) && !ok {
			return errors.New("credential exchange strategy's tenant mapping customType is not valid")
		}

		return validation.Validate(customType, validation.Match(regexp.MustCompile(CustomTypeCredentialExchangeStrategyRegex)))
	}
}

func validateCustomDescription(el gjson.Result) error {
	if el.Get("customDescription").Exists() && el.Get("type").String() != custom {
		return errors.New("if customDescription is provided, type should be set to 'custom'")
	}
	if el.Get("customDescription").Exists() && el.Get("type").String() == custom && (len(el.Get("customDescription").String()) < MinDescriptionLength || len(el.Get("customDescription").String()) > MaxDescriptionLength) {
		return errors.New(fmt.Sprintf("if customDescription is provided and type is 'custom', then the accepted length of customDescription is between %d - %d characters", MinDescriptionLength, MaxDescriptionLength))
	}
	return nil
}

func validateEventPartOfConsumptionBundles(value interface{}, regexPattern *regexp.Regexp) error {
	bundleReferences, ok := value.([]*model.ConsumptionBundleReference)
	if !ok {
		return errors.New("error while casting to ConsumptionBundleReference")
	}

	if bundleReferences != nil && len(bundleReferences) == 0 {
		return errors.New("bundleReference should not be empty if present")
	}

	bundleIDsPerEvent := make(map[string]bool)
	for _, br := range bundleReferences {
		if br.BundleOrdID == "" {
			return errors.New("bundleReference ordId is mandatory field")
		}

		if !regexPattern.MatchString(br.BundleOrdID) {
			return errors.Errorf("ordId should match %q", regexPattern.String())
		}

		if isPresent := bundleIDsPerEvent[br.BundleOrdID]; !isPresent {
			bundleIDsPerEvent[br.BundleOrdID] = true
		} else {
			return errors.Errorf("event can not reference the same bundle with ordId %q more than once", br.BundleOrdID)
		}

		if br.DefaultTargetURL != "" {
			return errors.New("events are not supposed to have defaultEntryPoint")
		}
	}
	return nil
}

func validateEntityTypeMappings(value interface{}) error {
	entityTypeMappings, ok := value.([]*model.EntityTypeMappingInput)
	if !ok {
		return errors.New("error while casting to EntityTypeMapping")
	}

	if entityTypeMappings != nil && len(entityTypeMappings) == 0 {
		return errors.New("entityTypeMappings should not be empty if present")
	}
	for _, entityTypeMapping := range entityTypeMappings {

		// Validate EntityTypeTargets
		var entityTypeTargets []*model.EntityTypeTarget
		if err := json.Unmarshal(entityTypeMapping.EntityTypeTargets, &entityTypeTargets); err != nil {
			return errors.New("error while unmarshalling EntityTypeTarget for EntityTypeMapping")
		}
		if len(entityTypeTargets) == 0 {
			return errors.New("entity type target should not be blank")
		}
		for _, entityTypeTarget := range entityTypeTargets {
			err := validateEntityTypeTarget(entityTypeTarget)
			if err != nil {
				return errors.Wrap(err, "error while validating EntityTypeTarget")
			}
		}

		// apiModelSelectors is optional field
		if entityTypeMapping.APIModelSelectors == nil {
			continue
		}

		// Validate APIModelSelectors
		var apiModelSelectors []*model.APIModelSelector
		if err := json.Unmarshal(entityTypeMapping.APIModelSelectors, &apiModelSelectors); err != nil {
			return errors.New("error while unmarshalling APIModelSelectors for EntityTypeMapping")
		}
		for _, apiModelSelector := range apiModelSelectors {
			err := validateAPIModelSelector(apiModelSelector)
			if err != nil {
				return errors.Wrap(err, "error while validating APIModelSelector")
			}
		}
	}
	return nil
}

func validateAPIModelSelector(value interface{}) error {
	apiModelSelector, ok := value.(*model.APIModelSelector)
	if !ok {
		return errors.New("error while casting to APIModelSelector")
	}
	return validation.ValidateStruct(apiModelSelector,
		validation.Field(&apiModelSelector.Type, validation.Required, validation.In(APIModelSelectorTypeODATA, APIModelSelectorTypeJSONPointer)),
		validation.Field(&apiModelSelector.EntitySetName,
			validation.When(apiModelSelector.Type == APIModelSelectorTypeODATA, validation.Required),
			validation.When(apiModelSelector.Type == APIModelSelectorTypeJSONPointer, validation.Nil),
		),
		validation.Field(&apiModelSelector.JSONPointer,
			validation.When(apiModelSelector.Type == APIModelSelectorTypeJSONPointer, validation.Required),
			validation.When(apiModelSelector.Type == APIModelSelectorTypeODATA, validation.Nil),
		),
	)
}

func validateEntityTypeTarget(value interface{}) error {
	entityTypeTarget, ok := value.(*model.EntityTypeTarget)
	if !ok {
		return errors.New("error while casting to EntityTypeTarget")
	}
	return validation.ValidateStruct(entityTypeTarget,
		validation.Field(&entityTypeTarget.OrdID,
			validation.When(entityTypeTarget.CorrelationID != nil, validation.Nil),
			validation.When(entityTypeTarget.CorrelationID == nil, validation.Required),
			validation.Length(MinOrdIDLength, MaxOrdIDLength),
			validation.Match(regexp.MustCompile(EntityTypeOrdIDRegex)),
		),
		validation.Field(&entityTypeTarget.CorrelationID,
			validation.When(entityTypeTarget.OrdID != nil, validation.Nil),
			validation.When(entityTypeTarget.OrdID == nil, validation.Required),
			validation.Length(MinCorrelationIDLength, MaxCorrelationIDLength),
			validation.Match(regexp.MustCompile(CorrelationIDsRegex)),
		),
	)
}

func validateAPIPartOfConsumptionBundles(value interface{}, targetURLs json.RawMessage, regexPattern *regexp.Regexp) error {
	bundleReferences, ok := value.([]*model.ConsumptionBundleReference)
	if !ok {
		return errors.New("error while casting to ConsumptionBundleReference")
	}

	if bundleReferences != nil && len(bundleReferences) == 0 {
		return errors.New("bundleReference should not be empty if present")
	}

	bundleIDsPerAPI := make(map[string]bool)
	for _, br := range bundleReferences {
		if br.BundleOrdID == "" {
			return errors.New("bundleReference ordId is mandatory field")
		}

		if !regexPattern.MatchString(br.BundleOrdID) {
			return errors.Errorf("ordId should match %q", regexPattern.String())
		}

		if isPresent := bundleIDsPerAPI[br.BundleOrdID]; !isPresent {
			bundleIDsPerAPI[br.BundleOrdID] = true
		} else {
			return errors.Errorf("api can not reference the same bundle with ordId %q more than once", br.BundleOrdID)
		}

		err := validation.Validate(br.DefaultTargetURL, is.RequestURI)
		if err != nil {
			return errors.New("defaultEntryPoint should be a valid URI")
		}

		lenTargetURLs := len(gjson.ParseBytes(targetURLs).Array())
		if br.DefaultTargetURL != "" && lenTargetURLs <= 1 {
			return errors.New("defaultEntryPoint must only be provided if an API has more than one entry point")
		}

		if br.DefaultTargetURL != "" && lenTargetURLs > 1 {
			if isDefaultTargetURLMissingFromTargetURLs(br.DefaultTargetURL, targetURLs) {
				return errors.New("defaultEntryPoint must be in the list of entryPoints for the given API")
			}
		}
	}

	return nil
}

func validateDefaultConsumptionBundle(value interface{}, partOfConsumptionBundles []*model.ConsumptionBundleReference) error {
	defaultConsumptionBundle, ok := value.(*string)
	if !ok {
		return errors.New(fmt.Sprintf("expected string value for defaultConsumptionBundle, found %T", value))
	}

	if defaultConsumptionBundle == nil {
		return nil
	}

	var isFound bool
	for _, bundleRef := range partOfConsumptionBundles {
		if *defaultConsumptionBundle == bundleRef.BundleOrdID {
			isFound = true
			break
		}
	}

	if !isFound {
		return errors.New("defaultConsumptionBundle must be an existing option in the corresponding partOfConsumptionBundles array")
	}
	return nil
}

func isDefaultTargetURLMissingFromTargetURLs(defaultTargetURL string, targetURLs json.RawMessage) bool {
	for _, targetURL := range gjson.ParseBytes(targetURLs).Array() {
		if targetURL.String() == defaultTargetURL {
			return false
		}
	}
	return true
}

func areThereEntryPointDuplicates(entryPoints []gjson.Result) bool {
	if len(entryPoints) <= 1 {
		return false
	}

	seen := make(map[string]bool)
	for _, val := range entryPoints {
		if seen[val.String()] {
			return true
		}
		seen[val.String()] = true
	}
	return false
}

func areThereLinkTitleDuplicates(links []gjson.Result) bool {
	if len(links) <= 1 {
		return false
	}

	seen := make(map[string]bool)
	for _, val := range links {
		if seen[val.Get("title").String()] {
			return true
		}
		seen[val.Get("title").String()] = true
	}
	return false
}

func isValidDate(d interface{}) error {
	var err error
	date, err := castDate(d)
	if err != nil {
		return err
	}

	if _, err = time.Parse(time.RFC3339, date); err == nil { // RFC3339 -> "2006-01-02T15:04:05Z" or "2006-01-02T15:04:05+07:00"
		return nil
	} else if _, err = time.Parse("2006-01-02", date); err == nil { // RFC3339 date without time extension
		return nil
	} else if _, err = time.Parse("2006-01-02T15:04:05", date); err == nil { // ISO8601 without Z/00+00
		return nil
	} else if _, err = time.Parse("2006-01-02T15:04:05Z0700", date); err == nil { // ISO8601 with skipped ':' in offset (e.g.: 2006-01-02T15:04:05+0700)
		return nil
	}
	return errors.New("invalid date")
}

func castDate(d interface{}) (string, error) {
	datePtr, ok := d.(*string)
	if ok {
		return *datePtr, nil
	}

	date, ok := d.(string)
	if ok {
		return date, nil
	}

	return "", errors.New(fmt.Sprintf("expected string or *string value for date, found %T", d))
}

func notPartOfConsumptionBundles(partOfConsumptionBundles []*model.ConsumptionBundleReference) validation.RuleFunc {
	return func(value interface{}) error {
		if len(partOfConsumptionBundles) > 0 {
			return errors.New("api without entry points can not be part of consumption bundle")
		}
		return nil
	}
}

func validateExtensibleField(value interface{}, policyLevelInput *string, shouldBeRequired bool) error {
	policyLevel := str.PtrStrToStr(policyLevelInput)

	if (policyLevel == PolicyLevelSap) && shouldBeRequired && (value == nil || value.(json.RawMessage) == nil) {
		return errors.Errorf("`extensible` field must be provided when `policyLevel` is `%s`", PolicyLevelSap)
	}

	return validateJSONObjects(value, map[string][]validation.Rule{
		"supported": {
			validation.Required,
			validation.In("no", "manual", "automatic"),
		},
		"description": {},
	}, validateExtensibleInnerFields)
}

func validateExtensibleInnerFields(el gjson.Result) error {
	supportedProperty := el.Get("supported")
	supportedValue, ok := supportedProperty.Value().(string)
	if !ok {
		return errors.New("`supported` value not provided")
	}

	descriptionProperty := el.Get("description")
	descriptionValue, ok := descriptionProperty.Value().(string)
	validLength := len(descriptionValue) >= MinDescriptionLength && len(descriptionValue) <= MaxDescriptionLength

	if supportedProperty.Exists() && (supportedValue == "manual" || supportedValue == "automatic") && (!validLength || !ok) {
		return errors.New(fmt.Sprintf("if supported field is either 'manual' or 'automatic', description should be provided with length of %d - %d characters", MinDescriptionLength, MaxDescriptionLength))
	}
	return nil
}

// HashObject hashes the given object
func HashObject(obj interface{}) (uint64, error) {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return 0, errors.New("failed to hash the given object")
	}

	return hash, nil
}

func checkHashEquality(rdFromDBVersion, rdFromDocVersion, hashFromDB, hashFromDoc string) error {
	rdFromDBVersion = fmt.Sprintf("v%s", rdFromDBVersion)
	rdFromDocVersion = fmt.Sprintf("v%s", rdFromDocVersion)

	areVersionsEqual := semver.Compare(rdFromDocVersion, rdFromDBVersion)
	if areHashesEqual := cmp.Equal(hashFromDB, hashFromDoc); !areHashesEqual && areVersionsEqual <= 0 {
		return errors.New("there is a change in the resource; version value should be incremented")
	}

	return nil
}
