package graphqlclient

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// GraphQLClient expects graphql implementation
//
//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type GraphQLClient interface {
	Run(context.Context, *gcli.Request, interface{}) error
}

// Director is an GraphQLClient implementation
type Director struct {
	client GraphQLClient
}

// TenantResponse represents Tenant object from GraphQL response
type TenantResponse struct {
	Result *graphql.Tenant `json:"result"`
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

// SubscribeTenant makes graphql query tenant subscription
func (d *Director) SubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionProviderAppName, subscriptionPayload string) error {
	var res map[string]interface{}
	subscriptionMutation := fmt.Sprintf(`mutation { subscribeTenant(providerID: "%s", subaccountID: "%s", providerSubaccountID: "%s", consumerTenantID: "%s", region: "%s", subscriptionAppName: "%s", subscriptionPayload: %q)}`, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionProviderAppName, subscriptionPayload)

	log.C(ctx).Infof("ALEX SubscribeTenant %s", subscriptionMutation)

	gRequest := gcli.NewRequest(subscriptionMutation)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql mutation")
	}
	return nil
}

// UnsubscribeTenant makes graphql query tenant unsubscription
func (d *Director) UnsubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionPayload string) error {
	var res map[string]interface{}

	unsubscriptionMutation := fmt.Sprintf(`mutation { unsubscribeTenant(providerID: "%s", subaccountID: "%s", providerSubaccountID: "%s", consumerTenantID: "%s", region: "%s", subscriptionPayload: %q)}`, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionPayload)
	gRequest := gcli.NewRequest(unsubscriptionMutation)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql mutation")
	}
	return nil
}

// ExistsTenantByExternalID makes graphql query to check if tenant exists
func (d *Director) ExistsTenantByExternalID(ctx context.Context, tenantID string) (bool, error) {
	query := fmt.Sprintf(`query { result: tenantByExternalID(id: "%s") { id internalID name type parents labels }}`, tenantID)
	var res TenantResponse
	gRequest := gcli.NewRequest(query)

	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		if d.isGQLNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "while executing gql query")
	}

	return true, nil
}

func (d *Director) isGQLNotFoundError(err error) bool {
	err = errors.Cause(err)

	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Object not found")
}
