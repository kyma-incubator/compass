package director

import (
	"context"
	"time"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlTools "github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
)

type TenantsResponse struct {
	Result []*schema.Tenant
}

func getToken(tenant string, scopes []string) (string, error) {
	token, err := jwtbuilder.Build(tenant, scopes, &jwtbuilder.Consumer{})
	if err != nil {
		return "", err
	}

	return token, nil
}

func GetTenants(directorURL string, externalTenantID string) ([]*schema.Tenant, error) {
	query := getTenantsQuery()

	req := gcli.NewRequest(query)

	token, err := getToken(externalTenantID, []string{"tenant:read"})
	if err != nil {
		return nil, err
	}

	client := gqlTools.NewAuthorizedGraphQLClientWithCustomURL(token, directorURL)

	var response TenantsResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Run(ctx, req, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func getTenantsQuery() string {
	return `query {
		result: tenants {
			id
			name
			internalID
		}
	}`
}
