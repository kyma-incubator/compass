package model

// TenantBusinessType represent structure for TenantBusinessType
type TenantBusinessType struct {
	ID   string
	Code string
	Name string
}

// TenantBusinessTypeInput is an input for creating a new TenantBusinessType
type TenantBusinessTypeInput struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ToModel converts TenantBusinessTypeInput to TenantBusinessType
func (i *TenantBusinessTypeInput) ToModel(id string) *TenantBusinessType {
	if i == nil {
		return nil
	}

	return &TenantBusinessType{
		ID:   id,
		Code: i.Code,
		Name: i.Name,
	}
}
