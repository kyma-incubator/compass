package tenant

import (
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
		Subdomain:      in.Subdomain,
		Parent:         str.NewNullString(in.ParentInternalId),
		Type:           in.Type,
		ProviderName:   in.Provider,
		Status:         in.Status,
	}
}

func (c *converter) FromEntity(in tenant.Entity) model.TenantModel {
	return model.TenantModel{
		ID:               in.ID,
		Name:             in.Name,
		TenantId:         in.ExternalTenant,
		Subdomain:        in.Subdomain,
		ParentInternalId: in.Parent.String,
		Type:             in.Type,
		Provider:         in.ProviderName,
		Status:           in.Status,
	}
}
