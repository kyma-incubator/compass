package tenant

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

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
		Initialized:    in.Initialized,
	}
}

func (c *converter) ToGraphQL(in *model.BusinessTenantMapping) *graphql.Tenant {
	if in == nil {
		return nil
	}

	return &graphql.Tenant{
		ID:          in.ExternalTenant,
		InternalID:  in.ID,
		Name:        str.Ptr(in.Name),
		Initialized: in.Initialized,
	}

}

func (c *converter) MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant {
	var tenants []*graphql.Tenant
	for _, r := range in {
		if r == nil {
			continue
		}

		tenants = append(tenants, c.ToGraphQL(r))
	}

	return tenants
}
