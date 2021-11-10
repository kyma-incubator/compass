package graphqlclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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
	return modifyTenants(ctx, tenants, "writeTenants", d.client)
}

func (d *Director) DeleteTenants(ctx context.Context, tenants []model.BusinessTenantMappingInput) error {
	return modifyTenants(ctx, tenants, "deleteTenants", d.client)
}

func (d *Director) CreateLabelDefinition(ctx context.Context, lblDef graphql.LabelDefinitionInput) error {
	res := struct {
		LabelDefinition graphql.LabelDefinition `json:"createLabelDefinition"`
	}{}

	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.LabelDefinitionInputToGQL(lblDef)
	gqlFieldProvider := graphqlizer.GqlFieldsProvider{}
	if err != nil {
		return errors.Wrap(err, "while creating label definition input")
	}
	lblDefQuery := fmt.Sprintf(`mutation { createLabelDefinition(in: %s ) {"%s"}}`, in, gqlFieldProvider.ForLabelDefinition())
	gRequest := gcli.NewRequest(lblDefQuery)

	if err = d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

func (d *Director) UpdateLabelDefinition(ctx context.Context, lblDef graphql.LabelDefinitionInput) error {
	res := struct {
		LabelDefinition graphql.LabelDefinition `json:"updateLabelDefinition"`
	}{}

	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.LabelDefinitionInputToGQL(lblDef)
	gqlFieldProvider := graphqlizer.GqlFieldsProvider{}
	if err != nil {
		return errors.Wrap(err, "while creating label definition input")
	}
	lblDefQuery := fmt.Sprintf(`mutation { updateLabelDefinition(in: %s ) {"%s"}}`, in, gqlFieldProvider.ForLabelDefinition())
	gRequest := gcli.NewRequest(lblDefQuery)

	if err = d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}

func modifyTenants(ctx context.Context, tenants []model.BusinessTenantMappingInput, mutation string, client GraphQLClient) error {
	var res map[string]interface{}

	mutationBuilder := strings.Builder{}
	for _, tnt := range tenants {
		mutationBuilder.WriteString(fmt.Sprintf(tenantQueryPattern, tnt.Name, tnt.ExternalTenant, tnt.Parent, tnt.Subdomain, tnt.Region, tnt.Type, tnt.Provider))
	}
	tenantsQuery := fmt.Sprintf("mutation { %s(in:[%s])}", mutation, mutationBuilder.String()[:mutationBuilder.Len()-1])
	gRequest := gcli.NewRequest(tenantsQuery)
	if err := client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql query")
	}
	return nil
}
