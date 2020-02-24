package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Package struct {
	ID                             string
	TenantID                       string
	ApplicationID                  string
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *interface{}
	DefaultInstanceAuth            *Auth
}

func (pkg *Package) SetFromUpdateInput(update PackageUpdateInput) {
	pkg.Name = update.Name
	pkg.Description = update.Description
	pkg.InstanceAuthRequestInputSchema = update.InstanceAuthRequestInputSchema
	pkg.DefaultInstanceAuth = update.DefaultInstanceAuth.ToAuth()
}

type PackageCreateInput struct {
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *interface{}
	DefaultInstanceAuth            *AuthInput
	APIDefinitions                 []*APIDefinitionInput
	EventDefinitions               []*EventDefinitionInput
	Documents                      []*DocumentInput
}

type PackageUpdateInput struct {
	Name                           string
	Description                    *string
	InstanceAuthRequestInputSchema *interface{}
	DefaultInstanceAuth            *AuthInput
}

type PackagePage struct {
	Data       []*Package
	PageInfo   *pagination.Page
	TotalCount int
}

func (PackagePage) IsPageable() {}

func (i *PackageCreateInput) ToPackage(id, applicationID, tenantID string) *Package {
	if i == nil {
		return nil
	}

	return &Package{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  applicationID,
		Name:                           i.Name,
		Description:                    i.Description,
		InstanceAuthRequestInputSchema: i.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            i.DefaultInstanceAuth.ToAuth(),
	}
}
