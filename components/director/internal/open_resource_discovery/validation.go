package ord

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processors"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	// SemVerRegex represents the valid structure of the field
	SemVerRegex = "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
	// PackageOrdIDRegex represents the valid structure of the ordID of the Package
	PackageOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(package):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// VendorOrdIDRegex represents the valid structure of the ordID of the Vendor
	VendorOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(vendor):([a-zA-Z0-9._\\-]+):()$"
	// ProductOrdIDRegex represents the valid structure of the ordID of the Product
	ProductOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(product):([a-zA-Z0-9._\\-]+):()$"
	// BundleOrdIDRegex represents the valid structure of the ordID of the ConsumptionBundle
	BundleOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(consumptionBundle):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*)$"
	// TombstoneOrdIDRegex represents the valid structure of the ordID of the Tombstone
	TombstoneOrdIDRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(package|consumptionBundle|product|vendor|apiResource|eventResource):([a-zA-Z0-9._\\-]+):(v0|v[1-9][0-9]*|)$"
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
	// CorrelationIDsRegex represents the valid structure of the field
	CorrelationIDsRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-\\/]+):([a-zA-Z0-9._\\-\\/]+)$"
	// LabelsKeyRegex represents the valid structure of the field
	LabelsKeyRegex = "^[a-zA-Z0-9-_.]*$"
	// NoNewLineRegex represents the valid structure of the field
	NoNewLineRegex = "^[^\\n]*$"
	// CustomImplementationStandardRegex represents the valid structure of the field
	CustomImplementationStandardRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	// VendorPartnersRegex represents the valid structure of the field
	VendorPartnersRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):(vendor):([a-zA-Z0-9._\\-]+):()$"
	// CustomPolicyLevelRegex represents the valid structure of the field
	CustomPolicyLevelRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	// CustomTypeCredentialExchangeStrategyRegex represents the valid structure of the field
	CustomTypeCredentialExchangeStrategyRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	// SAPProductOrdIDNamespaceRegex represents the valid structure of a SAP Product OrdID Namespace part
	SAPProductOrdIDNamespaceRegex = "^(sap)((\\.)([a-z0-9-]+(?:[.][a-z0-9-]+)*))*$"
	// OrdNamespaceRegex represents the valid structure of a Ord Namespace
	OrdNamespaceRegex = "^[a-z0-9]+(?:[.][a-z0-9]+)*$"

	// MinDescriptionLength represents the minimal accepted length of the Description field
	MinDescriptionLength = 1
	// MaxDescriptionLength represents the minimal accepted length of the Description field
	MaxDescriptionLength = 5000
	// MinLocalTenantIDLength represents the minimal accepted length of the LocalID field
	MinLocalTenantIDLength = 1
	// MaxLocalTenantIDLength represents the minimal accepted length of the LocalID field
	MaxLocalTenantIDLength = 255
	// MinSystemVersionTitleLength represents the minimal accepted length of the LocalID field
	MinSystemVersionTitleLength = 1
	// MaxSystemVersionTitleLength represents the minimal accepted length of the LocalID field
	MaxSystemVersionTitleLength = 255
)

const (
	custom string = "custom"
	none   string = "none"

	// PolicyLevelSap is one of the available policy options
	PolicyLevelSap string = "sap:core:v1"
	// PolicyLevelSapPartner is one of the available policy options
	PolicyLevelSapPartner string = "sap:partner:v1"
	// PolicyLevelCustom is one of the available policy options
	PolicyLevelCustom = custom
	// PolicyLevelNone is one of the available policy options
	PolicyLevelNone string = none

	// ReleaseStatusBeta is one of the available release status options
	ReleaseStatusBeta string = "beta"
	// ReleaseStatusActive is one of the available release status options
	ReleaseStatusActive string = "active"
	// ReleaseStatusDeprecated is one of the available release status options
	ReleaseStatusDeprecated string = "deprecated"

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

	// APIVisibilityPublic is one of the available api visibility options
	APIVisibilityPublic string = "public"
	// APIVisibilityPrivate is one of the available api visibility options
	APIVisibilityPrivate string = "private"
	// APIVisibilityInternal is one of the available api visibility options
	APIVisibilityInternal string = "internal"

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
		"mass-extraction": true,
		// "mass-import":     true, // will be added later in spec
	}
)

