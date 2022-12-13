package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const globalFormationTemplateName = "Side-by-side extensibility with Kyma"

var updatedFormationTemplateInput = graphql.FormationTemplateInput{
	Name:                   "updated-formation-template-name",
	ApplicationTypes:       []string{"app-type-3", "app-type-4"},
	RuntimeTypes:           []string{"runtime-type-2"},
	RuntimeTypeDisplayName: "test-display-name-2",
	RuntimeArtifactKind:    graphql.ArtifactTypeServiceInstance,
}

func TestCreateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "create-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)

	saveExample(t, createFormationTemplateRequest.Query(), "create formation template")

	t.Logf("Check if formation template with name %q was created", formationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, formationTemplateOutput)
}

func TestDeleteFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "delete-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	formationTemplateReq := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateReq.ID)

	deleteFormationTemplateRequest := fixtures.FixDeleteFormationTemplateRequest(formationTemplateReq.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Delete formation template with name %q", formationTemplateName)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, deleteFormationTemplateRequest.Query(), "delete formation template")

	t.Logf("Check if formation template with name: %q and ID: %q was deleted", formationTemplateName, formationTemplateReq.ID)

	getFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(output.ID)
	formationTemplateOutput := graphql.FormationTemplate{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, getFormationTemplateRequest, &formationTemplateOutput)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	saveExample(t, getFormationTemplateRequest.Query(), "query formation template")
}

func TestUpdateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	createdFormationTemplateName := "created-formation-template-name"
	createdFormationTemplateInput := fixtures.FixFormationTemplateInput(createdFormationTemplateName)

	updatedFormationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(updatedFormationTemplateInput)
	require.NoError(t, err)

	formationTemplateReq := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateReq.ID)

	updateFormationTemplateRequest := fixtures.FixUpdateFormationTemplateRequest(formationTemplateReq.ID, updatedFormationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Update formation template with name: %q and ID: %q", createdFormationTemplateName, formationTemplateReq.ID)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateFormationTemplateRequest.Query(), "update formation template")

	t.Logf("Check if formation template with ID: %q and old name: %q was successully updated to: %q", formationTemplateReq.ID, createdFormationTemplateName, updatedFormationTemplateInput.Name)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &updatedFormationTemplateInput, formationTemplateOutput)
}

func TestQueryFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "query-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	createdFormationRequest := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationRequest.ID)

	queryFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(createdFormationRequest.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Query formation template with name %q and id %q", formationTemplateName, createdFormationRequest.ID)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, queryFormationTemplateRequest.Query(), "query formation template")

	t.Logf("Check if formation template with name %q and ID %q was received", formationTemplateName, createdFormationRequest.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, &output)
}

func TestQueryFormationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "delete-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)
	runtimeType := "runtime-type-2"
	secondFormationInput := graphql.FormationTemplateInput{
		Name:                   "test-formation-template-2",
		ApplicationTypes:       []string{"app-type-3", "app-type-5"},
		RuntimeTypes:           []string{runtimeType},
		RuntimeTypeDisplayName: "test-display-name-2",
		RuntimeArtifactKind:    graphql.ArtifactTypeServiceInstance,
	}

	// Get current state
	first := 100
	currentFormationTemplatePage := fixtures.QueryFormationTemplatesWithPageSize(t, ctx, certSecuredGraphQLClient, first)

	createdFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationTemplate.ID)
	secondCreatedFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, secondFormationInput)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, secondCreatedFormationTemplate.ID)

	var output graphql.FormationTemplatePage
	queryFormationTemplatesRequest := fixtures.FixQueryFormationTemplatesRequestWithPageSize(first)

	// WHEN
	t.Log("Query formation templates")
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryFormationTemplatesRequest, &output)

	//THEN
	require.NotEmpty(t, output)
	require.NoError(t, err)
	assert.Equal(t, currentFormationTemplatePage.TotalCount+2, output.TotalCount)

	saveExample(t, queryFormationTemplatesRequest.Query(), "query formation templates")

	t.Log("Check if formation templates are in received slice")

	assert.Subset(t, output.Data, []*graphql.FormationTemplate{
		createdFormationTemplate,
		secondCreatedFormationTemplate,
	})
}

func TestTenantScopedFormationTemplates(t *testing.T) {
	ctx := context.Background()
	first := 100

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	scopedFormationTemplateName := "tenant-scoped-formation-template-test"
	scopedFormationTemplateInput := fixtures.FixFormationTemplateInput(scopedFormationTemplateName)

	t.Logf("Create tenant scoped formation template with name: %q", scopedFormationTemplateName)
	scopedFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, directorCertSecuredClient, scopedFormationTemplateInput) // tenant_id is extracted from the subject of the cert
	defer fixtures.CleanupFormationTemplate(t, ctx, directorCertSecuredClient, scopedFormationTemplate.ID)

	assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, scopedFormationTemplate)

	t.Logf("List all formation templates for the tenant in which formation template with name: %q was created and verify that it is visible there", scopedFormationTemplateName)
	formationTemplatePage := fixtures.QueryFormationTemplatesWithPageSize(t, ctx, directorCertSecuredClient, first)

	assert.Subset(t, formationTemplatePage.Data, []*graphql.FormationTemplate{
		scopedFormationTemplate,
	})

	t.Logf("List all formation templates for some other tenant in which formation template with name: %q was NOT created and verify that it is NOT visible there", scopedFormationTemplateName)
	formationTemplatePageForOtherTenant := fixtures.QueryFormationTemplatesWithPageSizeAndTenant(t, ctx, directorCertSecuredClient, first, tenant.TestTenants.GetDefaultTenantID())

	assert.NotContains(t, formationTemplatePageForOtherTenant.Data, []*graphql.FormationTemplate{
		scopedFormationTemplate,
	})

	var globalFormationTemplateID string
	for _, ft := range formationTemplatePage.Data {
		if ft.Name == globalFormationTemplateName {
			globalFormationTemplateID = ft.ID
			break
		}
	}
	require.NotEmpty(t, globalFormationTemplateID)

	t.Logf("Verify that tenant scoped call can NOT delete global formation template with name: %q", globalFormationTemplateName)
	deleteFormationTemplateRequest := fixtures.FixDeleteFormationTemplateRequest(globalFormationTemplateID)
	output := graphql.FormationTemplate{}

	err := testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, deleteFormationTemplateRequest, &output) // tenant_id is extracted from the subject of the cert
	require.Error(t, err)
	require.Contains(t, err.Error(), "Owner access is needed for resource modification")

	t.Logf("Verify that tenant scoped call can NOT update global formation template with name: %q", globalFormationTemplateName)

	updatedFormationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(updatedFormationTemplateInput)
	require.NoError(t, err)

	updateFormationTemplateRequest := fixtures.FixUpdateFormationTemplateRequest(globalFormationTemplateID, updatedFormationTemplateInputGQLString)
	updateOutput := graphql.FormationTemplate{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateFormationTemplateRequest, &updateOutput) // tenant_id is extracted from the subject of the cert
	require.Error(t, err)
	require.Contains(t, err.Error(), "Owner access is needed for resource modification")
}
