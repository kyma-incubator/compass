package model

import (
	"encoding/json"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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
	SupportedUseCases                       json.RawMessage
	DocumentationLabels                     json.RawMessage
	CorrelationIDs                          json.RawMessage
	Direction                               *string
	LastUpdate                              *string
	DeprecationDate                         *string
	Responsible                             *string
	Usage                                   *string
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
	LocalTenantID                           *string                       `json:"localId"`
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
	SupportedUseCases                       json.RawMessage               `json:"supportedUseCases"`
	EntityTypeMappings                      []*EntityTypeMappingInput     `json:"entityTypeMappings"`
	DocumentationLabels                     json.RawMessage               `json:"documentationLabels"`
	CorrelationIDs                          json.RawMessage               `json:"correlationIds,omitempty"`
	Direction                               *string                       `json:"direction"`
	LastUpdate                              *string                       `json:"lastUpdate" hash:"ignore"`
	DeprecationDate                         *string                       `json:"deprecationDate"`
	Responsible                             *string                       `json:"responsible"`
	Usage                                   *string                       `json:"usage"`
	*VersionInput                           `hash:"ignore"`
}

// APIResourceDefinition missing godoc
type APIResourceDefinition struct { // This is the place from where the specification for this API is fetched
	Type           APISpecType                     `json:"type"`
	CustomType     string                          `json:"customType"`
	MediaType      SpecFormat                      `json:"mediaType"`
	URL            string                          `json:"url"`
	AccessStrategy accessstrategy.AccessStrategies `json:"accessStrategies"`
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
		SupportedUseCases:   a.SupportedUseCases,
		DocumentationLabels: a.DocumentationLabels,
		CorrelationIDs:      a.CorrelationIDs,
		Direction:           a.Direction,
		LastUpdate:          a.LastUpdate,
		DeprecationDate:     a.DeprecationDate,
		Responsible:         a.Responsible,
		Usage:               a.Usage,
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
