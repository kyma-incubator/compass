package open_resource_discovery

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Disclaimer: All regexes below are provided by the ORD spec itself.
const (
	SemVerRegex         = "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"
	PackageOrdIDRegex   = "^([a-zA-Z0-9._\\-]+):(package):([a-zA-Z0-9._\\-]+):(alpha|beta|v[0-9]+)$"
	VendorOrdIDRegex    = "^([a-zA-Z0-9._\\-]+):(vendor):([a-zA-Z0-9._\\-]+):()$"
	ProductOrdIDRegex   = "^([a-zA-Z0-9._\\-]+):(product):([a-zA-Z0-9._\\-]+):()$"
	BundleOrdIDRegex    = "^([a-zA-Z0-9._\\-]+):(consumptionBundle):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	TombstoneOrdIDRegex = "^([a-zA-Z0-9._\\-]+):(package|consumptionBundle|product|vendor|apiResource|eventResource):([a-zA-Z0-9._\\-]+):(alpha|beta|v[0-9]+|)$"

	SystemInstanceBaseURLRegex        = "^http[s]?:\\/\\/[^:\\/\\s]+\\.[^:\\/\\s\\.]+(:\\d+)?$"
	StringArrayElementRegex           = "^[a-zA-Z0-9 -\\.\\/]*$"
	CountryRegex                      = "^[A-Z]{2}$"
	ApiOrdIDRegex                     = "^([a-zA-Z0-9._\\-]+):(apiResource):([a-zA-Z0-9._\\-]+):(alpha|beta|v[0-9]+)$"
	EventOrdIDRegex                   = "^([a-zA-Z0-9._\\-]+):(eventResource):([a-zA-Z0-9._\\-]+):(alpha|beta|v[0-9]+)$"
	CorrelationIDsRegex               = "^([a-zA-Z0-9._\\-]+):([a-zA-Z0-9._\\-\\/]+)$"
	LabelsKeyRegex                    = "^[a-zA-Z0-9-_.]*$"
	CustomImplementationStandardRegex = "^([a-z0-9.]+):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	VendorPartnersRegex               = "^([a-zA-Z0-9._\\-]+):(vendor):([a-zA-Z0-9._\\-]+):()$"
	CustomPolicyLevelRegex            = "^([a-z0-9.]+):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
)

const (
	PolicyLevelSap        string = "sap:core:v1"
	PolicyLevelSapPartner string = "sap:partner:v1"
	PolicyLevelCustom     string = "custom"

	ReleaseStatusBeta       string = "beta"
	ReleaseStatusActive     string = "active"
	ReleaseStatusDeprecated string = "deprecated"

	ApiProtocolODataV2      string = "odata-v2"
	ApiProtocolODataV4      string = "odata-v4"
	ApiProtocolSoapInbound  string = "soap-inbound"
	ApiProtocolSoapOutbound string = "soap-outbound"
	ApiProtocolRest         string = "rest"
	ApiProtocolSapRfc       string = "sap-rfc"

	ApiVisibilityPublic   string = "public"
	ApiVisibilityPrivate  string = "private"
	ApiVisibilityInternal string = "internal"

	ApiImplementationStandardDocumentApi   string = "sap:ord-document-api:v1"
	ApiImplementationStandardServiceBroker string = "cff:open-service-broker:v2"
	ApiImplementationStandardCsnExposure   string = "sap:csn-exposure:v1"
	ApiImplementationStandardCustom        string = "custom"

	SapTitle      = "SAP SE"
	SapVendor     = "sap:vendor:SAP:"
	PartnerVendor = "partner:vendor:SAP:"
)

var (
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
)

var shortDescriptionRules = []validation.Rule{
	validation.Required, validation.Length(1, 255), validation.NewStringRule(noNewLines, "short description should not contain line breaks"),
}

var descriptionRules = []validation.Rule{
	validation.Required, validation.Length(1, 255), validation.NewStringRule(noNewLines, "description should not contain line breaks"),
}

