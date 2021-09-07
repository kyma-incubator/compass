package tenant

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

type converter struct{}

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.BusinessTenantMapping) *tenant.Entity {
	if in == nil {
		return nil
	}
	return &tenant.Entity{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		Parent:         str.NewNullString(in.Parent),
		Type:           in.Type,
		ProviderName:   in.Provider,
		Status:         in.Status,
	}
}

// FromEntity missing godoc
func (c *converter) FromEntity(in *tenant.Entity) *model.BusinessTenantMapping {
	if in == nil {
		return nil
	}
	return &model.BusinessTenantMapping{
		ID:             in.ID,
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		Parent:         in.Parent.String,
		Type:           in.Type,
		Provider:       in.ProviderName,
		Status:         in.Status,
		Initialized:    in.Initialized,
	}
}

// ToGraphQL missing godoc
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

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant {
	tenants := make([]*graphql.Tenant, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		tenants = append(tenants, c.ToGraphQL(r))
	}

	return tenants
}
