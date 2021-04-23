package model

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
	Initialized    *bool // computed value
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
		Provider:       i.Provider,
		Status:         Active,
	}
}

func (i BusinessTenantMappingInput) WithExternalTenant(externalTenant string) BusinessTenantMappingInput {
	i.ExternalTenant = externalTenant
	return i
}
