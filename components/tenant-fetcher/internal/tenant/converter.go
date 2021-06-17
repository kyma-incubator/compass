package tenant

import (
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in model.TenantModel) tenant.Entity {
	return tenant.Entity{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.TenantId,
		Parent:         in.ParentInternal,
		Type:           in.Type,
		ProviderName:   in.Provider,
		Status:         in.Status,
	}
}

func (c *converter) FromEntity(in tenant.Entity) model.TenantModel {
	return model.TenantModel{
		ID:             in.ID,
		Name:           in.Name,
		TenantId:       in.ExternalTenant,
		ParentInternal: in.Parent,
		Type:           in.Type,
		Provider:       in.ProviderName,
		Status:         in.Status,
	}
}
