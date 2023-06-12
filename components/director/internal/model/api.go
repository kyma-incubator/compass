package model

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// APIDefinition missing godoc
type APIDefinition struct {
	ApplicationID                           *string
	ApplicationTemplateVersionID            *string
	PackageID                               *string
	Name                                    string
	Description                             *string
	TargetURLs                              json.RawMessage
	Group                                   *string //  group allows you to find the same API but in different version
	OrdID                                   *string
	LocalTenantID                           *string
	ShortDescription                        *string
	SystemInstanceAware                     *bool
	PolicyLevel                             *string
	CustomPolicyLevel                       *string
	APIProtocol                             *string
	Tags                                    json.RawMessage
	Countries                               json.RawMessage
	Links                                   json.RawMessage
	APIResourceLinks                        json.RawMessage
	ReleaseStatus                           *string
	SunsetDate                              *string
	Successors                              json.RawMessage
	ChangeLogEntries                        json.RawMessage
	Labels                                  json.RawMessage
	Visibility                              *string
	Disabled                                *bool
	PartOfProducts                          json.RawMessage
	LineOfBusiness                          json.RawMessage
	Industry                                json.RawMessage
	ImplementationStandard                  *string
	CustomImplementationStandard            *string
	CustomImplementationStandardDescription *string
	Version                                 *Version
	Extensible                              json.RawMessage
	ResourceHash                            *string
	Hierarchy                               json.RawMessage
	SupportedUseCases                       json.RawMessage
	DocumentationLabels                     json.RawMessage
	*BaseEntity
}

// GetType missing godoc
func (*APIDefinition) GetType() resource.Type {
	return resource.API
}

// APIDefinitionInput missing godoc
type APIDefinitionInput struct {
	OrdPackageID                            *string                       `json:"partOfPackage"`
	Tenant                                  string                        `json:",omitempty"`
	Name                                    string                        `json:"title"`
	Description                             *string                       `json:"description"`
	TargetURLs                              json.RawMessage               `json:"entryPoints"`
	Group                                   *string                       `json:",omitempty"` //  group allows you to find the same API but in different version
	OrdID                                   *string                       `json:"ordId"`
	LocalTenantID                           *string                       `json:"localTenantId"`
	ShortDescription                        *string                       `json:"shortDescription"`
	SystemInstanceAware                     *bool                         `json:"systemInstanceAware"`
	PolicyLevel                             *string                       `json:"policyLevel"`
	CustomPolicyLevel                       *string                       `json:"customPolicyLevel"`
	APIProtocol                             *string                       `json:"apiProtocol"`
	Tags                                    json.RawMessage               `json:"tags"`
	Countries                               json.RawMessage               `json:"countries"`
	Links                                   json.RawMessage               `json:"links"`
	APIResourceLinks                        json.RawMessage               `json:"apiResourceLinks"`
	ReleaseStatus                           *string                       `json:"releaseStatus"`
	SunsetDate                              *string                       `json:"sunsetDate"`
	Successors                              json.RawMessage               `json:"successors,omitempty"`
	ChangeLogEntries                        json.RawMessage               `json:"changelogEntries"`
	Labels                                  json.RawMessage               `json:"labels"`
	Visibility                              *string                       `json:"visibility"`
	Disabled                                *bool                         `json:"disabled"`
	PartOfProducts                          json.RawMessage               `json:"partOfProducts"`
	LineOfBusiness                          json.RawMessage               `json:"lineOfBusiness"`
	Industry                                json.RawMessage               `json:"industry"`
	ImplementationStandard                  *string                       `json:"implementationStandard"`
	CustomImplementationStandard            *string                       `json:"customImplementationStandard"`
	CustomImplementationStandardDescription *string                       `json:"customImplementationStandardDescription"`
	Extensible                              json.RawMessage               `json:"extensible"`
	ResourceDefinitions                     []*APIResourceDefinition      `json:"resourceDefinitions"`
	PartOfConsumptionBundles                []*ConsumptionBundleReference `json:"partOfConsumptionBundles"`
	DefaultConsumptionBundle                *string                       `json:"defaultConsumptionBundle"`
	Hierarchy                               json.RawMessage               `json:"hierarchy"`
	SupportedUseCases                       json.RawMessage               `json:"supported_use_cases"`
	DocumentationLabels                     json.RawMessage               `json:"documentationLabels"`

	*VersionInput `hash:"ignore"`
}

// APIResourceDefinition missing godoc
type APIResourceDefinition struct { // This is the place from where the specification for this API is fetched
	Type           APISpecType                     `json:"type"`
	CustomType     string                          `json:"customType"`
	MediaType      SpecFormat                      `json:"mediaType"`
	URL            string                          `json:"url"`
	AccessStrategy accessstrategy.AccessStrategies `json:"accessStrategies"`
}

