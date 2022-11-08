package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "create-formation-template-name"
	formationTemplate := fixtures.FixFormationTemplate(formationTemplateName)

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplate)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)

	saveExample(t, createFormationTemplateRequest.Query(), "create formation template")

	t.Logf("Check if formation template with name %q was created", formationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplate, formationTemplateOutput)
}

func TestDeleteFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "delete-formation-template-name"
	formationTemplate := fixtures.FixFormationTemplate(formationTemplateName)

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplate)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	deleteFormationTemplateRequest := fixtures.FixDeleteFormationTemplateRequest(createdFormationRequest.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Delete formation template with name %q", formationTemplateName)
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, deleteFormationTemplateRequest.Query(), "delete formation template")

	t.Logf("Check if formation template with name %q and id %q was deleted", formationTemplateName, createdFormationRequest.ID)

	getFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(output.ID)
	formationTemplateOutput := graphql.FormationTemplate{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationTemplateRequest, &formationTemplateOutput)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	saveExample(t, getFormationTemplateRequest.Query(), "query formation template")
}

func TestUpdateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	createdFormationTemplateName := "created-formation-template-name"
	createdFormationTemplate := fixtures.FixFormationTemplate(createdFormationTemplateName)

	var updatedFormationTemplateInput = graphql.FormationTemplateInput{
		Name:                   "updated-formation-template-name",
		ApplicationTypes:       []string{"app-type-3", "app-type-4"},
		RuntimeType:            str.Ptr("runtime-type-2"),
		RuntimeTypeDisplayName: "test-display-name-2",
		RuntimeArtifactKind:    graphql.ArtifactTypeServiceInstance,
	}

	updatedFormationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(updatedFormationTemplateInput)
	require.NoError(t, err)

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationTemplate)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	updateFormationTemplateRequest := fixtures.FixUpdateFormationTemplateRequest(createdFormationRequest.ID, updatedFormationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Update formation template with name %q and id %q", createdFormationTemplateName, createdFormationRequest.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateFormationTemplateRequest.Query(), "update formation template")

	t.Logf("Check if formation template with id %q and old name %q (new name %q) was updated", createdFormationRequest.ID, createdFormationTemplateName, updatedFormationTemplateInput.Name)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &updatedFormationTemplateInput, formationTemplateOutput)
}

func TestQueryFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "query-formation-template-name"
	formationTemplate := fixtures.FixFormationTemplate(formationTemplateName)

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplate)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	queryFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(createdFormationRequest.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Query formation template with name %q and id %q", formationTemplate.Name, createdFormationRequest.ID)
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, queryFormationTemplateRequest.Query(), "query formation template")

	t.Logf("Check if formation template with name %q and ID %q was received", formationTemplateName, createdFormationRequest.ID)

	assertions.AssertFormationTemplate(t, &formationTemplate, &output)
}

func TestQueryFormationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "delete-formation-template-name"
	formationTemplate := fixtures.FixFormationTemplate(formationTemplateName)

	secondFormationInput := graphql.FormationTemplateInput{
		Name:                   "test-formation-template-2",
		ApplicationTypes:       []string{"app-type-3", "app-type-5"},
		RuntimeType:            str.Ptr("runtime-type-2"),
		RuntimeTypeDisplayName: "test-display-name-2",
		RuntimeArtifactKind:    graphql.ArtifactTypeServiceInstance,
	}

	// Get current state
	first := 100
	currentFormationPage := fixtures.QueryFormationTemplatesWithPageSize(t, ctx, certSecuredGraphQLClient, first)

	createdFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplate)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationTemplate.ID)
	secondCreatedFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, secondFormationInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, secondCreatedFormationTemplate.ID)

	var output graphql.FormationTemplatePage
	queryFormationTemplatesRequest := fixtures.FixQueryFormationTemplatesRequestWithPageSize(first)

	// WHEN
	t.Log("Query formation templates")
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryFormationTemplatesRequest, &output)

	//THEN
	require.NotEmpty(t, output)
	require.NoError(t, err)
	assert.Equal(t, currentFormationPage.TotalCount+2, output.TotalCount)

	saveExample(t, queryFormationTemplatesRequest.Query(), "query formation templates")

	t.Log("Check if formation templates are in received slice")

	assert.Subset(t, output.Data, []*graphql.FormationTemplate{
		createdFormationTemplate,
		secondCreatedFormationTemplate,
	})
}
