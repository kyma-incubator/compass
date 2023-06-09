package model

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// EventDefinition missing godoc
type EventDefinition struct {
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	PackageID                    *string
	Name                         string
	Description                  *string
	Group                        *string
	OrdID                        *string
	LocalTenantID                *string
	ShortDescription             *string
	SystemInstanceAware          *bool
	PolicyLevel                  *string
	CustomPolicyLevel            *string
	ChangeLogEntries             json.RawMessage
	Links                        json.RawMessage
	Tags                         json.RawMessage
	Countries                    json.RawMessage
	ReleaseStatus                *string
	SunsetDate                   *string
	Successors                   json.RawMessage
	Labels                       json.RawMessage
	Visibility                   *string
	Disabled                     *bool
	PartOfProducts               json.RawMessage
	LineOfBusiness               json.RawMessage
	Industry                     json.RawMessage
	Extensible                   json.RawMessage
	ResourceHash                 *string
	Version                      *Version
	Hierarchy                    json.RawMessage
	DocumentationLabels          json.RawMessage
	*BaseEntity
}

// GetType missing godoc
func (*EventDefinition) GetType() resource.Type {
	return resource.EventDefinition
}

// EventDefinitionPage missing godoc
type EventDefinitionPage struct {
	Data       []*EventDefinition
	PageInfo   *pagination.Page
	TotalCount int
}

// IsPageable missing godoc
func (EventDefinitionPage) IsPageable() {}

// EventDefinitionInput missing godoc
type EventDefinitionInput struct {
	OrdPackageID             *string                       `json:"partOfPackage"`
	Name                     string                        `json:"title"`
	Description              *string                       `json:"description"`
	Group                    *string                       `json:",omitempty"`
	OrdID                    *string                       `json:"ordId"`
	LocalTenantID            *string                       `json:"localTenantId"`
	ShortDescription         *string                       `json:"shortDescription"`
	SystemInstanceAware      *bool                         `json:"systemInstanceAware"`
	PolicyLevel              *string                       `json:"policyLevel"`
	CustomPolicyLevel        *string                       `json:"customPolicyLevel"`
	ChangeLogEntries         json.RawMessage               `json:"changelogEntries"`
	Links                    json.RawMessage               `json:"links"`
	Tags                     json.RawMessage               `json:"tags"`
	Countries                json.RawMessage               `json:"countries"`
	ReleaseStatus            *string                       `json:"releaseStatus"`
	SunsetDate               *string                       `json:"sunsetDate"`
	Successors               json.RawMessage               `json:"successors,omitempty"`
	Labels                   json.RawMessage               `json:"labels"`
	Visibility               *string                       `json:"visibility"`
	Disabled                 *bool                         `json:"disabled"`
	PartOfProducts           json.RawMessage               `json:"partOfProducts"`
	LineOfBusiness           json.RawMessage               `json:"lineOfBusiness"`
	Industry                 json.RawMessage               `json:"industry"`
	Extensible               json.RawMessage               `json:"extensible"`
	ResourceDefinitions      []*EventResourceDefinition    `json:"resourceDefinitions"`
	PartOfConsumptionBundles []*ConsumptionBundleReference `json:"partOfConsumptionBundles"`
	DefaultConsumptionBundle *string                       `json:"defaultConsumptionBundle"`
	Hierarchy                json.RawMessage               `json:"hierarchy"`
	DocumentationLabels      json.RawMessage               `json:"documentationLabels"`

	*VersionInput `hash:"ignore"`
}

// EventResourceDefinition missing godoc
type EventResourceDefinition struct { // This is the place from where the specification for this API is fetched
	Type           EventSpecType                   `json:"type"`
	CustomType     string                          `json:"customType"`
	MediaType      SpecFormat                      `json:"mediaType"`
	URL            string                          `json:"url"`
	AccessStrategy accessstrategy.AccessStrategies `json:"accessStrategies"`
}

// Validate missing godoc
func (rd *EventResourceDefinition) Validate() error {
	const CustomTypeRegex = "^([a-z0-9-]+(?:[.][a-z0-9-]+)*):([a-zA-Z0-9._\\-]+):v([0-9]+)$"
	return validation.ValidateStruct(rd,
		validation.Field(&rd.Type, validation.Required, validation.In(EventSpecTypeAsyncAPIV2, EventSpecTypeCustom), validation.When(rd.CustomType != "", validation.In(EventSpecTypeCustom))),
		validation.Field(&rd.CustomType, validation.When(rd.CustomType != "", validation.Match(regexp.MustCompile(CustomTypeRegex)))),
		validation.Field(&rd.MediaType, validation.Required, validation.In(SpecFormatApplicationJSON, SpecFormatTextYAML, SpecFormatApplicationXML, SpecFormatPlainText, SpecFormatOctetStream)),
		validation.Field(&rd.URL, validation.Required, is.RequestURI),
		validation.Field(&rd.AccessStrategy, validation.Required),
	)
}

// ToSpec missing godoc
func (rd *EventResourceDefinition) ToSpec() *SpecInput {
	var auth *AuthInput
	if as, ok := rd.AccessStrategy.GetSupported(); ok {
		asString := string(as)
		auth = &AuthInput{
			AccessStrategy: &asString,
		}
	}

	specType := rd.Type
	return &SpecInput{
		Format:     rd.MediaType,
		EventType:  &specType,
		CustomType: &rd.CustomType,
		FetchRequest: &FetchRequestInput{
			URL:  rd.URL,
			Auth: auth,
		},
	}
}

// ToEventDefinition missing godoc
func (e *EventDefinitionInput) ToEventDefinition(id string, resourceType resource.Type, resourceID string, packageID *string, eventHash uint64) *EventDefinition {
	if e == nil {
		return nil
	}

	var hash *string
	if eventHash != 0 {
		hash = str.Ptr(strconv.FormatUint(eventHash, 10))
	}

	event := &EventDefinition{
		PackageID:           packageID,
		Name:                e.Name,
		Description:         e.Description,
		Group:               e.Group,
		OrdID:               e.OrdID,
		LocalTenantID:       e.LocalTenantID,
		ShortDescription:    e.ShortDescription,
		SystemInstanceAware: e.SystemInstanceAware,
		PolicyLevel:         e.PolicyLevel,
		CustomPolicyLevel:   e.CustomPolicyLevel,
		Tags:                e.Tags,
		Countries:           e.Countries,
		Links:               e.Links,
		ReleaseStatus:       e.ReleaseStatus,
		SunsetDate:          e.SunsetDate,
		Successors:          e.Successors,
		ChangeLogEntries:    e.ChangeLogEntries,
		Labels:              e.Labels,
		Visibility:          e.Visibility,
		Disabled:            e.Disabled,
		PartOfProducts:      e.PartOfProducts,
		LineOfBusiness:      e.LineOfBusiness,
		Industry:            e.Industry,
		Version:             e.VersionInput.ToVersion(),
		Extensible:          e.Extensible,
		Hierarchy:           e.Hierarchy,
		DocumentationLabels: e.DocumentationLabels,
		ResourceHash:        hash,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType == resource.ApplicationTemplateVersion {
		event.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		event.ApplicationID = &resourceID
	}

	return event
}
