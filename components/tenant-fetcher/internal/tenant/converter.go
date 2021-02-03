package tenant

import "github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in model.TenantModel) Entity {
	return Entity{
		ID:             "",
		Name:           in.GlobalAccountGUID,
		ExternalTenant: in.GlobalAccountGUID,
		ProviderName:   "",
		Status:         "",
	}
}

func (c *converter) FromEntity(in *Entity) *model.TenantModel {
	if in == nil {
		return nil
	}
	return &model.TenantModel{
		UserId:            in.ID,
		GlobalAccountGUID: in.ExternalTenant,
	}
}