var shortDescriptionRules = []validation.Rule{
	validation.Required, validation.RuneLength(1, 256), validation.NewStringRule(noNewLines, "short description should not contain line breaks"),
}

var optionalShortDescriptionRules = []validation.Rule{
	validation.NilOrNotEmpty, validation.RuneLength(1, 256), validation.NewStringRule(noNewLines, "short description should not contain line breaks"),
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
		validation.Field(&app.CorrelationIDs, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CorrelationIDsRegex))
		})),
		validation.Field(&app.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&app.ApplicationNamespace, validation.Match(regexp.MustCompile(OrdNamespaceRegex))),
		validation.Field(&app.BaseURL, is.RequestURI, validation.Match(regexp.MustCompile(SystemInstanceBaseURLRegex))),
		validation.Field(&app.OrdLabels, validation.By(validateORDLabels)),
		validation.Field(&app.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
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
		validation.Field(&doc.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelSapPartner, PolicyLevelCustom, PolicyLevelNone), validation.When(doc.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&doc.CustomPolicyLevel, validation.When(doc.PolicyLevel != nil && *doc.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
	)
}

func validatePackageInput(pkg *model.PackageInput, docPolicyLevel *string) error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.OrdID, validation.Required, validation.Match(regexp.MustCompile(PackageOrdIDRegex))),
		validation.Field(&pkg.Title, validation.Length(1, 255), validation.Required),
		validation.Field(&pkg.ShortDescription, shortDescriptionRules...),
		validation.Field(&pkg.Description, validation.Required, validation.Length(MinDescriptionLength, MaxDescriptionLength)),
		validation.Field(&pkg.SupportInfo, validation.NilOrNotEmpty),
		validation.Field(&pkg.Version, validation.Required, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&pkg.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelSapPartner, PolicyLevelCustom, PolicyLevelNone), validation.When(pkg.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&pkg.CustomPolicyLevel, validation.When(pkg.PolicyLevel != nil && *pkg.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&pkg.PackageLinks, validation.By(validatePackageLinks)),
		validation.Field(&pkg.Links, validation.By(validateORDLinks)),
		validation.Field(&pkg.Vendor, validation.Required, validation.When(checkPackagePolicyLevel(docPolicyLevel, pkg.PolicyLevel, PolicyLevelSap), validation.In(SapVendor)), validation.When(checkPackagePolicyLevel(docPolicyLevel, pkg.PolicyLevel, PolicyLevelSapPartner), validation.NotIn(SapVendor)), validation.Match(regexp.MustCompile(VendorOrdIDRegex))),
		validation.Field(&pkg.PartOfProducts, validation.Required, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ProductOrdIDRegex))
		})),
		validation.Field(&pkg.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&pkg.Labels, validation.By(validateORDLabels)),
		validation.Field(&pkg.Countries, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CountryRegex))
		})),
		validation.Field(&pkg.LineOfBusiness,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsSAP(docPolicyLevel, pkg.PolicyLevel, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&pkg.Industry,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsSAP(docPolicyLevel, pkg.PolicyLevel, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, Industries)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&pkg.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func checkPackagePolicyLevel(docPolicyLevel *string, pkgPolicyLevel *string, policyLevelValue string) bool {
	policyLevel := str.PtrStrToStr(docPolicyLevel)
	if pkgPolicyLevel != nil {
		policyLevel = str.PtrStrToStr(pkgPolicyLevel)
	}

	return policyLevel == policyLevelValue
}

func validatePackageInputWithSuppressedErrors(pkg *model.PackageInput, packagesFromDB map[string]*model.Package, resourceHashes map[string]uint64) error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.Version, validation.By(func(value interface{}) error {
			return validatePackageVersionInput(value, *pkg, packagesFromDB, resourceHashes)
		})))
}

