package tenant

import "github.com/kyma-incubator/compass/components/director/internal/model"

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in *model.TenantMapping) *Entity {
	if in == nil {
		return nil
	}
	return &Entity{
		ID:             in.ID,
		Name:           in.Name,
		InternalTenant: in.InternalTenant,
		ExternalTenant: in.ExternalTenant,
		ProviderName:   in.Provider,
	}
}

func (c *converter) FromEntity(in *Entity) *model.TenantMapping {
	if in == nil {
		return nil
	}
	return &model.TenantMapping{
		ID:             in.ID,
		Name:           in.Name,
		InternalTenant: in.InternalTenant,
		ExternalTenant: in.ExternalTenant,
		Provider:       in.ProviderName,
	}
}