// Validate missing godoc
func (rd *APIResourceDefinition) Validate() error {
	const CustomTypeRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	return validation.ValidateStruct(rd,
		validation.Field(&rd.Type, validation.Required, validation.In(APISpecTypeOpenAPIV2, APISpecTypeOpenAPIV3, APISpecTypeRaml, APISpecTypeEDMX,
			APISpecTypeCsdl, APISpecTypeWsdlV1, APISpecTypeWsdlV2, APISpecTypeRfcMetadata, APISpecTypeCustom, APISpecTypeSQLAPIDefinitionV1), validation.When(rd.CustomType != "", validation.In(APISpecTypeCustom))),
		validation.Field(&rd.CustomType, validation.When(rd.CustomType != "", validation.Match(regexp.MustCompile(CustomTypeRegex)))),
		validation.Field(&rd.MediaType, validation.Required, validation.In(SpecFormatApplicationJSON, SpecFormatTextYAML, SpecFormatApplicationXML, SpecFormatPlainText, SpecFormatOctetStream),
			validation.When(rd.Type == APISpecTypeOpenAPIV2 || rd.Type == APISpecTypeOpenAPIV3, validation.In(SpecFormatApplicationJSON, SpecFormatTextYAML)),
			validation.When(rd.Type == APISpecTypeRaml, validation.In(SpecFormatTextYAML)),
			validation.When(rd.Type == APISpecTypeEDMX, validation.In(SpecFormatApplicationXML)),
			validation.When(rd.Type == APISpecTypeCsdl, validation.In(SpecFormatApplicationJSON)),
			validation.When(rd.Type == APISpecTypeWsdlV1 || rd.Type == APISpecTypeWsdlV2, validation.In(SpecFormatApplicationXML)),
			validation.When(rd.Type == APISpecTypeRfcMetadata, validation.In(SpecFormatApplicationXML)),
			validation.When(rd.Type == APISpecTypeSQLAPIDefinitionV1, validation.In(SpecFormatApplicationJSON))),
		validation.Field(&rd.URL, validation.Required, is.RequestURI),
		validation.Field(&rd.AccessStrategy, validation.Required),
	)
}

// ToSpec missing godoc
func (rd *APIResourceDefinition) ToSpec() *SpecInput {
	var auth *AuthInput
	if as, ok := rd.AccessStrategy.GetSupported(); ok {
		asString := string(as)
		auth = &AuthInput{
			AccessStrategy: &asString,
		}
	}

	return &SpecInput{
		Format:     rd.MediaType,
		APIType:    &rd.Type,
		CustomType: &rd.CustomType,
		FetchRequest: &FetchRequestInput{
			URL:  rd.URL,
			Auth: auth,
		},
	}
}

// ConsumptionBundleReference missing godoc
type ConsumptionBundleReference struct {
	BundleOrdID      string `json:"ordId"`
	DefaultTargetURL string `json:"defaultEntryPoint"`
}

// APIDefinitionPage missing godoc
type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

// IsPageable missing godoc
func (APIDefinitionPage) IsPageable() {}

// ToAPIDefinition missing godoc
func (a *APIDefinitionInput) ToAPIDefinition(id string, resourceType resource.Type, resourceID string, packageID *string, apiHash uint64) *APIDefinition {
	if a == nil {
		return nil
	}

	var hash *string
	if apiHash != 0 {
		hash = str.Ptr(strconv.FormatUint(apiHash, 10))
	}

	api := &APIDefinition{
		PackageID:           packageID,
		Name:                a.Name,
		Description:         a.Description,
		TargetURLs:          a.TargetURLs,
		Group:               a.Group,
		OrdID:               a.OrdID,
		LocalTenantID:       a.LocalTenantID,
		ShortDescription:    a.ShortDescription,
		SystemInstanceAware: a.SystemInstanceAware,
		PolicyLevel:         a.PolicyLevel,
		CustomPolicyLevel:   a.CustomPolicyLevel,
		APIProtocol:         a.APIProtocol,
		Tags:                a.Tags,
		Countries:           a.Countries,
		Links:               a.Links,
		APIResourceLinks:    a.APIResourceLinks,
		ReleaseStatus:       a.ReleaseStatus,
		SunsetDate:          a.SunsetDate,
		Successors:          a.Successors,
		ChangeLogEntries:    a.ChangeLogEntries,
		Labels:              a.Labels,
		Visibility:          a.Visibility,
		Disabled:            a.Disabled,
		PartOfProducts:      a.PartOfProducts,
		LineOfBusiness:      a.LineOfBusiness,
		Industry:            a.Industry,
		Extensible:          a.Extensible,
		Version:             a.VersionInput.ToVersion(),
		Hierarchy:           a.Hierarchy,
		SupportedUseCases:   a.SupportedUseCases,
		DocumentationLabels: a.DocumentationLabels,
		ResourceHash:        hash,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType.IsTenantIgnorable() {
		api.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		api.ApplicationID = &resourceID
	}

	return api
}
