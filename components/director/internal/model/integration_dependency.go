package model

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"strconv"
)

// IntegrationDependency missing godoc
type IntegrationDependency struct {
	ApplicationID                  *string
	ApplicationTemplateVersionID   *string
	OrdID                          *string
	LocalTenantID                  *string
	CorrelationIDs                 json.RawMessage
	Name                           string
	ShortDescription               *string
	Description                    *string
	PackageID                      *string
	Version                        *Version
	LastUpdate                     *string
	Visibility                     string
	ReleaseStatus                  *string
	SunsetDate                     *string
	Successors                     json.RawMessage
	Mandatory                      bool
	Aspects                        json.RawMessage
	RelatedIntegrationDependencies json.RawMessage
	Links                          json.RawMessage
	Tags                           json.RawMessage
	Labels                         json.RawMessage
	DocumentationLabels            json.RawMessage
	ResourceHash                   *string
	*BaseEntity
}

// GetType missing godoc
func (*IntegrationDependency) GetType() resource.Type {
	return resource.IntegrationDependency
}

// IntegrationDependencyInput missing godoc
type IntegrationDependencyInput struct {
	OrdID                          *string         `json:"ordId"`
	LocalTenantID                  *string         `json:"localTenantId"`
	CorrelationIDs                 json.RawMessage `json:"correlationIds,omitempty"`
	Name                           string          `json:"title"`
	ShortDescription               *string         `json:"shortDescription"`
	Description                    *string         `json:"description"`
	OrdPackageID                   *string         `json:"partOfPackage"`
	LastUpdate                     *string         `json:"lastUpdate"`
	Visibility                     string          `json:"visibility"`
	ReleaseStatus                  *string         `json:"releaseStatus"`
	SunsetDate                     *string         `json:"sunsetDate"`
	Successors                     json.RawMessage `json:"successors,omitempty"`
	Mandatory                      bool            `json:"mandatory"`
	Aspects                        json.RawMessage `json:"aspects"`
	RelatedIntegrationDependencies json.RawMessage `json:"relatedIntegrationDependencies"`
	Links                          json.RawMessage `json:"links"`
	Tags                           json.RawMessage `json:"tags"`
	Labels                         json.RawMessage `json:"labels"`
	DocumentationLabels            json.RawMessage `json:"documentationLabels"`
	*VersionInput                  `hash:"ignore"`
}

// ToIntegrationDependency missing godoc
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
		Name:                           i.Name,
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
		Aspects:                        i.Aspects,
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
