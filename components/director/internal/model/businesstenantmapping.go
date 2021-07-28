package model

import "github.com/kyma-incubator/compass/components/director/pkg/tenant"

type BusinessTenantMapping struct {
	ID             string
	Name           string
	ExternalTenant string
	Parent         string
	Subdomain      string
	Type           tenant.Type
	Provider       string
	Status         tenant.Status
	Initialized    *bool // computed value
}

func (t BusinessTenantMapping) WithExternalTenant(externalTenant string) BusinessTenantMapping {
	t.ExternalTenant = externalTenant
	return t
}

func (t BusinessTenantMapping) WithStatus(status tenant.Status) BusinessTenantMapping {
	t.Status = status
	return t
}

type BusinessTenantMappingInput struct {
	Name           string `json:"name"`
	ExternalTenant string `json:"id"`
	Parent         string `json:"parent"`
	Subdomain      string `json:"subdomain"`
	Type           string `json:"type"`
	Provider       string
}

type MovedRuntimeByLabelMappingInput struct {
	LabelValue   string
	SourceTenant string
	TargetTenant string
}

func (i *BusinessTenantMappingInput) ToBusinessTenantMapping(id string) *BusinessTenantMapping {
	return &BusinessTenantMapping{
		ID:             id,
		Name:           i.Name,
		ExternalTenant: i.ExternalTenant,
		Parent:         i.Parent,
		Subdomain:      i.Subdomain,
		Type:           tenant.StrToType(i.Type),
		Provider:       i.Provider,
		Status:         tenant.Active,
	}
}

func (i BusinessTenantMappingInput) WithExternalTenant(externalTenant string) BusinessTenantMappingInput {
	i.ExternalTenant = externalTenant
	return i
}