func ValidateSystemInstanceInput(app *model.Application) error {
	return validation.ValidateStruct(app,
		validation.Field(&app.CorrelationIds, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CorrelationIDsRegex))
		})),
		validation.Field(&app.BaseURL, is.RequestURI, validation.Match(regexp.MustCompile(SystemInstanceBaseURLRegex))),
		validation.Field(&app.Labels, validation.By(validateORDLabels)),
	)
}

func validateDocumentInput(doc *Document) error {
	return validation.ValidateStruct(doc, validation.Field(&doc.OpenResourceDiscovery, validation.Required, validation.In("1.0")))
}

func validatePackageInput(pkg *model.PackageInput) error {
	return validation.ValidateStruct(pkg,
		validation.Field(&pkg.OrdID, validation.Required, validation.Match(regexp.MustCompile(PackageOrdIDRegex))),
		validation.Field(&pkg.Title, validation.Required),
		validation.Field(&pkg.ShortDescription, shortDescriptionRules...),
		validation.Field(&pkg.Description, validation.Required),
		validation.Field(&pkg.Version, validation.Required, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&pkg.PolicyLevel, validation.Required, validation.In(PolicyLevelSap, PolicyLevelSapPartner, PolicyLevelCustom), validation.When(pkg.CustomPolicyLevel != nil, validation.In(PolicyLevelCustom))),
		validation.Field(&pkg.CustomPolicyLevel, validation.When(pkg.PolicyLevel != PolicyLevelCustom, validation.Empty), validation.Match(regexp.MustCompile(CustomPolicyLevelRegex))),
		validation.Field(&pkg.PackageLinks, validation.By(validatePackageLinks)),
		validation.Field(&pkg.Links, validation.By(validateORDLinks)),
		validation.Field(&pkg.Vendor, validation.Required, validation.When(pkg.PolicyLevel == PolicyLevelSap, validation.In(SapVendor)), validation.When(pkg.PolicyLevel == PolicyLevelSapPartner, validation.NotIn(SapVendor)), validation.Match(regexp.MustCompile(VendorOrdIDRegex))),
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
				if pkg.PolicyLevel == PolicyLevelCustom {
					return nil
				}
				return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&pkg.Industry,
			validation.By(func(value interface{}) error {
				if pkg.PolicyLevel == PolicyLevelCustom {
					return nil
				}
				return validateJSONArrayOfStringsContainsInMap(value, Industries)
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
	)
}

func validateBundleInput(bndl *model.BundleCreateInput) error {
	return validation.ValidateStruct(bndl,
		validation.Field(&bndl.OrdID, validation.Required, validation.Match(regexp.MustCompile(BundleOrdIDRegex))),
		validation.Field(&bndl.Name, validation.Required),
		validation.Field(&bndl.ShortDescription, shortDescriptionRules...),
		validation.Field(&bndl.Description, descriptionRules...),
		validation.Field(&bndl.Links, validation.By(validateORDLinks)),
		validation.Field(&bndl.Labels, validation.By(validateORDLabels)),
		validation.Field(&bndl.CredentialExchangeStrategies, validation.By(func(value interface{}) error {
			return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
				"type": {
					validation.Required,
					validation.In("custom"),
				},
				"callbackUrl": {
					is.RequestURI,
				},
			}, validateCustomType, validateCustomDescription)
		})),
	)
}

