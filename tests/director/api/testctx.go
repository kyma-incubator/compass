package api

import (
	"context"
	"os"
	"reflect"
	"strings"

	gqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
	"github.com/kyma-incubator/compass/tests/director/pkg/retrier"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var tc *testContext

// testContext contains dependencies that help executing tests
type testContext struct {
	graphqlizer       gqlizer.Graphqlizer
	gqlFieldsProvider gqlizer.GqlFieldsProvider
	currentScopes     []string
	cli               *gcli.Client
}

const defaultScopes = "runtime:write application:write tenant:read label_definition:write integration_system:write application:read runtime:read label_definition:read integration_system:read health_checks:read application_template:read application_template:write eventing:manage automatic_scenario_assignment:read automatic_scenario_assignment:write"

func newTestContext() (*testContext, error) {

	scopesStr := os.Getenv("ALL_SCOPES")
	if scopesStr == "" {
		scopesStr = defaultScopes
	}

	currentScopes := strings.Split(scopesStr, " ")

	bearerToken, err := jwtbuilder.Do(testTenants.GetDefaultTenantID(), currentScopes)
	if err != nil {
		return nil, errors.Wrap(err, "while building JWT token")
	}

	return &testContext{
		graphqlizer:       gqlizer.Graphqlizer{},
		gqlFieldsProvider: gqlizer.GqlFieldsProvider{},
		currentScopes:     currentScopes,
		cli:               gql.NewAuthorizedGraphQLClient(bearerToken),
	}, nil
}

func (tc *testContext) RunOperation(ctx context.Context, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	return retrier.DoOnTemporaryConnectionProblems("DirectorAPITestContext", func() error { return tc.cli.Run(ctx, req, &m) })
}

func (tc *testContext) RunOperationWithCustomTenant(ctx context.Context, tenant string, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, tenant, tc.currentScopes, req, resp)
}

func (tc *testContext) RunOperationWithCustomScopes(ctx context.Context, scopes []string, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, testTenants.GetDefaultTenantID(), scopes, req, resp)
}

func (tc *testContext) RunOperationWithoutTenant(ctx context.Context, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, testTenants.emptyTenant(), tc.currentScopes, req, resp)
}

func (tc *testContext) runCustomOperation(ctx context.Context, tenant string, scopes []string, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	token, err := jwtbuilder.Do(tenant, scopes)
	if err != nil {
		return errors.Wrap(err, "while building JWT token")
	}

	cli := gql.NewAuthorizedGraphQLClient(token)
	return retrier.DoOnTemporaryConnectionProblems("DirectorAPITestContext", func() error { return cli.Run(ctx, req, &m) })
}

// resultMapperFor returns generic object that can be passed to Run method for storing response.
// In GraphQL, set `result` alias for your query
func resultMapperFor(target interface{}) genericGQLResponse {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		panic("target has to be a pointer")
	}
	return genericGQLResponse{
		Result: target,
	}
}

type genericGQLResponse struct {
	Result interface{} `json:"result"`
}
