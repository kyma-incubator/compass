package e2e

import (
	"context"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/common"

	"github.com/sirupsen/logrus"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const defaultTenant = "2a1502ba-aded-11e9-a2a3-2a2ae2dbcce4"

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
	Graphqlizer       common.Graphqlizer
	gqlFieldsProvider common.GqlFieldsProvider
	currentScopes     []string
	cli               *gcli.Client
}

func NewTestContext() (*TestContext, error) {
	scopesStr := os.Getenv("ALL_SCOPES")

	currentScopes := strings.Split(scopesStr, " ")

	bearerToken, err := jwtbuilder.Do(defaultTenant, currentScopes)
	if err != nil {
		return nil, errors.Wrap(err, "while building JWT token")
	}

	return &TestContext{
		Graphqlizer:       common.Graphqlizer{},
		gqlFieldsProvider: common.GqlFieldsProvider{},
		currentScopes:     currentScopes,
		cli:               common.NewAuthorizedGraphQLClient(bearerToken),
	}, nil
}

func (tc *TestContext) RunOperation(ctx context.Context, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	return tc.withRetryOnTemporaryConnectionProblems(func() error {
		return tc.cli.Run(ctx, req, &m)
	})
}

func (tc *TestContext) withRetryOnTemporaryConnectionProblems(risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "TestContext").Warnf("OnRetry: attempts: %d, error: %v", n, err)

	}), retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	}))
}

func (tc *TestContext) RunOperationWithCustomTenant(ctx context.Context, tenant string, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, tenant, tc.currentScopes, req, resp)
}

func (tc *TestContext) RunOperationWithCustomScopes(ctx context.Context, scopes []string, req *gcli.Request, resp interface{}) error {
	return tc.runCustomOperation(ctx, defaultTenant, scopes, req, resp)
}

func (tc *TestContext) runCustomOperation(ctx context.Context, tenant string, scopes []string, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)

	req.Header.Set("Tenant", tenant)

	token, err := jwtbuilder.Do(tenant, scopes)
	if err != nil {
		return errors.Wrap(err, "while building JWT token")
	}

	cli := common.NewAuthorizedGraphQLClient(token)
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
