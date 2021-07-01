package tenant

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in *model.BusinessTenantMapping) *tenant.Entity {
	if in == nil {
		return nil
	}
	return &tenant.Entity{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		Parent:         NewNullString(in.Parent),
		Type:           tenant.Type(in.Type),
		ProviderName:   in.Provider,
		Status:         tenant.Status(in.Status),
	}
}

func (c *converter) FromEntity(in *tenant.Entity) *model.BusinessTenantMapping {
	if in == nil {
		return nil
	}
	return &model.BusinessTenantMapping{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		Parent:         in.Parent.String,
		Type:           string(in.Type),
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