func validateBundleInput(bndl *model.BundleCreateInput, credentialExchangeStrategyTenantMappings map[string]processors.CredentialExchangeStrategyTenantMapping) error {
	return validation.ValidateStruct(bndl,
		validation.Field(&bndl.OrdID, validation.Required, validation.Match(regexp.MustCompile(BundleOrdIDRegex))),
		validation.Field(&bndl.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&bndl.Name, validation.Required),
		validation.Field(&bndl.ShortDescription, optionalShortDescriptionRules...),
		validation.Field(&bndl.Description, validation.NilOrNotEmpty, validation.Length(MinDescriptionLength, MaxDescriptionLength)),
		validation.Field(&bndl.Version, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&bndl.Links, validation.By(validateORDLinks)),
		validation.Field(&bndl.Labels, validation.By(validateORDLabels)),
		validation.Field(&bndl.CredentialExchangeStrategies, validation.By(func(value interface{}) error {
			return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
				"type": {
					validation.Required,
					validation.In(custom),
				},
				"callbackUrl": {
					is.RequestURI,
				},
			}, validateCustomType(credentialExchangeStrategyTenantMappings), validateCustomDescription)
		})),
		validation.Field(&bndl.CorrelationIDs, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CorrelationIDsRegex))
		})),
		validation.Field(&bndl.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
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
		validation.Field(&api.OrdID, validation.Required, validation.Match(regexp.MustCompile(APIOrdIDRegex))),
		validation.Field(&api.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&api.Name, validation.Required),
		validation.Field(&api.ShortDescription, shortDescriptionRules...),
		validation.Field(&api.Description, validation.Required, validation.Length(MinDescriptionLength, MaxDescriptionLength)),
		validation.Field(&api.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelSapPartner, PolicyLevelCustom, PolicyLevelNone), validation.When(api.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&api.CustomPolicyLevel, validation.When(api.PolicyLevel != nil && *api.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&api.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&api.OrdPackageID, validation.Required, validation.Match(regexp.MustCompile(PackageOrdIDRegex))),
		validation.Field(&api.APIProtocol, validation.Required, validation.In(APIProtocolODataV2, APIProtocolODataV4, APIProtocolSoapInbound, APIProtocolSoapOutbound, APIProtocolRest, APIProtocolSapRfc, APIProtocolWebsocket, APIProtocolSAPSQLAPIV1)),
		validation.Field(&api.Visibility, validation.Required, validation.In(APIVisibilityPublic, APIVisibilityInternal, APIVisibilityPrivate)),
		validation.Field(&api.PartOfProducts, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ProductOrdIDRegex))
		})),
		validation.Field(&api.SupportedUseCases,
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsContainsInMap(value, SupportedUseCases)
			}),
		),
		validation.Field(&api.Hierarchy, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&api.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&api.Countries, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CountryRegex))
		})),
		validation.Field(&api.LineOfBusiness,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsSAP(docPolicyLevel, api.PolicyLevel, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&api.Industry,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsSAP(docPolicyLevel, api.PolicyLevel, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, Industries)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&api.ResourceDefinitions, validation.By(func(value interface{}) error {
			return validateAPIResourceDefinitions(value, *api, docPolicyLevel)
		})),
		validation.Field(&api.APIResourceLinks, validation.By(validateAPILinks)),
		validation.Field(&api.Links, validation.By(validateORDLinks)),
		validation.Field(&api.ReleaseStatus, validation.Required, validation.In(ReleaseStatusBeta, ReleaseStatusActive, ReleaseStatusDeprecated)),
		validation.Field(&api.SunsetDate, validation.When(*api.ReleaseStatus == ReleaseStatusDeprecated, validation.Required), validation.When(api.SunsetDate != nil, validation.By(isValidDate))),
		validation.Field(&api.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&api.TargetURLs, validation.By(validateEntryPoints), validation.When(api.TargetURLs == nil, validation.By(notPartOfConsumptionBundles(api.PartOfConsumptionBundles)))),
		validation.Field(&api.Labels, validation.By(validateORDLabels)),
		validation.Field(&api.ImplementationStandard, validation.In(APIImplementationStandardDocumentAPI, APIImplementationStandardServiceBroker, APIImplementationStandardCsnExposure, APIImplementationStandardApeAPI, APIImplementationStandardCdiAPI, APIImplementationStandardCustom)),
		validation.Field(&api.CustomImplementationStandard, validation.When(api.ImplementationStandard != nil && *api.ImplementationStandard == APIImplementationStandardCustom, validation.Required, validation.Match(regexp.MustCompile(CustomImplementationStandardRegex))).Else(validation.Empty)),
		validation.Field(&api.CustomImplementationStandardDescription, validation.When(api.ImplementationStandard != nil && *api.ImplementationStandard == APIImplementationStandardCustom, validation.Required).Else(validation.Empty)),
		validation.Field(&api.PartOfConsumptionBundles, validation.By(func(value interface{}) error {
			return validateAPIPartOfConsumptionBundles(value, api.TargetURLs, regexp.MustCompile(BundleOrdIDRegex))
		})),
		validation.Field(&api.DefaultConsumptionBundle, validation.Match(regexp.MustCompile(BundleOrdIDRegex)), validation.By(func(value interface{}) error {
			return validateDefaultConsumptionBundle(value, api.PartOfConsumptionBundles)
		})),
		validation.Field(&api.Extensible, validation.By(func(value interface{}) error {
			return validateExtensibleField(value, docPolicyLevel)
		})),
		validation.Field(&api.DocumentationLabels, validation.By(validateDocumentationLabels)),
		validation.Field(&api.CorrelationIDs, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&api.Direction, validation.In(APIDirectionInbound, APIDirectionMixed, APIDirectionOutbound)),
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
		validation.Field(&event.OrdID, validation.Required, validation.Match(regexp.MustCompile(EventOrdIDRegex))),
		validation.Field(&event.LocalTenantID, validation.NilOrNotEmpty, validation.Length(MinLocalTenantIDLength, MaxLocalTenantIDLength)),
		validation.Field(&event.Name, validation.Required),
		validation.Field(&event.ShortDescription, shortDescriptionRules...),
		validation.Field(&event.Description, validation.Required, validation.Length(MinDescriptionLength, MaxDescriptionLength)),
		validation.Field(&event.PolicyLevel, validation.In(PolicyLevelSap, PolicyLevelSapPartner, PolicyLevelCustom, PolicyLevelNone), validation.When(event.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&event.CustomPolicyLevel, validation.When(event.PolicyLevel != nil && *event.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&event.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&event.OrdPackageID, validation.Required, validation.Match(regexp.MustCompile(PackageOrdIDRegex))),
		validation.Field(&event.Visibility, validation.Required, validation.In(APIVisibilityPublic, APIVisibilityInternal, APIVisibilityPrivate)),
		validation.Field(&event.PartOfProducts, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ProductOrdIDRegex))
		})),
		validation.Field(&event.Hierarchy, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&event.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&event.Countries, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CountryRegex))
		})),
		validation.Field(&event.LineOfBusiness,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsSAP(docPolicyLevel, event.PolicyLevel, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&event.Industry,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsSAP(docPolicyLevel, event.PolicyLevel, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, Industries)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&event.ResourceDefinitions),
		validation.Field(&event.Links, validation.By(validateORDLinks)),
		validation.Field(&event.ReleaseStatus, validation.Required, validation.In(ReleaseStatusBeta, ReleaseStatusActive, ReleaseStatusDeprecated)),
		validation.Field(&event.SunsetDate, validation.When(*event.ReleaseStatus == ReleaseStatusDeprecated, validation.Required), validation.When(event.SunsetDate != nil, validation.By(isValidDate))),
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
			return validateExtensibleField(value, docPolicyLevel)
		})),
		validation.Field(&event.DocumentationLabels, validation.By(validateDocumentationLabels)),
		validation.Field(&event.CorrelationIDs, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
	)
}

