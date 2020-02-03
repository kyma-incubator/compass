package tenant

import "github.com/kyma-incubator/compass/components/director/internal/model"

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in *model.BusinessTenantMapping) *Entity {
	if in == nil {
		return nil
	}
	return &Entity{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		ProviderName:   in.Provider,
		Status:         TenantStatus(in.Status),
	}
}

func (c *converter) FromEntity(in *Entity) *model.BusinessTenantMapping {
	if in == nil {
		return nil
	}
	return &model.BusinessTenantMapping{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		Provider:       in.ProviderName,
		Status:         model.TenantStatus(in.Status),
	}
}
