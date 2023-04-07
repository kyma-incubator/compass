package model

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Bundle missing godoc
type Bundle struct {
	ApplicationID                  string
	Name                           string
	Description                    *string
	Version                        *string
	ResourceHash                   *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *Auth
	OrdID                          *string
	LocalTenantID                  *string
	ShortDescription               *string
	Links                          json.RawMessage
	Labels                         json.RawMessage
	CredentialExchangeStrategies   json.RawMessage
	CorrelationIDs                 json.RawMessage
	Tags                           json.RawMessage
	DocumentationLabels            json.RawMessage
	*BaseEntity
}

// GetType missing godoc
func (*Bundle) GetType() resource.Type {
	return resource.Bundle
}

// SetFromUpdateInput missing godoc
func (bndl *Bundle) SetFromUpdateInput(update BundleUpdateInput, bndlHash uint64) {
	var hash *string
	if bndlHash != 0 {
		hash = str.Ptr(strconv.FormatUint(bndlHash, 10))
	}

	bndl.Name = update.Name
	bndl.Description = update.Description
	bndl.Version = update.Version
	bndl.ResourceHash = hash
	bndl.InstanceAuthRequestInputSchema = update.InstanceAuthRequestInputSchema
	bndl.DefaultInstanceAuth = update.DefaultInstanceAuth.ToAuth()
	bndl.OrdID = update.OrdID
	bndl.LocalTenantID = update.LocalTenantID
	bndl.ShortDescription = update.ShortDescription
	bndl.Links = update.Links
	bndl.Labels = update.Labels
	bndl.CredentialExchangeStrategies = update.CredentialExchangeStrategies
	bndl.CorrelationIDs = update.CorrelationIDs
	bndl.Tags = update.Tags
	bndl.DocumentationLabels = update.DocumentationLabels
}

// BundleCreateInput missing godoc
type BundleCreateInput struct {
	Name                           string                  `json:"title"`
	Description                    *string                 `json:"description"`
	Version                        *string                 `json:"version" hash:"ignore"`
	InstanceAuthRequestInputSchema *string                 `json:",omitempty"`
	DefaultInstanceAuth            *AuthInput              `json:",omitempty"`
	OrdID                          *string                 `json:"ordId"`
	LocalTenantID                  *string                 `json:"localTenantId"`
	ShortDescription               *string                 `json:"shortDescription"`
	Links                          json.RawMessage         `json:"links"`
	Labels                         json.RawMessage         `json:"labels"`
	CredentialExchangeStrategies   json.RawMessage         `json:"credentialExchangeStrategies"`
	APIDefinitions                 []*APIDefinitionInput   `json:",omitempty"`
	APISpecs                       []*SpecInput            `json:",omitempty"`
	EventDefinitions               []*EventDefinitionInput `json:",omitempty"`
	EventSpecs                     []*SpecInput            `json:",omitempty"`
	Documents                      []*DocumentInput        `json:",omitempty"`
	CorrelationIDs                 json.RawMessage         `json:"correlationIds"`
	Tags                           json.RawMessage         `json:"tags"`
	DocumentationLabels            json.RawMessage         `json:"documentationLabels"`
}

// BundleUpdateInput missing godoc
type BundleUpdateInput struct {
	Name                           string
	Description                    *string
	Version                        *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *AuthInput
	OrdID                          *string
	LocalTenantID                  *string
	ShortDescription               *string
	Links                          json.RawMessage
	Labels                         json.RawMessage
	CredentialExchangeStrategies   json.RawMessage
	CorrelationIDs                 json.RawMessage
	Tags                           json.RawMessage
	DocumentationLabels            json.RawMessage
}

// BundlePage missing godoc
type BundlePage struct {
	Data       []*Bundle
	PageInfo   *pagination.Page
	TotalCount int
}

// IsPageable missing godoc
func (BundlePage) IsPageable() {}

// ToBundle missing godoc
func (i *BundleCreateInput) ToBundle(id, applicationID string, bndlHash uint64) *Bundle {
	if i == nil {
		return nil
	}

	var hash *string
	if bndlHash != 0 {
		hash = str.Ptr(strconv.FormatUint(bndlHash, 10))
	}

	return &Bundle{
		ApplicationID:                  applicationID,
		Name:                           i.Name,
		Description:                    i.Description,
		Version:                        i.Version,
		ResourceHash:                   hash,
		InstanceAuthRequestInputSchema: i.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            i.DefaultInstanceAuth.ToAuth(),
		OrdID:                          i.OrdID,
		LocalTenantID:                  i.LocalTenantID,
		ShortDescription:               i.ShortDescription,
		Links:                          i.Links,
		Labels:                         i.Labels,
		CredentialExchangeStrategies:   i.CredentialExchangeStrategies,
		CorrelationIDs:                 i.CorrelationIDs,
		Tags:                           i.Tags,
		DocumentationLabels:            i.DocumentationLabels,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
