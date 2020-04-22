package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type TenantStatus string

const (
	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"
)

type BusinessTenantMapping struct {
	ID             string
	Name           string
	ExternalTenant string
	Provider       string
	Status         TenantStatus
}

type BusinessTenantMappingPage struct {
	Data       []*BusinessTenantMapping
	PageInfo   *pagination.Page
	TotalCount int
}

func (t BusinessTenantMapping) WithExternalTenant(externalTenant string) BusinessTenantMapping {
	t.ExternalTenant = externalTenant
	return t
}

func (t BusinessTenantMapping) WithStatus(status TenantStatus) BusinessTenantMapping {
	t.Status = status
	return t
}

type BusinessTenantMappingInput struct {
	Name           string `json:"name"`
	ExternalTenant string `json:"id"`
	Provider       string
}

func (i *BusinessTenantMappingInput) ToBusinessTenantMapping(id string) *BusinessTenantMapping {
	return &BusinessTenantMapping{
		ID:             id,
		Name:           i.Name,
		ExternalTenant: i.ExternalTenant,
		Provider:       i.Provider,
		Status:         Active,
	}
}

func (i BusinessTenantMappingInput) WithExternalTenant(externalTenant string) BusinessTenantMappingInput {
	i.ExternalTenant = externalTenant
	return i
}
