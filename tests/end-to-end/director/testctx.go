package director

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"os"
	"reflect"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const defaultTenant = "2a1502ba-aded-11e9-a2a3-2a2ae2dbcce4"

var tc *testContext

func init() {
	var err error
	tc, err = newTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

type scopeProvider interface {
	GetRequiredScopes(path string) ([]string, error)
	GetAllScopes() ([]string, error)
}

// testContext contains dependencies that help executing tests
type testContext struct {
	graphqlizer       graphqlizer
	gqlFieldsProvider gqlFieldsProvider
	scopeProvider           scopeProvider
	currentScopes     []string
	cli               *gcli.Client
}

func newTestContext() (*testContext, error) {
	scopesCfgPath := os.Getenv("SCOPES_CONFIGURATION_FILE")
	scopeProvider := scope.NewProvider(scopesCfgPath)

	err := scopeProvider.Load()
	if err != nil {
		return nil, errors.Wrap(err, "while loading config for scopes")
	}

	currentScopes, err := scopeProvider.GetAllScopes()
	if err != nil {
		return nil, err
	}

	bearerToken, err := jwtbuilder.Do(defaultTenant, currentScopes)
	if err != nil {
		return nil, errors.Wrap(err, "while building JWT token")
	}

	return &testContext{
		graphqlizer:       graphqlizer{},
		gqlFieldsProvider: gqlFieldsProvider{},
		scopeProvider: scopeProvider,
		currentScopes:     currentScopes,
		cli:               newAuthorizedGraphQLClient(bearerToken),
	}, nil
}

func (tc *testContext) RunOperation(ctx context.Context, req *gcli.Request, resp interface{}) error {
	// TODO: Remove tenant header after implementing https://github.com/kyma-incubator/compass/issues/288
	if req.Header["Tenant"] == nil {
		req.Header["Tenant"] = []string{defaultTenant}
	}

	m := resultMapperFor(&resp)
	return tc.cli.Run(ctx, req, &m)
}

func (tc *testContext) RunOperationWithCustomTenant(ctx context.Context, tenant string, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, tenant, tc.currentScopes, req, resp)
}

func (tc *testContext) RunOperationWithCustomScopes(ctx context.Context, scopes []string, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, defaultTenant, scopes, req, resp)
}

func (tc *testContext) runCustomOperation(ctx context.Context, tenant string, scopes []string, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	// TODO: Remove tenant header after implementing https://github.com/kyma-incubator/compass/issues/288
	req.Header["Tenant"] = []string{tenant}

	token, err := jwtbuilder.Do(tenant, scopes)
	if err != nil {
		return errors.Wrap(err, "while building JWT token")
	}

	cli := newAuthorizedGraphQLClient(token)
	return cli.Run(ctx, req, &m)
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
