package fixtures

import (
	"context"
	"fmt"
	"testing"

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
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, createRequest, &formationTemplate))
	require.NotEmpty(t, formationTemplate.ID)

	return &formationTemplate
}

func CreateFormationTemplateExpectError(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationTemplateInput) {
	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(in)
	require.NoError(t, err)
	createRequest := FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	formationTemplate := graphql.FormationTemplate{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, createRequest, &formationTemplate)

	// In case of successfully created formation template the require statement terminates the test case and the ID is not returned. Should clean up the formation template here.
	if formationTemplate.ID != "" {
		defer CleanupFormationTemplate(t, ctx, gqlClient, formationTemplate.ID)
	}

	require.Error(t, err)
	fmt.Println("Error: ", err.Error())
}
func CreateFormationTemplateWithoutInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, formationTemplateName, runtimeType string, applicationTypes []string, runtimeArtifactKind graphql.ArtifactType) graphql.FormationTemplate {
	formationTmplInput := graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: &formationTemplateName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
	}

	formationTmplGQLInput, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTmplInput)
	require.NoError(t, err)
	formationTmplRequest := FixCreateFormationTemplateRequest(formationTmplGQLInput)

	ft := graphql.FormationTemplate{}
	t.Logf("Creating formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, formationTmplRequest, &ft)
	require.NoError(t, err)
	return ft
}

func CreateFormationTemplateWithLeadingProductIDs(t *testing.T, ctx context.Context, gqlClient *gcli.Client, formationTemplateName, runtimeType string, applicationTypes []string, runtimeArtifactKind graphql.ArtifactType, leadingProductIDs []string) graphql.FormationTemplate {
	formationTmplInput := FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, runtimeType, applicationTypes, runtimeArtifactKind, leadingProductIDs)

	formationTmplGQLInput, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTmplInput)
	require.NoError(t, err)
	formationTmplRequest := FixCreateFormationTemplateRequest(formationTmplGQLInput)

	ft := graphql.FormationTemplate{}
	t.Logf("Creating formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, formationTmplRequest, &ft)
	require.NoError(t, err)
	return ft
}

func QueryFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationTemplate {
	queryRequest := FixQueryFormationTemplateRequest(id)

	formationTemplate := graphql.FormationTemplate{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, queryRequest, &formationTemplate))
	require.NotEmpty(t, formationTemplate.ID)

	return &formationTemplate
}

func QueryFormationTemplatesWithPageSize(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, pageSize int) *graphql.FormationTemplatePage {
	queryPaginationRequest := FixQueryFormationTemplatesRequestWithPageSize(pageSize)

	var formationTemplates graphql.FormationTemplatePage
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, queryPaginationRequest, &formationTemplates))
	require.NotEmpty(t, formationTemplates)

	return &formationTemplates
}

func QueryFormationTemplatesWithPageSizeAndTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, pageSize int, tenantID string) *graphql.FormationTemplatePage {
	queryPaginationRequest := FixQueryFormationTemplatesRequestWithPageSize(pageSize)

	var formationTemplates graphql.FormationTemplatePage
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, queryPaginationRequest, &formationTemplates))
	require.NotEmpty(t, formationTemplates)

	return &formationTemplates
}

func CleanupFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationTemplate {
	deleteRequest := FixDeleteFormationTemplateRequest(id)

	formationTemplate := graphql.FormationTemplate{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteRequest, &formationTemplate)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &formationTemplate
}

func UpdateFormationTemplateExpectError(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string, in graphql.FormationTemplateInput) {
	updatedFormationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(in)
	require.NoError(t, err)

	updateFormationTemplateRequest := FixUpdateFormationTemplateRequest(id, updatedFormationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, updateFormationTemplateRequest, &output)

	//THEN
	require.Error(t, err)
	fmt.Println("Error: ", err.Error())
}
