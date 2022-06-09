package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationTemplateInput) *graphql.FormationTemplate {
	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(in)
	require.NoError(t, err)
	createRequest := FixCreateFormationTemplateRequest(formationTemplateInputGQLString)

	formationTemplate := graphql.FormationTemplate{}
	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, createRequest, &formationTemplate))
	require.NotEmpty(t, formationTemplate.ID)

	return &formationTemplate
}

func QueryFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationTemplate {
	queryRequest := FixQueryFormationTemplateRequest(id)

	formationTemplate := graphql.FormationTemplate{}
	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, queryRequest, &formationTemplate))
	require.NotEmpty(t, formationTemplate.ID)

	return &formationTemplate
}

func QueryFormationTemplatesWithPageSize(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, pageSize int) *graphql.FormationTemplatePage {
	queryPaginationRequest := FixQueryFormationTemplatesRequestWithPageSize(pageSize)

	var formationTemplates graphql.FormationTemplatePage
	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, queryPaginationRequest, &formationTemplates))
	require.NotEmpty(t, formationTemplates)

	return &formationTemplates
}

func DeleteFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationTemplate {
	deleteRequest := FixDeleteFormationTemplateRequest(id)

	formationTemplate := graphql.FormationTemplate{}
	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, deleteRequest, &formationTemplate))
	require.NotEmpty(t, formationTemplate.ID)

	return &formationTemplate
}

func CleanupFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationTemplate {
	deleteRequest := FixDeleteFormationTemplateRequest(id)

	formationTemplate := graphql.FormationTemplate{}
	err := testctx.Tc.RunOperation(ctx, gqlClient, deleteRequest, &formationTemplate)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &formationTemplate
}
