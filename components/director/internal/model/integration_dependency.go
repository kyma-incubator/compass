package model

import (
	"encoding/json"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

// IntegrationDependency represent structure for IntegrationDependency
type IntegrationDependency struct {
	ApplicationID                  *string
	ApplicationTemplateVersionID   *string
	OrdID                          *string
	LocalTenantID                  *string
	CorrelationIDs                 json.RawMessage
	Title                          string
	ShortDescription               *string
	Description                    *string
	PackageID                      *string
	Version                        *Version
	LastUpdate                     *string
	Visibility                     string
	ReleaseStatus                  *string
	SunsetDate                     *string
	Successors                     json.RawMessage
	Mandatory                      *bool
	RelatedIntegrationDependencies json.RawMessage
	Links                          json.RawMessage
	Tags                           json.RawMessage
	Labels                         json.RawMessage
	DocumentationLabels            json.RawMessage
	ResourceHash                   *string
	*BaseEntity
}

// GetType returns Type integrationDependency
func (*IntegrationDependency) GetType() resource.Type {
	return resource.IntegrationDependency
}

// IntegrationDependencyInput is an input for creating a new IntegrationDependency
type IntegrationDependencyInput struct {
	OrdID                          *string         `json:"ordId"`
	LocalTenantID                  *string         `json:"localTenantId,omitempty"`
	CorrelationIDs                 json.RawMessage `json:"correlationIds,omitempty"`
	Title                          string          `json:"title"`
	ShortDescription               *string         `json:"shortDescription,omitempty"`
	Description                    *string         `json:"description,omitempty"`
	OrdPackageID                   *string         `json:"partOfPackage"`
	LastUpdate                     *string         `json:"lastUpdate,omitempty"`
	Visibility                     string          `json:"visibility"`
	ReleaseStatus                  *string         `json:"releaseStatus"`
	SunsetDate                     *string         `json:"sunsetDate,omitempty"`
	Successors                     json.RawMessage `json:"successors,omitempty"`
	Mandatory                      *bool           `json:"mandatory"`
	Aspects                        []*AspectInput  `json:"aspects,omitempty"`
	RelatedIntegrationDependencies json.RawMessage `json:"relatedIntegrationDependencies,omitempty"`
	Links                          json.RawMessage `json:"links,omitempty"`
	Tags                           json.RawMessage `json:"tags,omitempty"`
	Labels                         json.RawMessage `json:"labels,omitempty"`
	DocumentationLabels            json.RawMessage `json:"documentationLabels,omitempty"`
	*VersionInput                  `hash:"ignore"`
}

// ToIntegrationDependency converts IntegrationDependencyInput to IntegrationDependency
func (i *IntegrationDependencyInput) ToIntegrationDependency(id string, resourceType resource.Type, resourceID string, packageID *string, integrationDependencyHash uint64) *IntegrationDependency {
	if i == nil {
		return nil
	}

	var hash *string
	if integrationDependencyHash != 0 {
		hash = str.Ptr(strconv.FormatUint(integrationDependencyHash, 10))
	}

	integrationDependency := &IntegrationDependency{
		OrdID:                          i.OrdID,
		LocalTenantID:                  i.LocalTenantID,
		CorrelationIDs:                 i.CorrelationIDs,
		Title:                          i.Title,
		ShortDescription:               i.ShortDescription,
		Description:                    i.Description,
		PackageID:                      packageID,
		Version:                        i.VersionInput.ToVersion(),
		LastUpdate:                     i.LastUpdate,
		Visibility:                     i.Visibility,
		ReleaseStatus:                  i.ReleaseStatus,
		SunsetDate:                     i.SunsetDate,
		Successors:                     i.Successors,
		Mandatory:                      i.Mandatory,
		RelatedIntegrationDependencies: i.RelatedIntegrationDependencies,
		Links:                          i.Links,
		Tags:                           i.Tags,
		Labels:                         i.Labels,
		DocumentationLabels:            i.DocumentationLabels,
		ResourceHash:                   hash,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType.IsTenantIgnorable() {
		integrationDependency.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		integrationDependency.ApplicationID = &resourceID
	}

	return integrationDependency
}
