package graphqlclient

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// GraphQLClient expects graphql implementation
//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore
type GraphQLClient interface {
	Run(context.Context, *gcli.Request, interface{}) error
}

// Director is an GraphQLClient implementation
type Director struct {
	client GraphQLClient
}

// NewDirector creates new director with given client
func NewDirector(client GraphQLClient) *Director {
	return &Director{
		client: client,
	}
}

// WriteTenants makes graphql query for tenant creation
func (d *Director) WriteTenants(ctx context.Context, tenants []graphql.BusinessTenantMappingInput) error {
	var res map[string]interface{}

	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.WriteTenantsInputToGQL(tenants)
	if err != nil {
		return errors.Wrap(err, "while creating tenants input")
	}
	tenantsQuery := fmt.Sprintf("mutation { writeTenants(in:[%s])}", in)
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

// DeleteTenants makes graphql query for tenant deletion
func (d *Director) DeleteTenants(ctx context.Context, tenants []graphql.BusinessTenantMappingInput) error {
	var res map[string]interface{}

	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.DeleteTenantsInputToGQL(tenants)
	if err != nil {
		return errors.Wrap(err, "while creating tenants input")
	}
	tenantsQuery := fmt.Sprintf("mutation { deleteTenants(in:[%s])}", in)
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

// UpdateTenant makes graphql query for tenant update
func (d *Director) UpdateTenant(ctx context.Context, id string, tenant graphql.BusinessTenantMappingInput) error {
	var res map[string]interface{}

	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.UpdateTenantsInputToGQL(tenant)
	if err != nil {
		return errors.Wrap(err, "while creating tenants input")
	}
	tenantsQuery := fmt.Sprintf(`mutation { updateTenants(id: "%s", in:%s)}`, id, in)
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

// CreateLabelDefinition makes graphql query for labeldef creation
func (d *Director) CreateLabelDefinition(ctx context.Context, lblDef graphql.LabelDefinitionInput, tenant string) error {
	return modifyLabelDefs(ctx, lblDef, tenant, "createLabelDefinition", d.client)
}

// UpdateLabelDefinition makes graphql query for updating labeldef
func (d *Director) UpdateLabelDefinition(ctx context.Context, lblDef graphql.LabelDefinitionInput, tenant string) error {
	return modifyLabelDefs(ctx, lblDef, tenant, "updateLabelDefinition", d.client)
}

// SetRuntimeTenant makes graphql query for setting runtime's tenant
func (d *Director) SetRuntimeTenant(ctx context.Context, runtimeID, tenantID string) error {
	res := struct {
		Runtime graphql.Runtime `json:"result"`
	}{}

	gqlFieldProvider := graphqlizer.GqlFieldsProvider{}
	lblDefQuery := fmt.Sprintf(`mutation { result: setRuntimeTenant(runtimeID:"%s", tenantID:"%s") {%s}}`, runtimeID, tenantID, gqlFieldProvider.ForRuntime())
	gRequest := gcli.NewRequest(lblDefQuery)

	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

func modifyLabelDefs(ctx context.Context, lblDef graphql.LabelDefinitionInput, tenant string, mutation string, client GraphQLClient) error {
	res := struct {
		LabelDefinition graphql.LabelDefinition `json:"result"`
	}{}

	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.LabelDefinitionInputToGQL(lblDef)
	gqlFieldProvider := graphqlizer.GqlFieldsProvider{}
	if err != nil {
		return errors.Wrap(err, "while creating label definition input")
	}
	lblDefQuery := fmt.Sprintf(`mutation { result: %s(in: %s ) {%s}}`, mutation, in, gqlFieldProvider.ForLabelDefinition())
	gRequest := gcli.NewRequest(lblDefQuery)
	gRequest.Header.Set("Tenant", tenant)
	if err = client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}
