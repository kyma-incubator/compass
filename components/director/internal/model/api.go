package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
	ApplicationID       string
	BundleID            *string
	PackageID           *string
	Tenant              string
	Name                string
	Description         *string
	TargetURL           string
	Group               *string //  group allows you to find the same API but in different version
	OrdID               *string
	ShortDescription    *string
	SystemInstanceAware *bool
	ApiProtocol         *string
	Tags                json.RawMessage
	Countries           json.RawMessage
	Links               json.RawMessage
	APIResourceLinks    json.RawMessage
	ReleaseStatus       *string
	SunsetDate          *string
	Successor           *string
	ChangeLogEntries    json.RawMessage
	Labels              json.RawMessage
	Visibility          *string
	Disabled            *bool
	PartOfProducts      json.RawMessage
	LineOfBusiness      json.RawMessage
	Industry            json.RawMessage
	Version             *Version
	*BaseEntity
}

func (_ *APIDefinition) GetType() resource.Type {
	return resource.API
}

type APIDefinitionInput struct {
	OrdBundleID         *string `json:"partOfConsumptionBundle"`
	OrdPackageID        *string `json:"partOfPackage"`
	Tenant              string
	Name                string          `json:"title"`
	Description         *string         `json:"description"`
	TargetURL           string          `json:"entryPoint"`
	Group               *string         //  group allows you to find the same API but in different version
	OrdID               *string         `json:"ordId"`
	ShortDescription    *string         `json:"shortDescription"`
	SystemInstanceAware *bool           `json:"systemInstanceAware"`
	ApiProtocol         *string         `json:"apiProtocol"`
	Tags                json.RawMessage `json:"tags"`
	Countries           json.RawMessage `json:"countries"`
	Links               json.RawMessage `json:"links"`
	APIResourceLinks    json.RawMessage `json:"apiResourceLinks"`
	ReleaseStatus       *string         `json:"releaseStatus"`
	SunsetDate          *string         `json:"sunsetDate"`
	Successor           *string         `json:"successor"`
	ChangeLogEntries    json.RawMessage `json:"changelogEntries"`
	Labels              json.RawMessage `json:"labels"`
	Visibility          *string         `json:"visibility"`
	Disabled            *bool           `json:"disabled"`
	PartOfProducts      json.RawMessage `json:"partOfProducts"`
	LineOfBusiness      json.RawMessage `json:"lineOfBusiness"`
	Industry            json.RawMessage `json:"industry"`

	ResourceDefinitions []APIResourceDefinition `json:"resourceDefinitions"`

	*VersionInput
}

type APIResourceDefinition struct { // This is the place from where the specification for this API is fetched
	Type           string           `json:"type"`
	CustomType     string           `json:"customType"`
	MediaType      string           `json:"mediaType"`
	URL            string           `json:"url"`
	AccessStrategy []AccessStrategy `json:"accessStrategies"`
}

func (a APIResourceDefinition) ToSpec() *SpecInput {
	specType := APISpecType(a.Type)
	return &SpecInput{
		Format:     SpecFormat(a.MediaType),
		APIType:    &specType,
		CustomType: &a.CustomType,
		FetchRequest: &FetchRequestInput{ // TODO: Convert AccessStrategies to FetchRequestAuths once ORD defines them
			URL:  a.URL,
			Auth: nil, // Currently only open AccessStrategy is defined by ORD, which means no auth
		},
	}
}

type AccessStrategy struct {
	Type              string `json:"type"`
	CustomType        string `json:"customType"`
	CustomDescription string `json:"customDescription"`
}

type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (APIDefinitionPage) IsPageable() {}

func (a *APIDefinitionInput) ToAPIDefinitionWithinBundle(id, appID, bundleID, tenant string) *APIDefinition {
	return a.ToAPIDefinition(id, appID, &bundleID, nil, tenant)
}

func (a *APIDefinitionInput) ToAPIDefinition(id, appID string, bundleID *string, packageID *string, tenant string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
		ApplicationID:       appID,
		BundleID:            bundleID,
		PackageID:           packageID,
		Tenant:              tenant,
		Name:                a.Name,
		Description:         a.Description,
		TargetURL:           a.TargetURL,
		Group:               a.Group,
		OrdID:               a.OrdID,
		ShortDescription:    a.ShortDescription,
		SystemInstanceAware: a.SystemInstanceAware,
		ApiProtocol:         a.ApiProtocol,
		Tags:                a.Tags,
		Countries:           a.Countries,
		Links:               a.Links,
		APIResourceLinks:    a.APIResourceLinks,
		ReleaseStatus:       a.ReleaseStatus,
		SunsetDate:          a.SunsetDate,
		Successor:           a.Successor,
		ChangeLogEntries:    a.ChangeLogEntries,
		Labels:              a.Labels,
		Visibility:          a.Visibility,
		Disabled:            a.Disabled,
		PartOfProducts:      a.PartOfProducts,
		LineOfBusiness:      a.LineOfBusiness,
		Industry:            a.Industry,
		Version:             a.VersionInput.ToVersion(),
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
