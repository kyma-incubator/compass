package model

import (
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
	Tags                           *string
	LastUpdated                    time.Time
	Extensions                     *string
}

type BundleInput struct {
	ID                             string
	Title                          string
	ShortDescription               string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	Tags                           *string
	LastUpdated                    time.Time
	Extensions                     *string
	DefaultInstanceAuth            *AuthInput
	APIDefinitions                 []*APIDefinitionInput
	EventDefinitions               []*EventDefinitionInput
	Documents                      []*DocumentInput
}

func (bundle *Bundle) SetFromUpdateInput(update BundleInput) {
	bundle.Title = update.Title
	bundle.ShortDescription = update.ShortDescription
	bundle.Description = update.Description
	bundle.InstanceAuthRequestInputSchema = update.InstanceAuthRequestInputSchema
	bundle.DefaultInstanceAuth = update.DefaultInstanceAuth.ToAuth()
	bundle.Tags = update.Tags
	bundle.LastUpdated = update.LastUpdated
	bundle.Extensions = update.Extensions
}

type BundlePage struct {
	Data       []*Bundle
	PageInfo   *pagination.Page
	TotalCount int
}

func (BundlePage) IsPageable() {}

func (i *BundleInput) ToBundle(applicationID, tenantID string) *Bundle {
	if i == nil {
		return nil
	}

	return &Bundle{
		ID:                             i.ID,
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
