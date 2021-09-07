package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

type Bundle struct {
	TenantID                       string
	ApplicationID                  string
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *Auth
	OrdID                          *string
	ShortDescription               *string
	Links                          json.RawMessage
	Labels                         json.RawMessage
	CredentialExchangeStrategies   json.RawMessage
	*BaseEntity
}

func (*Bundle) GetType() resource.Type {
	return resource.Bundle
}

func (bndl *Bundle) SetFromUpdateInput(update BundleUpdateInput) {
	bndl.Name = update.Name
	bndl.Description = update.Description
	bndl.InstanceAuthRequestInputSchema = update.InstanceAuthRequestInputSchema
	bndl.DefaultInstanceAuth = update.DefaultInstanceAuth.ToAuth()
	bndl.OrdID = update.OrdID
	bndl.ShortDescription = update.ShortDescription
	bndl.Links = update.Links
	bndl.Labels = update.Labels
	bndl.CredentialExchangeStrategies = update.CredentialExchangeStrategies
}

type BundleCreateInput struct {
	Name                           string                  `json:"title"`
	Description                    *string                 `json:"description"`
	InstanceAuthRequestInputSchema *string                 `json:",omitempty"`
	DefaultInstanceAuth            *AuthInput              `json:",omitempty"`
	OrdID                          *string                 `json:"ordId"`
	ShortDescription               *string                 `json:"shortDescription"`
	Links                          json.RawMessage         `json:"links"`
	Labels                         json.RawMessage         `json:"labels"`
	CredentialExchangeStrategies   json.RawMessage         `json:"credentialExchangeStrategies"`
	APIDefinitions                 []*APIDefinitionInput   `json:",omitempty"`
	APISpecs                       []*SpecInput            `json:",omitempty"`
	EventDefinitions               []*EventDefinitionInput `json:",omitempty"`
	EventSpecs                     []*SpecInput            `json:",omitempty"`
	Documents                      []*DocumentInput        `json:",omitempty"`
}

type BundleUpdateInput struct {
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *AuthInput
	OrdID                          *string
	ShortDescription               *string
	Links                          json.RawMessage
	Labels                         json.RawMessage
	CredentialExchangeStrategies   json.RawMessage
}

type BundlePage struct {
	Data       []*Bundle
	PageInfo   *pagination.Page
	TotalCount int
}

func (BundlePage) IsPageable() {}

func (i *BundleCreateInput) ToBundle(id, applicationID, tenantID string) *Bundle {
	if i == nil {
		return nil
	}

	return &Bundle{
		TenantID:                       tenantID,
		ApplicationID:                  applicationID,
		Name:                           i.Name,
		Description:                    i.Description,
		InstanceAuthRequestInputSchema: i.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            i.DefaultInstanceAuth.ToAuth(),
		OrdID:                          i.OrdID,
		ShortDescription:               i.ShortDescription,
		Links:                          i.Links,
		Labels:                         i.Labels,
		CredentialExchangeStrategies:   i.CredentialExchangeStrategies,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}
