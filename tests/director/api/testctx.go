package api

import (
	"context"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	gqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var tc *testContext

// testContext contains dependencies that help executing tests
type testContext struct {
	graphqlizer       gqlizer.Graphqlizer
	gqlFieldsProvider gqlizer.GqlFieldsProvider
	currentScopes     []string
}

const defaultScopes = "runtime:write application:write tenant:read label_definition:write integration_system:write application:read runtime:read label_definition:read integration_system:read health_checks:read application_template:read application_template:write eventing:manage automatic_scenario_assignment:read automatic_scenario_assignment:write"

func newTestContext() (*testContext, error) {
	scopesStr := os.Getenv("ALL_SCOPES")
	if scopesStr == "" {
		scopesStr = defaultScopes
	}

	currentScopes := strings.Split(scopesStr, " ")

	return &testContext{
		graphqlizer:       gqlizer.Graphqlizer{},
		gqlFieldsProvider: gqlizer.GqlFieldsProvider{},
		currentScopes:     currentScopes,
	}, nil
}

func (tc *testContext) NewOperation(ctx context.Context) *Operation {
	return &Operation{
		ctx:    ctx,
		tenant: testTenants.GetDefaultTenantID(),
		queryParams: map[string]string{},
		scopes:   tc.currentScopes,
		consumer: &jwtbuilder.Consumer{},
	}
}

type Operation struct {
	ctx context.Context

	tenant      string
	queryParams map[string]string
	scopes      []string
	consumer    *jwtbuilder.Consumer
}

func (o *Operation) WithTenant(tenant string) *Operation {
	o.tenant = tenant
	return o
}

func (o *Operation) WithScopes(scopes []string) *Operation {
	o.scopes = scopes
	return o
}

func (o *Operation) WithConsumer(consumer *jwtbuilder.Consumer) *Operation {
	o.consumer = consumer
	return o
}

func (o *Operation) WithQueryParam(key, value string) *Operation {
	o.queryParams[key] = value
	return o
}

func (o *Operation) WithQueryParams(queryParams map[string]string) *Operation {
	o.queryParams = queryParams
	return o
}

func (o *Operation) Run(req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	token, err := jwtbuilder.Build(o.tenant, o.scopes, o.consumer)
	if err != nil {
		return errors.Wrap(err, "while building JWT token")
	}

	url, err := url.Parse(gql.GetDirectorGraphQLURL())
	if err != nil {
		return err
	}

	query := url.Query()
	for key, val := range o.queryParams {
		query.Set(key, val)
	}
	url.RawQuery = query.Encode()

	cli := gql.NewAuthorizedGraphQLClientWithCustomURL(token, url.String())

	return withRetryOnTemporaryConnectionProblems(func() error {
		return cli.Run(o.ctx, req, &m)
	})
}

func (tc *testContext) RunOperation(ctx context.Context, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).Run(req, resp)
}

func (tc *testContext) RunOperationWithCustomTenant(ctx context.Context, tenant string, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithTenant(tenant).Run(req, resp)
}

func (tc *testContext) RunOperationWithCustomScopes(ctx context.Context, scopes []string, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithScopes(scopes).Run(req, resp)
}

func (tc *testContext) RunOperationWithoutTenant(ctx context.Context, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithTenant(testTenants.emptyTenant()).Run(req, resp)
}

func withRetryOnTemporaryConnectionProblems(risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "testContext").Warnf("OnRetry: attempts: %d, error: %v", n, err)

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