func validateAPIInput(api *model.APIDefinitionInput, packagePolicyLevels map[string]string) error {
	return validation.ValidateStruct(api,
		validation.Field(&api.OrdID, validation.Required, validation.Match(regexp.MustCompile(ApiOrdIDRegex))),
		validation.Field(&api.Name, validation.Required),
		validation.Field(&api.ShortDescription, shortDescriptionRules...),
		validation.Field(&api.Description, validation.Required),
		validation.Field(&api.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&api.OrdPackageID, validation.Required, validation.Match(regexp.MustCompile(PackageOrdIDRegex))),
		validation.Field(&api.ApiProtocol, validation.Required, validation.In(ApiProtocolODataV2, ApiProtocolODataV4, ApiProtocolSoapInbound, ApiProtocolSoapOutbound, ApiProtocolRest, ApiProtocolSapRfc)),
		validation.Field(&api.Visibility, validation.Required, validation.In(ApiVisibilityPublic, ApiVisibilityInternal, ApiVisibilityPrivate)),
		validation.Field(&api.PartOfProducts, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ProductOrdIDRegex))
		})),
		validation.Field(&api.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&api.Countries, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CountryRegex))
		})),
		validation.Field(&api.LineOfBusiness,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsNotCustom(api.OrdPackageID, packagePolicyLevels, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&api.Industry,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsNotCustom(api.OrdPackageID, packagePolicyLevels, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, Industries)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&api.ResourceDefinitions, validation.By(func(value interface{}) error {
			return validateAPIResourceDefinitions(value, *api, packagePolicyLevels)
		})),
		validation.Field(&api.APIResourceLinks, validation.By(validateAPILinks)),
		validation.Field(&api.Links, validation.By(validateORDLinks)),
		validation.Field(&api.ReleaseStatus, validation.Required, validation.In(ReleaseStatusBeta, ReleaseStatusActive, ReleaseStatusDeprecated)),
		validation.Field(&api.SunsetDate, validation.When(*api.ReleaseStatus == ReleaseStatusDeprecated, validation.Required), validation.When(api.SunsetDate != nil, validation.By(isValidDate(api.SunsetDate)))),
		validation.Field(&api.Successors, validation.When(*api.ReleaseStatus == ReleaseStatusDeprecated, validation.Required), validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ApiOrdIDRegex))
		})),
		validation.Field(&api.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&api.TargetURLs, validation.By(validateEntryPoints), validation.When(api.TargetURLs == nil, validation.By(notPartOfConsumptionBundles(api.PartOfConsumptionBundles)))),
		validation.Field(&api.Labels, validation.By(validateORDLabels)),
		validation.Field(&api.ImplementationStandard, validation.In(ApiImplementationStandardDocumentApi, ApiImplementationStandardServiceBroker, ApiImplementationStandardCsnExposure, ApiImplementationStandardCustom)),
		validation.Field(&api.CustomImplementationStandard, validation.When(api.ImplementationStandard != nil && *api.ImplementationStandard == ApiImplementationStandardCustom, validation.Required, validation.Match(regexp.MustCompile(CustomImplementationStandardRegex))).Else(validation.Empty)),
		validation.Field(&api.CustomImplementationStandardDescription, validation.When(api.ImplementationStandard != nil && *api.ImplementationStandard == ApiImplementationStandardCustom, validation.Required).Else(validation.Empty)),
		validation.Field(&api.PartOfConsumptionBundles, validation.By(func(value interface{}) error {
			return validateAPIPartOfConsumptionBundles(value, api.TargetURLs, regexp.MustCompile(BundleOrdIDRegex))
		})),
		validation.Field(&api.Extensible, validation.By(func(value interface{}) error {
			return validateExtensibleField(value, api.OrdPackageID, packagePolicyLevels)
		})),
	)
}

