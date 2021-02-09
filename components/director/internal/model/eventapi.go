package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type EventDefinition struct {
	ID                  string
	Tenant              string
	ApplicationID       string
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
}

type EventDefinitionPage struct {
	Data       []*EventDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

func (EventDefinitionPage) IsPageable() {}

type EventDefinitionInput struct {
	OrdBundleID         *string `json:"partOfConsumptionBundle"`
	OrdPackageID        *string `json:"partOfPackage"`
	Name                string  `json:"title"`
	Description         *string `json:"description"`
	Group               *string
	OrdID               *string         `json:"ordId"`
	ShortDescription    *string         `json:"shortDescription"`
	SystemInstanceAware *bool           `json:"systemInstanceAware"`
	ChangeLogEntries    json.RawMessage `json:"changelogEntries"`
	Links               json.RawMessage `json:"links"`
	Tags                json.RawMessage `json:"tags"`
	Countries           json.RawMessage `json:"countries"`
	ReleaseStatus       *string         `json:"releaseStatus"`
	SunsetDate          *string         `json:"sunsetDate"`
	Successor           *string         `json:"successor"`
	Labels              json.RawMessage `json:"labels"`
	Visibility          *string         `json:"visibility"`
	Disabled            *bool           `json:"disabled"`
	PartOfProducts      json.RawMessage `json:"partOfProducts"`
	LineOfBusiness      json.RawMessage `json:"lineOfBusiness"`
	Industry            json.RawMessage `json:"industry"`

	EventResourceDefinition []EventResourceDefinition `json:"resourceDefinitions"`

	*VersionInput
}

type EventResourceDefinition struct { // This is the place from where the specification for this API is fetched
	Type           string           `json:"type"`
	CustomType     string           `json:"customType"`
	MediaType      string           `json:"mediaType"`
	URL            string           `json:"url"`
	AccessStrategy []AccessStrategy `json:"accessStrategies"`
}

func (e *EventDefinitionInput) ToEventDefinitionWithinBundle(id, appID, bndlID, tenant string) *EventDefinition {
	return e.ToEventDefinition(id, appID, &bndlID, nil, tenant)
}

func (e *EventDefinitionInput) ToEventDefinition(id, appID string, bundleID *string, packageID *string, tenant string) *EventDefinition {
	if e == nil {
		return nil
	}

	return &EventDefinition{
		ID:                  id,
		BundleID:            bundleID,
		ApplicationID:       appID,
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
		Version:             e.VersionInput.ToVersion(),
	}
}
