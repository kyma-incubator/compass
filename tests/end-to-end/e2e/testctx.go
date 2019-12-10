package e2e

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"

	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"

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
	Graphqlizer       gql.Graphqlizer
	gqlFieldsProvider gql.GqlFieldsProvider
}

func NewTestContext() (*TestContext, error) {
	return &TestContext{
		Graphqlizer:       gql.Graphqlizer{},
		gqlFieldsProvider: gql.GqlFieldsProvider{},
	}, nil
}

func (tc *TestContext) withRetryOnTemporaryConnectionProblems(risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "TestContext").Warnf("OnRetry: attempts: %d, error: %v", n, err)

	}), retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	}))
}

func (tc *TestContext) RunOperationWithCustomTenant(ctx context.Context, cli *gcli.Client, tenant string, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	req.Header.Set("Tenant", tenant)

	return tc.withRetryOnTemporaryConnectionProblems(func() error { return cli.Run(ctx, req, &m) })
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
