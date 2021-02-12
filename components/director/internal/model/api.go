package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type APIDefinition struct {
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
	Version             *VersionInput
}

type APIDefinitionPage struct {
	Data       []*APIDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (APIDefinitionPage) IsPageable() {}

func (a *APIDefinitionInput) ToAPIDefinitionWithinBundle(id string, bundleID string, tenant string) *APIDefinition {
	return a.ToAPIDefinition(id, &bundleID, nil, tenant)
}

func (a *APIDefinitionInput) ToAPIDefinition(id string, bundleID *string, packageID *string, tenant string) *APIDefinition {
	if a == nil {
		return nil
	}

	return &APIDefinition{
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
		Version:             a.Version.ToVersion(),
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
