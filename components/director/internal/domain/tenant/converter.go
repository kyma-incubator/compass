package tenant

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

type converter struct{}

// NewConverter returns a new Converter that can later be used to make the conversions between the GraphQL, service, and repository layer representations of a Compass tenant.
func NewConverter() *converter {
	return &converter{}
}

// ToEntity converts the provided service-layer representation of a tenant to the repository-layer one tenant.Entity.
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

// FromEntity converts the provided tenant.Entity repo-layer representation of a tenant to the service-layer representation model.BusinessTenantMapping.
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

// ToGraphQL converts the provided model.BusinessTenantMapping service-layer representation of a tenant to the GraphQL-layer representation graphql.Tenant.
func (c *converter) ToGraphQL(in *model.BusinessTenantMapping) *graphql.Tenant {
	if in == nil {
		return nil
	}

	return &graphql.Tenant{
		ID:          in.ExternalTenant,
		InternalID:  in.ID,
		Name:        str.Ptr(in.Name),
		Type:        tenant.TypeToStr(in.Type),
		ParentID:    in.Parent,
		Initialized: in.Initialized,
	}
}

// MultipleToGraphQL converts all the provided model.BusinessTenantMapping service-layer representations of a tenant to the GraphQL-layer representations graphql.Tenant.
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