func validateEventInputWithSuppressedErrors(event *model.EventDefinitionInput, eventsFromDB map[string]*model.EventDefinition, eventHashes map[string]uint64) error {
	return validation.ValidateStruct(event,
		validation.Field(&event.VersionInput.Value, validation.By(func(value interface{}) error {
			return validateEventDefinitionVersionInput(value, *event, eventsFromDB, eventHashes)
		})))
}

func validateProductInput(product *model.ProductInput) error {
	productOrdIDNamespace := strings.Split(product.OrdID, ":")[0]

	return validation.ValidateStruct(product,
		validation.Field(&product.OrdID, validation.Required, validation.Match(regexp.MustCompile(ProductOrdIDRegex))),
		validation.Field(&product.Title, validation.Length(1, 255), validation.Required),
		validation.Field(&product.ShortDescription, shortDescriptionRules...),
		validation.Field(&product.Vendor, validation.Required,
			validation.Match(regexp.MustCompile(VendorOrdIDRegex)),
			validation.When(regexp.MustCompile(SAPProductOrdIDNamespaceRegex).MatchString(productOrdIDNamespace), validation.In(SapVendor)).Else(validation.NotIn(SapVendor)),
		),
		validation.Field(&product.Parent, validation.When(product.Parent != nil, validation.Match(regexp.MustCompile(ProductOrdIDRegex)))),
		validation.Field(&product.CorrelationIDs, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CorrelationIDsRegex))
		})),
		validation.Field(&product.Labels, validation.By(validateORDLabels)),
		validation.Field(&product.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&product.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func validateVendorInput(vendor *model.VendorInput) error {
	return validation.ValidateStruct(vendor,
		validation.Field(&vendor.OrdID, validation.Required, validation.Match(regexp.MustCompile(VendorOrdIDRegex))),
		validation.Field(&vendor.Title, validation.Length(1, 255), validation.Required),
		validation.Field(&vendor.Labels, validation.By(validateORDLabels)),
		validation.Field(&vendor.Partners, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(VendorPartnersRegex))
		})),
		validation.Field(&vendor.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&vendor.DocumentationLabels, validation.By(validateDocumentationLabels)),
	)
}

