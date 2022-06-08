package tests

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var formationTemplateInput = graphql.FormationTemplateInput{
	Name:                          "test-formation-template",
	ApplicationTypes:              []string{"app-type-1", "app-type-2"},
	RuntimeTypes:                  []string{"runtime-type"},
	MissingArtifactInfoMessage:    "Missing Artifact Message Info: %s",
	MissingArtifactWarningMessage: "Missing Artifact Message Warning: %s",
}

func TestCreateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Log("Create formation template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)

	saveExample(t, createFormationTemplateRequest.Query(), "create formation template")

	t.Log("Check if formation template was created")

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, formationTemplateOutput)
}

func TestDeleteFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	deleteFormationTemplateRequest := fixtures.FixDeleteFormationTemplateRequest(createdFormationRequest.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Log("Delete formation template")
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, deleteFormationTemplateRequest.Query(), "delete formation template")

	t.Log("Check if formation template was deleted")

	getFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(output.ID)
	formationTemplateOutput := graphql.FormationTemplate{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationTemplateRequest, &formationTemplateOutput)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	saveExample(t, getFormationTemplateRequest.Query(), "query formation template")
}

func TestUpdateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	var updatedFormationTemplateInput = graphql.FormationTemplateInput{
		Name:                          "new-name-for-formation-template",
		ApplicationTypes:              []string{"app-type-3", "app-type-4"},
		RuntimeTypes:                  []string{"runtime-type-2"},
		MissingArtifactInfoMessage:    "missing info message %s",
		MissingArtifactWarningMessage: "missing warning message %s",
	}

	updatedFormationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(updatedFormationTemplateInput)
	require.NoError(t, err)

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	updateFormationTemplateRequest := fixtures.FixUpdateFormationTemplateRequest(createdFormationRequest.ID, updatedFormationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Log("Update formation template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateFormationTemplateRequest.Query(), "update formation template")

	t.Log("Check if formation template was updated")

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &updatedFormationTemplateInput, formationTemplateOutput)
}

func TestQueryFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	queryFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(createdFormationRequest.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Log("Query formation template")
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, queryFormationTemplateRequest.Query(), "query formation template")

	t.Log("Check if formation template was received")

	assertions.AssertFormationTemplate(t, &formationTemplateInput, &output)
}

func TestQueryFormationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	secondFormationInput := graphql.FormationTemplateInput{
		Name:                          "test-formation-template-2",
		ApplicationTypes:              []string{"app-type-3", "app-type-3"},
		RuntimeTypes:                  []string{"runtime-type-2"},
		MissingArtifactInfoMessage:    "Missing Artifact Message Info 2: %s",
		MissingArtifactWarningMessage: "Missing Artifact Message Warning 2: %s",
	}

	// Get current state
	first := 100
	currentFormationPage := fixtures.QueryFormationTemplatesWithPageSize(t, ctx, certSecuredGraphQLClient, first)

	createdFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
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
