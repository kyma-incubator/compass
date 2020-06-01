package gateway_integration

import (
	"context"
	"reflect"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/director/pkg/retrier"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var tc *TestContext

func init() {
	var err error
	tc, err = NewTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

// TestContext contains dependencies that help executing tests
type TestContext struct {
	Graphqlizer       graphqlizer.Graphqlizer
	gqlFieldsProvider graphqlizer.GqlFieldsProvider
}

func NewTestContext() (*TestContext, error) {
	return &TestContext{
		Graphqlizer:       graphqlizer.Graphqlizer{},
		gqlFieldsProvider: graphqlizer.GqlFieldsProvider{},
	}, nil
}

func (tc *TestContext) RunOperationWithCustomTenant(ctx context.Context, cli *gcli.Client, tenant string, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	req.Header.Set("Tenant", tenant)

	return retrier.DoOnTemporaryConnectionProblems("GatewayIntegrationTestContext", func() error { return cli.Run(ctx, req, &m) })
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