func validateTombstoneInput(tombstone *model.TombstoneInput) error {
	return validation.ValidateStruct(tombstone,
		validation.Field(&tombstone.OrdID, validation.Required, validation.Match(regexp.MustCompile(TombstoneOrdIDRegex))),
		validation.Field(&tombstone.RemovalDate, validation.Required, validation.By(isValidDate)))
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

func validateORDChangeLogEntries(value interface{}) error {
	return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"version": {
			validation.Required,
			validation.Match(regexp.MustCompile(SemVerRegex)),
		},
		"releaseStatus": {
			validation.Required,
			validation.In(ReleaseStatusBeta, ReleaseStatusActive, ReleaseStatusDeprecated),
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
	return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"title": {
			validation.Length(1, 255),
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
	})
}

func validatePackageLinks(value interface{}) error {
	return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
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

func validateAPILinks(value interface{}) error {
	return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"type": {
			validation.Required,
			validation.In("api-documentation", "authentication", "client-registration", "console", "payment", "service-level-agreement", "support", "custom"),
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

	isPolicyCoreOrPartner := policyLevel == PolicyLevelSap || policyLevel == PolicyLevelSapPartner
	wsdlTypeExists := resourceDefinitionTypes[model.APISpecTypeWsdlV1] || resourceDefinitionTypes[model.APISpecTypeWsdlV2]
	if isPolicyCoreOrPartner && (apiProtocol == APIProtocolSoapInbound || apiProtocol == APIProtocolSoapOutbound) && !wsdlTypeExists {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='soap-inbound' or 'soap-outbound' it is mandatory to provide either WSDL V2 or WSDL V1 definitions")
	}

	edmxTypeExists := resourceDefinitionTypes[model.APISpecTypeEDMX]
	openAPITypeExists := resourceDefinitionTypes[model.APISpecTypeOpenAPIV2] || resourceDefinitionTypes[model.APISpecTypeOpenAPIV3]
	if isPolicyCoreOrPartner && (apiProtocol == APIProtocolODataV2 || apiProtocol == APIProtocolODataV4) && !(edmxTypeExists && openAPITypeExists) {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='odata-v2' or 'odata-v4' it is mandatory to not only provide edmx definitions, but also OpenAPI definitions")
	}

	if isPolicyCoreOrPartner && apiProtocol == APIProtocolRest && !openAPITypeExists {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='rest' it is mandatory to provide either OpenAPI 3 or OpenAPI 2 definitions")
	}

	rfcMetadataTypeExists := resourceDefinitionTypes[model.APISpecTypeRfcMetadata]
	if isPolicyCoreOrPartner && apiProtocol == APIProtocolSapRfc && !rfcMetadataTypeExists {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='sap-rfc' it is mandatory to provide SAP RFC definitions")
	}

	if apiProtocol == APIProtocolWebsocket && (api.ImplementationStandard == nil || !resourceDefinitionTypes[model.APISpecTypeCustom]) {
		return errors.New("for APIResources with apiProtocol='websocket' it is mandatory to provide implementationStandard definition and type to be set to custom")
	}

	if apiProtocol == APIProtocolSAPSQLAPIV1 && !(resourceDefinitionTypes[model.APISpecTypeCustom] || resourceDefinitionTypes[model.APISpecTypeSQLAPIDefinitionV1]) {
		return errors.New("for APIResources with apiProtocol='sap-sql-api-v1' it is mandatory type to be set either to sap-sql-api-definition-v1 or custom")
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

func noNewLines(s string) bool {
	return !strings.Contains(s, "\\n")
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

func validateJSONArrayOfObjects(arr interface{}, elementFieldRules map[string][]validation.Rule, crossFieldRules ...func(gjson.Result) error) error {
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
		for field, rules := range elementFieldRules {
			if err := validation.Validate(el.Get(field).Value(), rules...); err != nil {
				return errors.Wrapf(err, "error validating field %s", field)
			}
			for _, f := range crossFieldRules {
				if err := f(el); err != nil {
					return err
				}
			}
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

func validateCustomType(credentialExchangeStrategyTenantMappings map[string]processors.CredentialExchangeStrategyTenantMapping) func(el gjson.Result) error {
	return func(el gjson.Result) error {
		if el.Get("customType").Exists() && el.Get("type").String() != custom {
			return errors.New("if customType is provided, type should be set to 'custom'")
		}

		customType := el.Get("customType").String()
		if _, ok := credentialExchangeStrategyTenantMappings[customType]; strings.Contains(customType, processors.TenantMappingCustomTypeIdentifier) && !ok {
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

func validateExtensibleField(value interface{}, policyLevelInput *string) error {
	policyLevel := str.PtrStrToStr(policyLevelInput)

	if (policyLevel == PolicyLevelSap || policyLevel == PolicyLevelSapPartner) && (value == nil || value.(json.RawMessage) == nil) {
		return errors.Errorf("`extensible` field must be provided when `policyLevel` is either `%s` or `%s`", PolicyLevelSap, PolicyLevelSapPartner)
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
