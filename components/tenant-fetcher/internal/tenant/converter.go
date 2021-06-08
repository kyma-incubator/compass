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
		Name:           in.TenantId,
		ExternalTenant: in.TenantId,
		CustomerId:     in.CustomerId,
		Subdomain:      in.Subdomain,
		ProviderName:   in.TenantProvider,
		Status:         in.Status,
	}
}

func (c *converter) FromEntity(in *tenant.Entity) *model.TenantModel {
	if in == nil {
		return nil
	}
	return &model.TenantModel{
		ID:             in.ID,
		TenantId:       in.ExternalTenant,
		CustomerId:     in.CustomerId,
		Subdomain:      in.Subdomain,
		TenantProvider: in.ProviderName,
		Status:         in.Status,
	}
}
