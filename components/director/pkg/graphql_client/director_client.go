package graphqlclient

import (
	"context"
	"fmt"
	"strings"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tenantQueryPattern = `{name: "%s",externalTenant: "%s",parent: "%s",subdomain: "%s",region:"%s",type:"%s",provider: "%s"},`
)

// GraphQLClient expects graphql implementation
//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore
type GraphQLClient interface {
	Run(context.Context, *gcli.Request, interface{}) error
}

type Director struct {
	client GraphQLClient
}

func NewDirector(client GraphQLClient) *Director {
	return &Director{
		client: client,
	}
}

func (d *Director) WriteTenants(ctx context.Context, tenants []model.BusinessTenantMappingInput) error {
	var res map[string]interface{}

	mutationBuilder := strings.Builder{}
	for _, tnt := range tenants {
		mutationBuilder.WriteString(fmt.Sprintf(tenantQueryPattern, tnt.Name, tnt.ExternalTenant, tnt.Parent, tnt.Subdomain, tnt.Region, tnt.Type, tnt.Provider))
	}
	tenantsQuery := fmt.Sprintf("mutation { writeTenants(in:[%s])}", mutationBuilder.String()[:mutationBuilder.Len()-1])
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

func (d *Director) DeleteTenants(ctx context.Context, tenants []model.BusinessTenantMappingInput) error {
	var res map[string]interface{}

	mutationBuilder := strings.Builder{}
	for _, tnt := range tenants {
		mutationBuilder.WriteString(fmt.Sprintf(tenantQueryPattern, tnt.Name, tnt.ExternalTenant, tnt.Parent, tnt.Subdomain, tnt.Region, tnt.Type, tnt.Provider))
	}
	tenantsQuery := fmt.Sprintf("mutation { deleteTenants(in:[%s])}", mutationBuilder.String()[:mutationBuilder.Len()-1])
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}
