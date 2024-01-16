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
	LocalTenantID                string
	CorrelationIDs               json.RawMessage
	Level                        string
	Title                        string
	ShortDescription             *string
	Description                  *string
	SystemInstanceAware          *bool
	ChangeLogEntries             json.RawMessage
	PackageID                    string
	Visibility                   string
	Links                        json.RawMessage
	PartOfProducts               json.RawMessage
	LastUpdate                   *string
	PolicyLevel                  *string
	CustomPolicyLevel            *string
	ReleaseStatus                string
	SunsetDate                   *string
	DeprecationDate              *string
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
	LocalTenantID       string          `json:"localId"`
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
	LastUpdate          *string         `json:"lastUpdate,omitempty" hash:"ignore"`
	PolicyLevel         *string         `json:"policyLevel,omitempty"`
	CustomPolicyLevel   *string         `json:"customPolicyLevel,omitempty"`
	ReleaseStatus       string          `json:"releaseStatus"`
	SunsetDate          *string         `json:"sunsetDate,omitempty"`
	DeprecationDate     *string         `json:"deprecationDate,omitempty"`
	Successors          json.RawMessage `json:"successors,omitempty"`
	Extensible          json.RawMessage `json:"extensible,omitempty"`
	Tags                json.RawMessage `json:"tags,omitempty"`
	Labels              json.RawMessage `json:"labels,omitempty"`
	DocumentationLabels json.RawMessage `json:"documentationLabels,omitempty"`

	*VersionInput `hash:"ignore"`
}

// ToEntityType missing godoc
func (i *EntityTypeInput) ToEntityType(id string, resourceType resource.Type, resourceID string, packageID string, entityTypeHash uint64) *EntityType {
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
		LocalTenantID:       i.LocalTenantID,
		CorrelationIDs:      i.CorrelationIDs,
		Level:               i.Level,
		Title:               i.Title,
		ShortDescription:    i.ShortDescription,
		Description:         i.Description,
		SystemInstanceAware: i.SystemInstanceAware,
		ChangeLogEntries:    i.ChangeLogEntries,
		PackageID:           packageID,
		Visibility:          i.Visibility,
		Links:               i.Links,
		PartOfProducts:      i.PartOfProducts,
		LastUpdate:          i.LastUpdate,
		PolicyLevel:         i.PolicyLevel,
		CustomPolicyLevel:   i.CustomPolicyLevel,
		ReleaseStatus:       i.ReleaseStatus,
		SunsetDate:          i.SunsetDate,
		DeprecationDate:     i.DeprecationDate,
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
func (entityType *EntityType) SetFromUpdateInput(update EntityTypeInput, packageID string, entityTypeHash uint64) {
	var hash *string
	if entityTypeHash != 0 {
		hash = str.Ptr(strconv.FormatUint(entityTypeHash, 10))
	}
	entityType.OrdID = update.OrdID
	entityType.LocalTenantID = update.LocalTenantID
	entityType.CorrelationIDs = update.CorrelationIDs
	entityType.Level = update.Level
	entityType.Title = update.Title
	entityType.ShortDescription = update.ShortDescription
	entityType.Description = update.Description
	entityType.SystemInstanceAware = update.SystemInstanceAware
	entityType.ChangeLogEntries = update.ChangeLogEntries
	entityType.PackageID = packageID
	entityType.Visibility = update.Visibility
	entityType.Links = update.Links
	entityType.PartOfProducts = update.PartOfProducts
	entityType.LastUpdate = update.LastUpdate
	entityType.PolicyLevel = update.PolicyLevel
	entityType.CustomPolicyLevel = update.CustomPolicyLevel
	entityType.ReleaseStatus = update.ReleaseStatus
	entityType.SunsetDate = update.SunsetDate
	entityType.DeprecationDate = update.DeprecationDate
	entityType.Successors = update.Successors
	entityType.Extensible = update.Extensible
	entityType.Tags = update.Tags
	entityType.Labels = update.Labels
	entityType.DocumentationLabels = update.DocumentationLabels
	entityType.Version = update.VersionInput.ToVersion()
	entityType.ResourceHash = hash
}