func validateEventInput(event *model.EventDefinitionInput, packagePolicyLevels map[string]string) error {
	return validation.ValidateStruct(event,
		validation.Field(&event.OrdID, validation.Required, validation.Match(regexp.MustCompile(EventOrdIDRegex))),
		validation.Field(&event.Name, validation.Required),
		validation.Field(&event.ShortDescription, shortDescriptionRules...),
		validation.Field(&event.Description, validation.Required),
		validation.Field(&event.VersionInput.Value, validation.Required, validation.Match(regexp.MustCompile(SemVerRegex))),
		validation.Field(&event.OrdPackageID, validation.Required, validation.Match(regexp.MustCompile(PackageOrdIDRegex))),
		validation.Field(&event.Visibility, validation.Required, validation.In(ApiVisibilityPublic, ApiVisibilityInternal, ApiVisibilityPrivate)),
		validation.Field(&event.PartOfProducts, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(ProductOrdIDRegex))
		})),
		validation.Field(&event.Tags, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
		})),
		validation.Field(&event.Countries, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CountryRegex))
		})),
		validation.Field(&event.LineOfBusiness,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsNotCustom(event.OrdPackageID, packagePolicyLevels, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, LineOfBusinesses)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&event.Industry,
			validation.By(func(value interface{}) error {
				return validateWhenPolicyLevelIsNotCustom(event.OrdPackageID, packagePolicyLevels, func() error {
					return validateJSONArrayOfStringsContainsInMap(value, Industries)
				})
			}),
			validation.By(func(value interface{}) error {
				return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(StringArrayElementRegex))
			}),
		),
		validation.Field(&event.ResourceDefinitions, validation.By(func(value interface{}) error {
			return validateEventResourceDefinition(value, *event, packagePolicyLevels)
		})),
		validation.Field(&event.Links, validation.By(validateORDLinks)),
		validation.Field(&event.ReleaseStatus, validation.Required, validation.In(ReleaseStatusBeta, ReleaseStatusActive, ReleaseStatusDeprecated)),
		validation.Field(&event.SunsetDate, validation.When(*event.ReleaseStatus == ReleaseStatusDeprecated, validation.Required), validation.When(event.SunsetDate != nil, validation.By(isValidDate(event.SunsetDate)))),
		validation.Field(&event.Successors, validation.When(*event.ReleaseStatus == ReleaseStatusDeprecated, validation.Required), validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(EventOrdIDRegex))
		})),
		validation.Field(&event.ChangeLogEntries, validation.By(validateORDChangeLogEntries)),
		validation.Field(&event.Labels, validation.By(validateORDLabels)),
		validation.Field(&event.PartOfConsumptionBundles, validation.By(func(value interface{}) error {
			return validateEventPartOfConsumptionBundles(value, regexp.MustCompile(BundleOrdIDRegex))
		})),
		validation.Field(&event.Extensible, validation.By(func(value interface{}) error {
			return validateExtensibleField(value, event.OrdPackageID, packagePolicyLevels)
		})),
	)
}

func validateProductInput(product *model.ProductInput) error {
	productOrdIDNamespace := strings.Split(product.OrdID, ":")[0]

	return validation.ValidateStruct(product,
		validation.Field(&product.OrdID, validation.Required, validation.Match(regexp.MustCompile(ProductOrdIDRegex))),
		validation.Field(&product.Title, validation.Required),
		validation.Field(&product.ShortDescription, shortDescriptionRules...),
		validation.Field(&product.Vendor, validation.Required,
			validation.Match(regexp.MustCompile(VendorOrdIDRegex)),
			validation.When(regexp.MustCompile("^(sap)((\\.)([a-zA-Z0-9._\\-])+)*$").MatchString(productOrdIDNamespace), validation.In(SapVendor)).Else(validation.NotIn(SapVendor)),
		),
		validation.Field(&product.Parent, validation.When(product.Parent != nil, validation.Match(regexp.MustCompile(ProductOrdIDRegex)))),
		validation.Field(&product.CorrelationIds, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(CorrelationIDsRegex))
		})),
		validation.Field(&product.Labels, validation.By(validateORDLabels)),
	)
}

func validateVendorInput(vendor *model.VendorInput) error {
	return validation.ValidateStruct(vendor,
		validation.Field(&vendor.OrdID, validation.Required, validation.Match(regexp.MustCompile(VendorOrdIDRegex))),
		validation.Field(&vendor.Title, validation.Required),
		validation.Field(&vendor.Labels, validation.By(validateORDLabels)),
		validation.Field(&vendor.Partners, validation.By(func(value interface{}) error {
			return validateJSONArrayOfStringsMatchPattern(value, regexp.MustCompile(VendorPartnersRegex))
		})),
	)
}

