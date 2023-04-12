package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

// BusinessTenantMapping missing godoc
type BusinessTenantMapping struct {
	ID             string
	Name           string
	ExternalTenant string
	Parent         string
	Type           tenant.Type
	Provider       string
	Status         tenant.Status
	Initialized    *bool // computed value
	LicenseType    *string
}

// WithExternalTenant missing godoc
func (t BusinessTenantMapping) WithExternalTenant(externalTenant string) BusinessTenantMapping {
	t.ExternalTenant = externalTenant
	return t
}

// WithStatus missing godoc
func (t BusinessTenantMapping) WithStatus(status tenant.Status) BusinessTenantMapping {
	t.Status = status
	return t
}

// ToInput converts BusinessTenantMapping to BusinessTenantMappingInput
func (t BusinessTenantMapping) ToInput() BusinessTenantMappingInput {
	return BusinessTenantMappingInput{
		Name:           t.Name,
		ExternalTenant: t.ExternalTenant,
		Parent:         t.Parent,
		Subdomain:      "",
		Region:         "",
		Type:           tenant.TypeToStr(t.Type),
		Provider:       t.Provider,
		LicenseType:    t.LicenseType,
	}
}

// BusinessTenantMappingInput missing godoc
type BusinessTenantMappingInput struct {
	Name           string `json:"name"`
	ExternalTenant string `json:"id"`
	Parent         string `json:"parent"`
	Subdomain      string `json:"subdomain"`
	Region         string `json:"region"`
	Type           string `json:"type"`
	Provider       string
	LicenseType    *string `json:"licenseType"`
}

// MovedSubaccountMappingInput missing godoc
type MovedSubaccountMappingInput struct {
	TenantMappingInput BusinessTenantMappingInput
	SubaccountID       string
	SourceTenant       string
	TargetTenant       string
}

// ToBusinessTenantMapping missing godoc
func (i *BusinessTenantMappingInput) ToBusinessTenantMapping(id string) *BusinessTenantMapping {
	return &BusinessTenantMapping{
		ID:             id,
		Name:           i.Name,
		ExternalTenant: i.ExternalTenant,
		Parent:         i.Parent,
		Type:           tenant.StrToType(i.Type),
		Provider:       i.Provider,
		Status:         tenant.Active,
		LicenseType:    i.LicenseType,
	}
}

// WithExternalTenant missing godoc
func (i BusinessTenantMappingInput) WithExternalTenant(externalTenant string) BusinessTenantMappingInput {
	i.ExternalTenant = externalTenant
	return i
}

// BusinessTenantMappingPage Struct for BusinessTenantMapping data with page info
type BusinessTenantMappingPage struct {
	Data       []*BusinessTenantMapping
	PageInfo   *pagination.Page
	TotalCount int
}
