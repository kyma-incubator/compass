package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Bundle struct {
	ID                             string
	TenantID                       string
	ApplicationID                  string
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *Auth
}

func (bndl *Bundle) SetFromUpdateInput(update BundleUpdateInput) {
	bndl.Name = update.Name
	bndl.Description = update.Description
	bndl.InstanceAuthRequestInputSchema = update.InstanceAuthRequestInputSchema
	bndl.DefaultInstanceAuth = update.DefaultInstanceAuth.ToAuth()
}

type BundleCreateInput struct {
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *AuthInput
	APIDefinitions                 []*APIDefinitionInput
	EventDefinitions               []*EventDefinitionInput
	Documents                      []*DocumentInput
}

type BundleUpdateInput struct {
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *string
	DefaultInstanceAuth            *AuthInput
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
		Name:                           i.Name,
		Description:                    i.Description,
		InstanceAuthRequestInputSchema: i.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            i.DefaultInstanceAuth.ToAuth(),
	}
}
