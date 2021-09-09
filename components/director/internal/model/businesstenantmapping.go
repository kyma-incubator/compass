package model

import "github.com/kyma-incubator/compass/components/director/pkg/tenant"

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

// BusinessTenantMappingInput missing godoc
type BusinessTenantMappingInput struct {
	Name           string `json:"name"`
	ExternalTenant string `json:"id"`
	Parent         string `json:"parent"`
	Subdomain      string `json:"subdomain"`
	Region         string `json:"region"`
	Type           string `json:"type"`
	Provider       string
}

// MovedRuntimeByLabelMappingInput missing godoc
type MovedRuntimeByLabelMappingInput struct {
	LabelValue   string
	SourceTenant string
	TargetTenant string
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
	}
}

// WithExternalTenant missing godoc
func (i BusinessTenantMappingInput) WithExternalTenant(externalTenant string) BusinessTenantMappingInput {
	i.ExternalTenant = externalTenant
	return i
}
