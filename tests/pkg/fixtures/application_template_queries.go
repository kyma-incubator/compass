package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateApplicationTemplateFromInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, input graphql.ApplicationTemplateInput) (graphql.ApplicationTemplate, error) {
	appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(input)
	require.NoError(t, err)

	req := FixCreateApplicationTemplateRequest(appTemplate)
	appTpl := graphql.ApplicationTemplate{}

	return appTpl, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &appTpl)
}

func CreateApplicationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, name string) (graphql.ApplicationTemplate, error) {
	return CreateApplicationTemplateFromInput(t, ctx, gqlClient, tenant, FixApplicationTemplate(name))
}

func GetApplicationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.ApplicationTemplate {
	req := FixApplicationTemplateRequest(id)
	appTpl := graphql.ApplicationTemplate{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &appTpl)
	require.NoError(t, err)
	return appTpl
}

func DeleteApplicationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	if id == "" {
		return
	}
	req := FixDeleteApplicationTemplateRequest(id)

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}

func CleanupApplicationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, appTemplate graphql.ApplicationTemplate) {
	if appTemplate.ID == "" {
		return
	}
	req := FixDeleteApplicationTemplateRequest(appTemplate.ID)

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)
}
