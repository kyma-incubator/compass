package testctx

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	gqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/compass/tests/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

var Tc *TestContext

// TestContext contains dependencies that help executing tests
type TestContext struct {
	Graphqlizer       gqlizer.Graphqlizer
	GQLFieldsProvider gqlizer.GqlFieldsProvider
	CurrentScopes     []string
}

const defaultScopes = "runtime:write application:write tenant:read label_definition:write integration_system:write application:read runtime:read label_definition:read integration_system:read health_checks:read application_template:read application_template:write eventing:manage automatic_scenario_assignment:read automatic_scenario_assignment:write"

func Init() {
	var err error
	Tc, err = NewTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

func NewTestContext() (*TestContext, error) {
	scopesStr := os.Getenv("ALL_SCOPES")
	if scopesStr == "" {
		scopesStr = defaultScopes
	}

	currentScopes := strings.Split(scopesStr, " ")

	return &TestContext{
		Graphqlizer:       gqlizer.Graphqlizer{},
		GQLFieldsProvider: gqlizer.GqlFieldsProvider{},
		CurrentScopes:     currentScopes,
	}, nil
}

func (tc *TestContext) NewOperation(ctx context.Context) *Operation {
	return &Operation{
		ctx:         ctx,
		tenant:      pkg.TestTenants.GetDefaultTenantID(),
		queryParams: map[string]string{},
		scopes:      tc.CurrentScopes,
		consumer:    &jwtbuilder.Consumer{},
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

func (o *Operation) Run(req *gcli.Request, cli *gcli.Client, resp interface{}) error {
	m := resultMapperFor(&resp)

	req.Header.Set("Tenant", o.tenant)

	return withRetryOnTemporaryConnectionProblems(func() error {
		return cli.Run(o.ctx, req, &m)
	})
}

func (o *Operation) AddQueryParamsToURL() (*url.URL, error) {
	url, err := url.Parse(gql.GetDirectorGraphQLURL())
	if err != nil {
		return nil, err
	}

	query := url.Query()
	for key, val := range o.queryParams {
		query.Set(key, val)
	}
	url.RawQuery = query.Encode()
	return url, nil
}

func (tc *TestContext) RunOperation(ctx context.Context, cli *gcli.Client, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).Run(req, cli, resp)
}

func (tc *TestContext) RunOperationWithCustomTenant(ctx context.Context, cli *gcli.Client, tenant string, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithTenant(tenant).Run(req, cli, resp)
}

func (tc *TestContext) RunOperationWithCustomScopes(ctx context.Context, cli *gcli.Client, tenant string, scopes []string, req *gcli.Request, resp interface{}) error {
	return tc.NewOperation(ctx).WithTenant(tenant).WithScopes(scopes).Run(req, cli, resp)
}

func (tc *TestContext) RunOperationWithQueryParam(ctx context.Context, key, value, tenant string, req *gcli.Request, resp interface{}) error {
	o := tc.NewOperation(ctx).WithTenant(tenant).WithQueryParam(key, value)

	url, err := o.AddQueryParamsToURL()
	if err != nil {
		return errors.Wrap(err, "while adding query params to URL")
	}

	dexToken, err := idtokenprovider.GetDexToken()
	if err != nil {
		return errors.Wrap(err, "while getting DEX token")
	}

	cli := gql.NewAuthorizedGraphQLClientWithCustomURL(dexToken, url.String())
	return o.Run(req, cli, resp)
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
