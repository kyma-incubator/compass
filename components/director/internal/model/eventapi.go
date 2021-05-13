package model

import (
	"encoding/json"
	"regexp"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type EventDefinition struct {
	Tenant              string
	ApplicationID       string
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
	Extensible          json.RawMessage

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
	OrdPackageID             *string                       `json:"partOfPackage"`
	Name                     string                        `json:"title"`
	Description              *string                       `json:"description"`
	Group                    *string                       `json:",omitempty"`
	OrdID                    *string                       `json:"ordId"`
	ShortDescription         *string                       `json:"shortDescription"`
	SystemInstanceAware      *bool                         `json:"systemInstanceAware"`
	ChangeLogEntries         json.RawMessage               `json:"changelogEntries"`
	Links                    json.RawMessage               `json:"links"`
	Tags                     json.RawMessage               `json:"tags"`
	Countries                json.RawMessage               `json:"countries"`
	ReleaseStatus            *string                       `json:"releaseStatus"`
	SunsetDate               *string                       `json:"sunsetDate"`
	Successor                *string                       `json:"successor"`
	Labels                   json.RawMessage               `json:"labels"`
	Visibility               *string                       `json:"visibility"`
	Disabled                 *bool                         `json:"disabled"`
	PartOfProducts           json.RawMessage               `json:"partOfProducts"`
	LineOfBusiness           json.RawMessage               `json:"lineOfBusiness"`
	Industry                 json.RawMessage               `json:"industry"`
	Extensible               json.RawMessage               `json:"extensible"`
	ResourceDefinitions      []*EventResourceDefinition    `json:"resourceDefinitions"`
	PartOfConsumptionBundles []*ConsumptionBundleReference `json:"partOfConsumptionBundles"`

	*VersionInput
}

type EventResourceDefinition struct { // This is the place from where the specification for this API is fetched
	Type           EventSpecType    `json:"type"`
	CustomType     string           `json:"customType"`
	MediaType      SpecFormat       `json:"mediaType"`
	URL            string           `json:"url"`
	AccessStrategy []AccessStrategy `json:"accessStrategies"`
}

func (rd *EventResourceDefinition) Validate() error {
	const CustomTypeRegex = "^([a-z0-9.]+):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	return validation.ValidateStruct(rd,
		validation.Field(&rd.Type, validation.Required, validation.In(EventSpecTypeAsyncAPIV2, EventSpecTypeCustom), validation.When(rd.CustomType != "", validation.In(EventSpecTypeCustom))),
		validation.Field(&rd.CustomType, validation.When(rd.CustomType != "", validation.Match(regexp.MustCompile(CustomTypeRegex)))),
		validation.Field(&rd.MediaType, validation.Required, validation.In(SpecFormatApplicationJSON, SpecFormatTextYAML, SpecFormatApplicationXML, SpecFormatPlainText, SpecFormatOctetStream)),
		validation.Field(&rd.URL, validation.Required, is.RequestURI),
		validation.Field(&rd.AccessStrategy, validation.Required),
	)
}

func (a *EventResourceDefinition) ToSpec() *SpecInput {
	specType := EventSpecType(a.Type)
	return &SpecInput{
		Format:     SpecFormat(a.MediaType),
		EventType:  &specType,
		CustomType: &a.CustomType,
		FetchRequest: &FetchRequestInput{ // TODO: Convert AccessStrategies to FetchRequest Auths once ORD defines them
			URL:  a.URL,
			Auth: nil, // Currently only open AccessStrategy is defined by ORD, which means no auth
		},
	}
}

func (e *EventDefinitionInput) ToEventDefinitionWithinBundle(id, appID, bndlID, tenant string) *EventDefinition {
	return e.ToEventDefinition(id, appID, nil, tenant)
}

func (e *EventDefinitionInput) ToEventDefinition(id, appID string, packageID *string, tenant string) *EventDefinition {
	if e == nil {
		return nil
	}

	return &EventDefinition{
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
		Extensible:          e.Extensible,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
