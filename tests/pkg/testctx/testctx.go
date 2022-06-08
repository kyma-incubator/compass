package testctx

import (
	"context"
	"reflect"
	"strings"
	"time"

	gqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var Tc *TestContext

// TestContext contains dependencies that help executing tests
type TestContext struct {
	Graphqlizer       gqlizer.Graphqlizer
	GQLFieldsProvider gqlizer.GqlFieldsProvider
}

func init() {
	var err error
	Tc, err = NewTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

func NewTestContext() (*TestContext, error) {
	return &TestContext{
		Graphqlizer:       gqlizer.Graphqlizer{},
		GQLFieldsProvider: gqlizer.GqlFieldsProvider{},
	}, nil
}

func (tc *TestContext) NewOperation(ctx context.Context) *Operation {
	return &Operation{
		ctx:         ctx,
		tenant:      tenant.TestTenants.GetDefaultTenantID(),
		queryParams: map[string]string{},
	}
}

type Operation struct {
	ctx context.Context

	tenant      string
	queryParams map[string]string
}

func (o *Operation) WithTenant(tenant string) *Operation {
	o.tenant = tenant
	return o
}

func (o *Operation) Run(req *gcli.Request, cli *gcli.Client, resp interface{}) error {
	m := resultMapperFor(&resp)
	req.Header.Set("Tenant", o.tenant)

	return withRetryOnTemporaryConnectionProblems(func() error {
		return cli.Run(o.ctx, req, &m)
	})
}

func (tc *TestContext) RunOperation(ctx context.Context, cli *gcli.Client, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).Run(req, cli, resp)
}

func (tc *TestContext) RunOperationWithoutTenant(ctx context.Context, cli *gcli.Client, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithTenant("").Run(req, cli, resp)
}

func (tc *TestContext) RunOperationWithCustomTenant(ctx context.Context, cli *gcli.Client, tenant string, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithTenant(tenant).Run(req, cli, resp)
}

func withRetryOnTemporaryConnectionProblems(risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "TestContext").Warnf("OnRetry: attempts: %d, error: %v", n, err)

	}), retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	}))
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
