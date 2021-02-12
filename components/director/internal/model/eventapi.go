package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type EventDefinition struct {
	Tenant              string
	BundleID            *string
	PackageID           *string
	Name                string
	Description         *string
	Group               *string
	OrdID               *string
	ShortDescription    *string
	SystemInstanceAware *bool
	ChangeLogEntries    json.RawMessage
	Links               json.RawMessage
	Tags                json.RawMessage
	Countries           json.RawMessage
	ReleaseStatus       *string
	SunsetDate          *string
	Successor           *string
	Labels              json.RawMessage
	Visibility          *string
	Disabled            *bool
	PartOfProducts      json.RawMessage
	LineOfBusiness      json.RawMessage
	Industry            json.RawMessage

	Version *Version
	*BaseEntity
}

func (_ *EventDefinition) GetType() resource.Type {
	return resource.EventDefinition
}

type EventDefinitionPage struct {
	Data       []*EventDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventDefinitionPage) IsPageable() {}

type EventDefinitionInput struct {
	Tenant              string
	BundleID            *string
	PackageID           *string
	Name                string
	Description         *string
	Group               *string
	OrdID               *string
	ShortDescription    *string
	SystemInstanceAware *bool
	ChangeLogEntries    json.RawMessage
	Links               json.RawMessage
	Tags                json.RawMessage
	Countries           json.RawMessage
	ReleaseStatus       *string
	SunsetDate          *string
	Successor           *string
	Labels              json.RawMessage
	Visibility          *string
	Disabled            *bool
	PartOfProducts      json.RawMessage
	LineOfBusiness      json.RawMessage
	Industry            json.RawMessage

	Version *VersionInput
}

func (e *EventDefinitionInput) ToEventDefinitionWithinBundle(id string, bndlID string, tenant string) *EventDefinition {
	return e.ToEventDefinition(id, &bndlID, nil, tenant)
}

func (e *EventDefinitionInput) ToEventDefinition(id string, bundleID *string, packageID *string, tenant string) *EventDefinition {
	if e == nil {
		return nil
	}

	return &EventDefinition{
		BundleID:            bundleID,
		PackageID:           packageID,
		Tenant:              tenant,
		Name:                e.Name,
		Description:         e.Description,
		Group:               e.Group,
		OrdID:               e.OrdID,
		ShortDescription:    e.ShortDescription,
		SystemInstanceAware: e.SystemInstanceAware,
		Tags:                e.Tags,
		Countries:           e.Countries,
		Links:               e.Links,
		ReleaseStatus:       e.ReleaseStatus,
		SunsetDate:          e.SunsetDate,
		Successor:           e.Successor,
		ChangeLogEntries:    e.ChangeLogEntries,
		Labels:              e.Labels,
		Visibility:          e.Visibility,
		Disabled:            e.Disabled,
		PartOfProducts:      e.PartOfProducts,
		LineOfBusiness:      e.LineOfBusiness,
		Industry:            e.Industry,
		Version:             e.Version.ToVersion(),
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
