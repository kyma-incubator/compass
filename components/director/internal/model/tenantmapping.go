package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type TenantStatus string

const (
	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"
)

type TenantMapping struct {
	ID             string
	Name           string
	ExternalTenant string
	InternalTenant string
	Provider       string
	Status         TenantStatus
}

type TenantMappingPage struct {
	Data       []*TenantMapping
	PageInfo   *pagination.Page
	TotalCount int
}

func (t TenantMapping) IsIn(tenantsPage TenantMappingPage) bool {
	for _, tenant := range tenantsPage.Data {
		if tenant.ExternalTenant == t.ExternalTenant {
			return true
		}
	}
	return false
}

type TenantMappingInput struct {
	Name           string
	ExternalTenant string
	Provider       string
}

func (i *TenantMappingInput) ToTenantMapping(id, internalTenant string) *TenantMapping {
	return &TenantMapping{
		ID:             id,
		Name:           i.Name,
		ExternalTenant: i.ExternalTenant,
		InternalTenant: internalTenant,
		Provider:       i.Provider,
		Status:         Active,
	}
}
