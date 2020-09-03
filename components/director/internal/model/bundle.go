package model

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"time"
)

type Bundle struct {
	ID                             string
	TenantID                       string
	ApplicationID                  string
	Title                          string
	ShortDescription               string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *Auth
	Tags                           json.RawMessage
	LastUpdated                    time.Time
	Extensions                     json.RawMessage
}

func (bundle *Bundle) SetFromUpdateInput(update BundleUpdateInput) {
	bundle.Title = update.Title
	bundle.ShortDescription = update.ShortDescription
	bundle.Description = update.Description
	bundle.InstanceAuthRequestInputSchema = update.InstanceAuthRequestInputSchema
	bundle.DefaultInstanceAuth = update.DefaultInstanceAuth.ToAuth()
	bundle.Tags = update.Tags
	bundle.LastUpdated = update.LastUpdated
	bundle.Extensions = update.Extensions
}

type BundleCreateInput struct {
	Title                          string
	ShortDescription               string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	Tags                           json.RawMessage
	LastUpdated                    time.Time
	Extensions                     json.RawMessage
	DefaultInstanceAuth            *AuthInput
	APIDefinitions                 []*APIDefinitionInput
	EventDefinitions               []*EventDefinitionInput
	Documents                      []*DocumentInput
}

type BundleUpdateInput struct {
	Title                          string
	ShortDescription               string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *AuthInput
	Tags                           json.RawMessage
	LastUpdated                    time.Time
	Extensions                     json.RawMessage
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
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  applicationID,
		Title:                          i.Title,
		ShortDescription:               i.ShortDescription,
		Description:                    i.Description,
		InstanceAuthRequestInputSchema: i.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            i.DefaultInstanceAuth.ToAuth(),
		Tags:                           i.Tags,
		LastUpdated:                    i.LastUpdated,
		Extensions:                     i.Extensions,
	}
}