func validateTombstoneInput(tombstone *model.TombstoneInput) error {
	return validation.ValidateStruct(tombstone,
		validation.Field(&tombstone.OrdID, validation.Required, validation.Match(regexp.MustCompile(TombstoneOrdIDRegex))),
		validation.Field(&tombstone.RemovalDate, validation.Required, validation.Date(time.RFC3339)))
}

func validateORDLabels(val interface{}) error {
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
		if err = validation.Validate(key.String(), validation.Match(regexp.MustCompile(LabelsKeyRegex))); err != nil {
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
			validation.Date("2006-01-02"),
		},
		"url": {
			is.RequestURI,
		},
	})
}

func validateORDLinks(value interface{}) error {
	return validateJSONArrayOfObjects(value, map[string][]validation.Rule{
		"title": {
			validation.Required,
		},
		"url": {
			validation.Required,
			is.RequestURI,
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
		if el.Get("customType").Exists() && el.Get("type").String() != "custom" {
			return errors.New("if customType is provided, type should be set to 'custom'")
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
		if el.Get("customType").Exists() && el.Get("type").String() != "custom" {
			return errors.New("if customType is provided, type should be set to 'custom'")
		}
		return nil
	})
}

func validateAPIResourceDefinitions(value interface{}, api model.APIDefinitionInput, packagePolicyLevels map[string]string) error {
	if value == nil {
		return nil
	}

	pkgOrdID := str.PtrStrToStr(api.OrdPackageID)
	policyLevel := packagePolicyLevels[pkgOrdID]
	apiVisibility := str.PtrStrToStr(api.Visibility)
	apiProtocol := str.PtrStrToStr(api.ApiProtocol)
	resourceDefinitions := api.ResourceDefinitions

	isResourceDefinitionMandatory := !(policyLevel == PolicyLevelSap && apiVisibility == ApiVisibilityPrivate)
	if len(resourceDefinitions) == 0 && isResourceDefinitionMandatory {
		return errors.New("when api resource visibility is public or internal, resource definitions must be provided")
	}

	if len(resourceDefinitions) == 0 && !isResourceDefinitionMandatory {
		return nil
	}

	resourceDefinitionTypes := make(map[model.APISpecType]bool, 0)

	for _, rd := range resourceDefinitions {
		resourceDefinitionType := rd.Type
		resourceDefinitionTypes[resourceDefinitionType] = true
	}

	isPolicyCoreOrPartner := policyLevel == PolicyLevelSap || policyLevel == PolicyLevelSapPartner
	wsdlTypeExists := resourceDefinitionTypes[model.APISpecTypeWsdlV1] || resourceDefinitionTypes[model.APISpecTypeWsdlV2]
	if isPolicyCoreOrPartner && (apiProtocol == ApiProtocolSoapInbound || apiProtocol == ApiProtocolSoapOutbound) && !wsdlTypeExists {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='soap-inbound' or 'soap-outbound' it is mandatory to provide either WSDL V2 or WSDL V1 definitions")
	}

	edmxTypeExists := resourceDefinitionTypes[model.APISpecTypeEDMX]
	openAPITypeExists := resourceDefinitionTypes[model.APISpecTypeOpenAPIV2] || resourceDefinitionTypes[model.APISpecTypeOpenAPIV3]
	if isPolicyCoreOrPartner && (apiProtocol == ApiProtocolODataV2 || apiProtocol == ApiProtocolODataV4) && !(edmxTypeExists && openAPITypeExists) {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='odata-v2' or 'odata-v4' it is mandatory to not only provide edmx definitions, but also OpenAPI definitions.")
	}

	if isPolicyCoreOrPartner && apiProtocol == ApiProtocolRest && !openAPITypeExists {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='rest' it is mandatory to provide either OpenAPI 3 or OpenAPI 2 definitions")
	}

	rfcMetadataTypeExists := resourceDefinitionTypes[model.APISpecTypeRfcMetadata]
	if isPolicyCoreOrPartner && apiProtocol == ApiProtocolSapRfc && !rfcMetadataTypeExists {
		return errors.New("for APIResources of policyLevel='sap' or 'sap-partner' and with apiProtocol='sap-rfc' it is mandatory to provide SAP RFC definitions")
	}

	return nil
}

func validateEventResourceDefinition(value interface{}, event model.EventDefinitionInput, packagePolicyLevels map[string]string) error {
	if value == nil {
		return nil
	}

	pkgOrdID := str.PtrStrToStr(event.OrdPackageID)
	policyLevel := packagePolicyLevels[pkgOrdID]
	apiVisibility := str.PtrStrToStr(event.Visibility)

	if policyLevel == PolicyLevelSap && apiVisibility == ApiVisibilityPrivate {
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

func noNewLines(s string) bool {
	return !strings.Contains(s, "\\n")
}

func validateWhenPolicyLevelIsNotCustom(apiOrdID *string, packagePolicyLevels map[string]string, validationFunc func() error) error {
	pkgOrdID := str.PtrStrToStr(apiOrdID)
	policyLevel := packagePolicyLevels[pkgOrdID]

	if policyLevel == PolicyLevelCustom {
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
		return errors.New("the json array should not be empty")
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
		return errors.New("the json array should not be empty")
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

func validateCustomType(el gjson.Result) error {
	if el.Get("customType").Exists() && el.Get("type").String() != "custom" {
		return errors.New("if customType is provided, type should be set to 'custom'")
	}
	return validation.Validate(el.Get("customType").String(), validation.Match(regexp.MustCompile("^([a-z0-9.]+):([a-zA-Z0-9._\\-]+):v([0-9]+)$")))
}

func validateCustomDescription(el gjson.Result) error {
	if el.Get("customDescription").Exists() && el.Get("type").String() != "custom" {
		return errors.New("if customDescription is provided, type should be set to 'custom'")
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

	seen := make(map[gjson.Result]bool)
	for _, val := range entryPoints {
		if seen[val] {
			return true
		} else {
			seen[val] = true
		}
	}
	return false
}

func isValidDate(date *string) validation.RuleFunc {
	return func(value interface{}) error {
		var err error
		if _, err = time.Parse("2006-01-02T15:04:05Z0700", *date); err == nil {
			return nil
		} else if _, err = time.Parse("2006-01-02T15:04:05Z07:00", *date); err == nil {
			return nil
		}
		return errors.New("invalid date")
	}
}

func notPartOfConsumptionBundles(partOfConsumptionBundles []*model.ConsumptionBundleReference) validation.RuleFunc {
	return func(value interface{}) error {
		if partOfConsumptionBundles != nil && len(partOfConsumptionBundles) > 0 {
			return errors.New("api without entry points can not be part of consumption bundle")
		}
		return nil
	}
}

func validateExtensibleField(value interface{}, ordPackageID *string, packagePolicyLevels map[string]string) error {
	pkgOrdID := str.PtrStrToStr(ordPackageID)
	policyLevel := packagePolicyLevels[pkgOrdID]

	if (policyLevel == PolicyLevelSap || policyLevel == PolicyLevelSapPartner) && (value == nil || value.(json.RawMessage) == nil) {
		return errors.New(fmt.Sprintf("`extensible` field must be provided when `policyLevel` is either `%s` or `%s`", PolicyLevelSap, PolicyLevelSapPartner))
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

	if supportedProperty.Exists() && (supportedValue == "manual" || supportedValue == "automatic") && (descriptionValue == "" || !ok) {
		return errors.New("if supported is either 'manual' or 'automatic', description should be provided")
	}
	return nil
}
