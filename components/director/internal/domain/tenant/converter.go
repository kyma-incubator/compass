package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
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
		Parents:        []string{},
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
		Parents:     in.Parents,
		Initialized: in.Initialized,
		Provider:    in.Provider,
	}
}

func (c *converter) ToGraphQLInput(in model.BusinessTenantMappingInput) graphql.BusinessTenantMappingInput {
	return graphql.BusinessTenantMappingInput{
		Name:           in.Name,
		ExternalTenant: in.ExternalTenant,
		Parents:        stringsToPointerStrings(in.Parents),
		Subdomain:      str.Ptr(in.Subdomain),
		Region:         str.Ptr(in.Region),
		Type:           in.Type,
		Provider:       in.Provider,
		LicenseType:    in.LicenseType,
		CustomerID:     in.CustomerID,
		CostObjectType: in.CostObjectType,
		CostObjectID:   in.CostObjectID,
	}
}

func (c *converter) MultipleInputFromGraphQL(ctx context.Context, in []*graphql.BusinessTenantMappingInput, retrieveTenantTypeFn func(ctx context.Context, t string) (string, error)) ([]model.BusinessTenantMappingInput, error) {
	externalTenantToType := make(map[string]string)
	for _, tnt := range in {
		externalTenantToType[tnt.ExternalTenant] = tnt.Type
	}

	res := make([]model.BusinessTenantMappingInput, 0, len(in))

	for _, tnt := range in {
		if tnt != nil {
			btm, err := c.InputFromGraphQL(ctx, *tnt, externalTenantToType, retrieveTenantTypeFn)
			if err != nil {
				return nil, err
			}
			res = append(res, btm)
		}
	}

	return res, nil
}

func (c *converter) InputFromGraphQL(ctx context.Context, tnt graphql.BusinessTenantMappingInput, externalTenantToType map[string]string, retrieveTenantTypeFn func(ctx context.Context, t string) (string, error)) (model.BusinessTenantMappingInput, error) {
	externalTenant := tnt.ExternalTenant
	trimmedParents := pointerStringsToStrings(tnt.Parents)

	switch tnt.Type {
	case tenant.TypeToStr(tenant.Customer):
		externalTenant = tenant.TrimCustomerIDLeadingZeros(tnt.ExternalTenant)
	case tenant.TypeToStr(tenant.Account):
		trimmedParents = make([]string, 0, len(tnt.Parents))
		for _, parent := range tnt.Parents {
			if parent != nil && str.PtrStrToStr(parent) != "" {
				parentType, err := c.getTenantType(ctx, externalTenantToType, *parent, retrieveTenantTypeFn)
				if err != nil {
					return model.BusinessTenantMappingInput{}, err
				}

				if parentType == tenant.TypeToStr(tenant.Customer) {
					trimmedParent := tenant.TrimCustomerIDLeadingZeros(*parent)
					trimmedParents = append(trimmedParents, trimmedParent)
				} else {
					trimmedParents = append(trimmedParents, *parent)
				}
			}
		}
	case tenant.TypeToStr(tenant.Organization):
		trimmedParents = make([]string, 0, len(tnt.Parents))
		for _, parent := range tnt.Parents {
			if parent != nil {
				trimmedParent := tenant.TrimCustomerIDLeadingZeros(*parent)
				trimmedParents = append(trimmedParents, trimmedParent)
			}
		}
	}

	customerID := tnt.CustomerID
	if tnt.CustomerID != nil {
		customerID = str.Ptr(tenant.TrimCustomerIDLeadingZeros(*customerID))
	}

	return model.BusinessTenantMappingInput{
		Name:           tnt.Name,
		ExternalTenant: externalTenant,
		Parents:        trimmedParents,
		Subdomain:      str.PtrStrToStr(tnt.Subdomain),
		Region:         str.PtrStrToStr(tnt.Region),
		Type:           tnt.Type,
		Provider:       tnt.Provider,
		LicenseType:    tnt.LicenseType,
		CustomerID:     customerID,
		CostObjectType: tnt.CostObjectType,
		CostObjectID:   tnt.CostObjectID,
	}, nil
}

