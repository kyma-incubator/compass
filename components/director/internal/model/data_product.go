package model

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type DataProduct struct {
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	OrdID                        *string
	LocalTenantID                *string
	CorrelationIDs               json.RawMessage
	Title                        string
	ShortDescription             *string
	Description                  *string
	PackageID                    *string
	Version                      *Version
	LastUpdate                   *string
	Visibility                   *string
	ReleaseStatus                *string
	Disabled                     *bool
	DeprecationDate              *string
	SunsetDate                   *string
	Successors                   json.RawMessage
	ChangeLogEntries             json.RawMessage
	Type                         string
	Category                     string
	EntityTypes                  json.RawMessage
	InputPorts                   json.RawMessage
	OutputPorts                  json.RawMessage
	Responsible                  *string
	DataProductLinks             json.RawMessage
	Links                        json.RawMessage
	Industry                     json.RawMessage
	LineOfBusiness               json.RawMessage
	Tags                         json.RawMessage
	Labels                       json.RawMessage
	DocumentationLabels          json.RawMessage
	PolicyLevel                  *string
	CustomPolicyLevel            *string
	SystemInstanceAware          *bool
	*BaseEntity
}

// GetType missing godoc
func (*DataProduct) GetType() resource.Type {
	return resource.DataProduct
}

type DataProductInput struct {
	OrdID               *string         `json:"ordID"`
	LocalTenantID       *string         `json:"localId,omitempty"`
	CorrelationIDs      json.RawMessage `json:"correlationIds,omitempty"`
	Title               string          `json:"title"`
	ShortDescription    *string         `json:"shortDescription,omitempty"`
	Description         *string         `json:"description"`
	OrdPackageID        *string         `json:"partOfPackage"`
	LastUpdate          *string         `json:"lastUpdate,omitempty"`
	Visibility          *string         `json:"visibility"`
	ReleaseStatus       *string         `json:"releaseStatus"`
	Disabled            *bool           `json:"disabled"`
	DeprecationDate     *string         `json:"deprecationDate"`
	SunsetDate          *string         `json:"sunsetDate,omitempty"`
	Successors          json.RawMessage `json:"successors,omitempty"`
	ChangeLogEntries    json.RawMessage `json:"changeLogEntries"`
	Type                string          `json:"type"`
	Category            string          `json:"category"`
	EntityTypes         json.RawMessage `json:"entityTypes"`
	InputPorts          json.RawMessage `json:"inputPorts"`
	OutputPorts         json.RawMessage `json:"outputPorts"`
	Responsible         *string         `json:"responsible"`
	DataProductLinks    json.RawMessage `json:"dataProductLinks"`
	Links               json.RawMessage `json:"links,omitempty"`
	Industry            json.RawMessage `json:"industry"`
	LineOfBusiness      json.RawMessage `json:"lineOfBusiness"`
	Tags                json.RawMessage `json:"tags,omitempty"`
	Labels              json.RawMessage `json:"labels,omitempty"`
	DocumentationLabels json.RawMessage `json:"documentationLabels,omitempty"`
	PolicyLevel         *string         `json:"policyLevel"`
	CustomPolicyLevel   *string         `json:"customPolicyLevel"`
	SystemInstanceAware *bool           `json:"systemInstanceAware"`
	*VersionInput       `hash:"ignore"`
}
