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

	fieldProvider := graphqlizer.GqlFieldsProvider{}
	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.UpdateTenantsInputToGQL(tenant)
	if err != nil {
		return errors.Wrap(err, "while creating tenants input")
	}
	tenantsQuery := fmt.Sprintf(`mutation { updateTenant(id: "%s", in:%s) { %s }}`, id, in, fieldProvider.ForTenant())
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}
