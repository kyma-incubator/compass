package service_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestContextProvider_ForRequest(t *testing.T) {
	rq, err := http.NewRequest("ANY_METHOD", "", nil)
	require.NoError(t, err)
	rq.Header.Set(gqlcli.AuthorizationHeaderKey, "foo")

	gqlCli := gqlcli.NewAuthorizedGraphQLClient("", rq)
	app := graphql.ApplicationExt{Application: graphql.Application{ID: "app-id"}}
	expected := service.RequestContext{
		AppID:          app.ID,
		DirectorClient: director.NewClient(gqlCli, &graphqlizer.Graphqlizer{}, &graphqlizer.GqlFieldsProvider{}),
	}

	testCases := []struct {
		Name                string
		InputRequestContext context.Context
		ExpectedResult      *service.RequestContext
		ExpectedErrMessage  string
	}{
		{
			Name:                "Success",
			InputRequestContext: fixContext(&app, gqlCli),
			ExpectedResult:      &expected,
			ExpectedErrMessage:  "",
		},
		{
			Name:                "Error - Load App Details",
			InputRequestContext: fixContext(nil, gqlCli),
			ExpectedResult:      nil,
			ExpectedErrMessage:  "while loading Application details from context: cannot read Application details from context",
		},
		{
			Name:                "Error - Load GraphQL client",
			InputRequestContext: fixContext(&app, nil),
			ExpectedResult:      nil,
			ExpectedErrMessage:  "while loading GraphQL client from context: cannot read GraphQL client from context",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			req := rq.WithContext(testCase.InputRequestContext)
			require.NoError(t, err)

			// when
			reqProvider := service.NewRequestContextProvider()

			result, err := reqProvider.ForRequest(req)

			// then
			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				require.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			require.NotNil(t, result)
			require.NotNil(t, testCase.ExpectedResult)
			assert.Equal(t, *testCase.ExpectedResult, result)
		})
	}
}

func fixContext(app *graphql.ApplicationExt, gqlCli gqlcli.GraphQLClient) context.Context {
	ctx := context.TODO()

	if app != nil {
		ctx = appdetails.SaveToContext(ctx, *app)
	}

	if gqlCli != nil {
		ctx = gqlcli.SaveToContext(ctx, gqlCli)
	}

	return ctx
}