func (c *converter) getTenantType(ctx context.Context, externalTenantToType map[string]string, externalTenant string, retrieveTenantTypeFn func(ctx context.Context, t string) (string, error)) (string, error) {
	tenantType, ok := externalTenantToType[externalTenant]
	if ok {
		return tenantType, nil
	}

	tenantType, err := retrieveTenantTypeFn(ctx, externalTenant)
	if err != nil {
		return "", errors.Wrapf(err, "while retrieving tenant type for tenant %q", externalTenant)
	}

	return tenantType, nil
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

func (c *converter) MultipleInputToGraphQLInput(in []model.BusinessTenantMappingInput) []graphql.BusinessTenantMappingInput {
	tenants := make([]graphql.BusinessTenantMappingInput, 0, len(in))
	for _, tnt := range in {
		tenants = append(tenants, c.ToGraphQLInput(tnt))
	}
	return tenants
}

// TenantAccessInputFromGraphQL converts the provided graphql.TenantAccessInput GraphQL-layer representation of a tenant accessto the service-layer representation model.TenantAccess.
func (c *converter) TenantAccessInputFromGraphQL(in graphql.TenantAccessInput) (*model.TenantAccess, error) {
	resourceType, err := fromTenantAccessObjectTypeToResourceType(in.ResourceType)
	if err != nil {
		return nil, err
	}

	return &model.TenantAccess{
		ExternalTenantID: in.TenantID,
		ResourceType:     resourceType,
		ResourceID:       in.ResourceID,
		Owner:            in.Owner,
	}, nil
}

// TenantAccessToGraphQL converts the provided model.TenantAccess service-layer representation of a tenant access to the GraphQL-layer representation graphql.TenantAccess.
func (c *converter) TenantAccessToGraphQL(in *model.TenantAccess) (*graphql.TenantAccess, error) {
	if in == nil {
		return nil, nil
	}

	resourceType, err := fromResourceTypeToTenantAccessObjectType(in.ResourceType)
	if err != nil {
		return nil, err
	}

	return &graphql.TenantAccess{
		TenantID:     in.ExternalTenantID,
		ResourceType: resourceType,
		ResourceID:   in.ResourceID,
		Owner:        in.Owner,
	}, nil
}

// TenantAccessToEntity converts the provided service-layer representation of a tenant access to the repository-layer one.
func (c *converter) TenantAccessToEntity(in *model.TenantAccess) *repo.TenantAccess {
	if in == nil {
		return nil
	}

	return &repo.TenantAccess{
		TenantID:   in.InternalTenantID,
		ResourceID: in.ResourceID,
		Owner:      in.Owner,
		Source:     in.Source,
	}
}

// TenantAccessFromEntity converts the provided repository-layer representation of a tenant access to the service-layer one.
func (c *converter) TenantAccessFromEntity(in *repo.TenantAccess) *model.TenantAccess {
	if in == nil {
		return nil
	}

	return &model.TenantAccess{
		InternalTenantID: in.TenantID,
		ResourceID:       in.ResourceID,
		Owner:            in.Owner,
		Source:           in.Source,
	}
}

func fromTenantAccessObjectTypeToResourceType(objectType graphql.TenantAccessObjectType) (resource.Type, error) {
	switch objectType {
	case graphql.TenantAccessObjectTypeApplication:
		return resource.Application, nil
	case graphql.TenantAccessObjectTypeRuntime:
		return resource.Runtime, nil
	case graphql.TenantAccessObjectTypeRuntimeContext:
		return resource.RuntimeContext, nil
	default:
		return "", errors.Errorf("Unknown tenant access resource type %q", objectType)
	}
}

func fromResourceTypeToTenantAccessObjectType(objectType resource.Type) (graphql.TenantAccessObjectType, error) {
	switch objectType {
	case resource.Application:
		return graphql.TenantAccessObjectTypeApplication, nil
	case resource.Runtime:
		return graphql.TenantAccessObjectTypeRuntime, nil
	case resource.RuntimeContext:
		return graphql.TenantAccessObjectTypeRuntimeContext, nil
	default:
		return "", errors.Errorf("Unknown tenant access resource type %q", objectType)
	}
}

func stringsToPointerStrings(input []string) []*string {
	result := make([]*string, len(input))
	for i := range input {
		result[i] = &input[i]
	}
	return result
}

func pointerStringsToStrings(input []*string) []string {
	result := make([]string, len(input))
	for i, s := range input {
		if s != nil {
			result[i] = *s
		}
	}
	return result
}
