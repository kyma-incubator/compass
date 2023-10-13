package model

import (
	"encoding/json"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

// EntityType missing godoc
type EntityType struct {
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	OrdID                        string
	LocalID                      string
	CorrelationIDs               json.RawMessage
	Level                        string
	Title                        string
	ShortDescription             *string
	Description                  *string
	SystemInstanceAware          *bool
	ChangeLogEntries             json.RawMessage
	OrdPackageID                 string
	Visibility                   string
	Links                        json.RawMessage
	PartOfProducts               json.RawMessage
	PolicyLevel                  *string
	CustomPolicyLevel            *string
	ReleaseStatus                string
	SunsetDate                   *string
	Successors                   json.RawMessage
	Extensible                   json.RawMessage
	Tags                         json.RawMessage
	Labels                       json.RawMessage
	DocumentationLabels          json.RawMessage
	ResourceHash                 *string
	Version                      *Version
	*BaseEntity
}

// GetType missing godoc
func (*EntityType) GetType() resource.Type {
	return resource.EntityType
}

// EntityTypePage missing godoc
type EntityTypePage struct {
	Data       []*EntityType
	PageInfo   *pagination.Page
	TotalCount int
}

// IsPageable missing godoc
func (EntityTypePage) IsPageable() {}

// EntityTypeInput missing godoc
type EntityTypeInput struct {
	OrdID               string          `json:"ordId"`
	LocalID             string          `json:"localId"`
	CorrelationIDs      json.RawMessage `json:"correlationIds,omitempty"`
	Level               string          `json:"level"`
	Title               string          `json:"title"`
	ShortDescription    *string         `json:"shortDescription,omitempty"`
	Description         *string         `json:"description,omitempty"`
	SystemInstanceAware *bool           `json:"systemInstanceAware"`
	ChangeLogEntries    json.RawMessage `json:"changelogEntries,omitempty"`
	OrdPackageID        string          `json:"partOfPackage"`
	Visibility          string          `json:"visibility"`
	Links               json.RawMessage `json:"links,omitempty"`
	PartOfProducts      json.RawMessage `json:"partOfProducts,omitempty"`
	PolicyLevel         *string         `json:"policyLevel,omitempty"`
	CustomPolicyLevel   *string         `json:"customPolicyLevel,omitempty"`
	ReleaseStatus       string          `json:"releaseStatus"`
	SunsetDate          *string         `json:"sunsetDate,omitempty"`
	Successors          json.RawMessage `json:"successors,omitempty"`
	Extensible          json.RawMessage `json:"extensible,omitempty"`
	Tags                json.RawMessage `json:"tags,omitempty"`
	Labels              json.RawMessage `json:"labels,omitempty"`
	DocumentationLabels json.RawMessage `json:"documentationLabels,omitempty"`

	*VersionInput `hash:"ignore"`
}

// ToEntityType missing godoc
func (i *EntityTypeInput) ToEntityType(id string, resourceType resource.Type, resourceID string, entityTypeHash uint64) *EntityType {
	if i == nil {
		return nil
	}

	var hash *string
	if entityTypeHash != 0 {
		hash = str.Ptr(strconv.FormatUint(entityTypeHash, 10))
	}

	entityType := &EntityType{
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
		OrdID:               i.OrdID,
		LocalID:             i.LocalID,
		CorrelationIDs:      i.CorrelationIDs,
		Level:               i.Level,
		Title:               i.Title,
		ShortDescription:    i.ShortDescription,
		Description:         i.Description,
		SystemInstanceAware: i.SystemInstanceAware,
		ChangeLogEntries:    i.ChangeLogEntries,
		OrdPackageID:        i.OrdPackageID,
		Visibility:          i.Visibility,
		Links:               i.Links,
		PartOfProducts:      i.PartOfProducts,
		PolicyLevel:         i.PolicyLevel,
		CustomPolicyLevel:   i.CustomPolicyLevel,
		ReleaseStatus:       i.ReleaseStatus,
		SunsetDate:          i.SunsetDate,
		Successors:          i.Successors,
		Extensible:          i.Extensible,
		Tags:                i.Tags,
		Labels:              i.Labels,
		DocumentationLabels: i.DocumentationLabels,
		Version:             i.VersionInput.ToVersion(),
		ResourceHash:        hash,
	}

	if resourceType.IsTenantIgnorable() {
		entityType.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		entityType.ApplicationID = &resourceID
	}

	return entityType
}

// SetFromUpdateInput missing godoc
func (entityType *EntityType) SetFromUpdateInput(update EntityTypeInput, entityTypeHash uint64) {
	var hash *string
	if entityTypeHash != 0 {
		hash = str.Ptr(strconv.FormatUint(entityTypeHash, 10))
	}
	entityType.OrdID = update.OrdID
	entityType.LocalID = update.LocalID
	entityType.CorrelationIDs = update.CorrelationIDs
	entityType.Level = update.Level
	entityType.Title = update.Title
	entityType.ShortDescription = update.ShortDescription
	entityType.Description = update.Description
	entityType.SystemInstanceAware = update.SystemInstanceAware
	entityType.ChangeLogEntries = update.ChangeLogEntries
	entityType.OrdPackageID = update.OrdPackageID
	entityType.Visibility = update.Visibility
	entityType.Links = update.Links
	entityType.PartOfProducts = update.PartOfProducts
	entityType.PolicyLevel = update.PolicyLevel
	entityType.CustomPolicyLevel = update.CustomPolicyLevel
	entityType.ReleaseStatus = update.ReleaseStatus
	entityType.SunsetDate = update.SunsetDate
	entityType.Successors = update.Successors
	entityType.Extensible = update.Extensible
	entityType.Tags = update.Tags
	entityType.Labels = update.Labels
	entityType.DocumentationLabels = update.DocumentationLabels
	entityType.Version = update.VersionInput.ToVersion()
	entityType.ResourceHash = hash
}
