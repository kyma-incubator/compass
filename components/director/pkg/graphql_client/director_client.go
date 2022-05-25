package graphqlclient

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"reflect"
	"unsafe"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// GraphQLClient expects graphql implementation
//go:generate mockery --name=GraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type GraphQLClient interface {
	Run(context.Context, *gcli.Request, interface{}) error
}

// Director is an GraphQLClient implementation
type Director struct {
	client GraphQLClient
}

type RuntimeResponse struct {
	Result graphql.RuntimePage `json:"result"`
}

type AppTemplateResponse struct {
	Result graphql.ApplicationTemplatePage `json:"result"`
}

type ApplicationResponse struct {
	Result graphql.Application `json:"result"`
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

// SubscribeTenantToRuntime makes graphql query tenant-runtime subscription
func (d *Director) SubscribeTenantToRuntime(ctx context.Context, providerID string, subaccountID string, providerSubaccountID string, region string) error {
	var res map[string]interface{}

	subscriptionMutation := fmt.Sprintf(`mutation { subscribeTenantToRuntime(providerID: "%s", subaccountID: "%s", providerSubaccountID: "%s", region: "%s")}`, providerID, subaccountID, providerSubaccountID, region)
	gRequest := gcli.NewRequest(subscriptionMutation)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql mutation")
	}
	return nil
}

// UnsubscribeTenantFromRuntime makes graphql query tenant-runtime unsubscription
func (d *Director) UnsubscribeTenantFromRuntime(ctx context.Context, providerID string, subaccountID string, providerSubaccountID string, region string) error {
	var res map[string]interface{}

	unsubscriptionMutation := fmt.Sprintf(`mutation { unsubscribeTenantFromRuntime(providerID: "%s", subaccountID: "%s", providerSubaccountID: "%s", region: "%s")}`, providerID, subaccountID, providerSubaccountID, region)
	gRequest := gcli.NewRequest(unsubscriptionMutation)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql mutation")
	}
	return nil
}

// RegisterApplicationFromTemplate makes graphql mutation for registering application from a template
func (d *Director) RegisterApplicationFromTemplate(ctx context.Context, appTemplateName, subaccountTenantID, subscriptionAppName string) error {
	var res map[string]interface{}

	applicationFromTemplateMutation := fmt.Sprintf(`mutation { registerApplicationFromTemplate( in: { templateName: "%s" values: [{ placeholder:"name", value:"%s" }, { placeholder:"display-name", value:"%s" }] } ) { id name labels } }`, appTemplateName, subscriptionAppName, subscriptionAppName)

	gRequest := gcli.NewRequest(applicationFromTemplateMutation)
	gRequest.Header.Set("tenant", subaccountTenantID)
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return errors.Wrap(err, "while executing gql mutation")
	}
	return nil
}

// GetRuntimes makes a graphql query for fetching runtimes by labels
func (d *Director) GetRuntimes(ctx context.Context, region, selfRegisterDistinguishLabelKey, selfRegisterDistinguishLabelValue string) (graphql.RuntimePage, error) {
	var res RuntimeResponse

	printContextInternals(ctx, false)

	applicationFromTemplateMutation := fmt.Sprintf(`query { result: runtimes(filter:[{key:"%s", query:"\"%s\""}, { key: "%s", query: "\"%s\""}]) { data { id name labels } totalCount } }`, tenant.RegionLabelKey, region, selfRegisterDistinguishLabelKey, selfRegisterDistinguishLabelValue)
	gRequest := gcli.NewRequest(applicationFromTemplateMutation)
	gRequest.Header.Set("tenant", "c7ebc4a9-01a6-4d77-82db-de0bae7a645a")
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return graphql.RuntimePage{}, errors.Wrap(err, "while executing gql query")
	}
	return res.Result, nil
}

// GetApplicationTemplates makes a graphql query for fetching application templates by labels
func (d *Director) GetApplicationTemplates(ctx context.Context, region, selfRegisterDistinguishLabelKey, selfRegisterDistinguishLabelValue string) (graphql.ApplicationTemplatePage, error) {
	var res AppTemplateResponse

	applicationFromTemplateMutation := fmt.Sprintf(`query { result: applicationTemplates(filter:[{key: "%s", query: "\"%s\""}, { key: "%s", query: "\"%s\""}]) { data { id name labels } totalCount } }`, tenant.RegionLabelKey, region, selfRegisterDistinguishLabelKey, selfRegisterDistinguishLabelValue)

	gRequest := gcli.NewRequest(applicationFromTemplateMutation)
	gRequest.Header.Set("tenant", "c7ebc4a9-01a6-4d77-82db-de0bae7a645a")
	if err := d.client.Run(ctx, gRequest, &res); err != nil {
		return graphql.ApplicationTemplatePage{}, errors.Wrap(err, "while executing gql query")
	}
	return res.Result, nil
}

func printContextInternals(ctx interface{}, inner bool) {
	contextValues := reflect.ValueOf(ctx).Elem()
	contextKeys := reflect.TypeOf(ctx).Elem()

	if !inner {
		fmt.Printf("\nFields for %s.%s\n", contextKeys.PkgPath(), contextKeys.Name())
	}

	if contextKeys.Kind() == reflect.Struct {
		for i := 0; i < contextValues.NumField(); i++ {
			reflectValue := contextValues.Field(i)
			reflectValue = reflect.NewAt(reflectValue.Type(), unsafe.Pointer(reflectValue.UnsafeAddr())).Elem()

			reflectField := contextKeys.Field(i)

			if reflectField.Name == "Context" {
				printContextInternals(reflectValue.Interface(), true)
			} else {
				fmt.Printf("field name: %+v\n", reflectField.Name)
				fmt.Printf("value: %+v\n", reflectValue.Interface())
			}
		}
	} else {
		fmt.Printf("context is empty (int)\n")
	}
}
