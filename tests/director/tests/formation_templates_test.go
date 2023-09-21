package tests

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	CreateFormationTemplateCategory = "create formation template"
	globalFormationTemplateName     = "Side-by-Side Extensibility with Kyma"
)

var (
	runtimeType                   = "runtimeTypeTest"
	serviceInstanceArtifactType   = graphql.ArtifactTypeServiceInstance
	updatedFormationTemplateInput = graphql.FormationTemplateInput{
		Name:                   "updated-formation-template-name",
		ApplicationTypes:       []string{"app-type-3", "app-type-4"},
		RuntimeTypes:           []string{"runtime-type-2"},
		RuntimeTypeDisplayName: str.Ptr("test-display-name-2"),
		RuntimeArtifactKind:    &serviceInstanceArtifactType,
	}
)

func TestCreateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appType := "async-app-type-1"
	formationTemplateName := "create-formation-template-name"
	leadingProductIDs := []string{"leading-product-id"}
	formationTemplateInput := fixtures.FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, []string{appType}, []string{"runtimeTypeTest"}, graphql.ArtifactTypeEnvironmentInstance, leadingProductIDs)

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &output)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)

	example.SaveExampleInCustomDir(t, createFormationTemplateRequest.Query(), CreateFormationTemplateCategory, "create formation template")

	t.Logf("Check if formation template with name %q was created", formationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, formationTemplateOutput)
}

func TestCreateAppOnlyFormationTemplate(t *testing.T) {
	ctx := context.Background()

	appOnlyFormationTemplateName := "app-only-formation-template"
	t.Logf("Create formation template with name: %q", appOnlyFormationTemplateName)

	appOnlyFormationTemplateInput := fixtures.FixAppOnlyFormationTemplateInput(appOnlyFormationTemplateName)
	var output graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &output)
	output = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, appOnlyFormationTemplateInput)

	t.Logf("Check if formation template with name %q was created", appOnlyFormationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertAppOnlyFormationTemplate(t, &appOnlyFormationTemplateInput, formationTemplateOutput)

	invalidFormationTemplateWithArtifactKindName := "invalid-formation-template-with-artifact-kind"
	t.Logf("Should fail to create formation template with name: %q", invalidFormationTemplateWithArtifactKindName)

	invalidFormationTemplateWithArtifactKindInput := fixtures.FixInvalidFormationTemplateInputWithRuntimeArtifactKind(invalidFormationTemplateWithArtifactKindName)
	fixtures.CreateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, invalidFormationTemplateWithArtifactKindInput)

	invalidFormationTemplateWithDisplayName := "invalid-formation-template-with-display-name"
	t.Logf("Should fail to create formation template with name: %q", invalidFormationTemplateWithDisplayName)

	invalidFormationTemplateWithDisplayNameInput := fixtures.FixInvalidFormationTemplateInputWithRuntimeTypeDisplayName(invalidFormationTemplateWithDisplayName)
	fixtures.CreateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, invalidFormationTemplateWithDisplayNameInput)

	invalidFormationTemplateWithRuntimeTypesName := "invalid-formation-template-with-runtime-types"
	t.Logf("Should fail to create formation template with name: %q", invalidFormationTemplateWithRuntimeTypesName)

	invalidFormationTemplateWithRuntimeTypesInput := fixtures.FixInvalidFormationTemplateInputWithRuntimeTypes(invalidFormationTemplateWithRuntimeTypesName, runtimeType)
	fixtures.CreateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, invalidFormationTemplateWithRuntimeTypesInput)

	invalidFormationTemplateWithoutArtifactKindName := "invalid-formation-template-without-artifact-kind"
	t.Logf("Should fail to create formation template with name: %q", invalidFormationTemplateWithoutArtifactKindName)

	invalidFormationTemplateWithoutArtifactKindInput := fixtures.FixInvalidFormationTemplateInputWithoutArtifactKind(invalidFormationTemplateWithoutArtifactKindName, runtimeType)
	fixtures.CreateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, invalidFormationTemplateWithoutArtifactKindInput)

	invalidFormationTemplateWithoutDisplayName := "invalid-formation-template-without-display-name"
	t.Logf("Should fail to create formation template with name: %q", invalidFormationTemplateWithoutDisplayName)

	invalidFormationTemplateWithoutDisplayNameInput := fixtures.FixInvalidFormationTemplateInputWithoutDisplayName(invalidFormationTemplateWithoutDisplayName, runtimeType)
	fixtures.CreateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, invalidFormationTemplateWithoutDisplayNameInput)
}

func TestCreateFormationTemplateThatSupportsReset(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appType := "formation-app-type1"
	formationTemplateName := "create-formation-template-name"
	leadingProductIDs := []string{"leading-product-id"}
	formationTemplateInput := fixtures.FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, []string{appType}, []string{"runtimeTypeTest"}, graphql.ArtifactTypeEnvironmentInstance, leadingProductIDs)
	supportsReset := false
	formationTemplateInput.SupportsReset = &supportsReset

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &output)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)

	example.SaveExampleInCustomDir(t, createFormationTemplateRequest.Query(), CreateFormationTemplateCategory, "create formation template with reset")

	t.Logf("Check if formation template with name %q was created", formationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, formationTemplateOutput)

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	formationName := "test-formation"
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, formationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, formationName, &formationTemplateName)

	expectedErrorMsg := "graphql: The operation is not allowed [reason=formation template \"create-formation-template-name\" does not support resetting]"
	t.Logf("Resynchronize formation %q with reset should fail", formation.Name)
	resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequestWithResetOption(formation.ID, reset)
	example.SaveExample(t, resynchronizeReq.Query(), "resynchronize formation notifications with reset")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, resynchronizeReq, &formation)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), expectedErrorMsg)

	t.Logf("Resynchronize formation %q without reset should succeed", formation.Name)
	resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequestWithResetOption(formation.ID, dontReset)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, resynchronizeReq, &formation)
	require.Nil(t, err)
}

func TestCreateFormationTemplateWithFormationLifecycleWebhook(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "create-formation-template-with-webhook-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	webhookSyncMode := graphql.WebhookModeSync

	formationTemplateInput.Webhooks = []*graphql.WebhookInput{
		{
			Type: graphql.WebhookTypeFormationLifecycle,
			Mode: &webhookSyncMode,
			URL:  str.Ptr("http://localhost:6439/"),
		},
	}

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &output)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)
	assertions.AssertFormationTemplate(t, &formationTemplateInput, &output)

	example.SaveExampleInCustomDir(t, createFormationTemplateRequest.Query(), CreateFormationTemplateCategory, "create formation template with webhooks")

	t.Logf("Check if formation template with name %q was created", formationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, formationTemplateOutput)
}

func TestDeleteFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "delete-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	var formationTemplateReq graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplateReq)
	formationTemplateReq = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	deleteFormationTemplateRequest := fixtures.FixDeleteFormationTemplateRequest(formationTemplateReq.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Delete formation template with name %q", formationTemplateName)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	example.SaveExample(t, deleteFormationTemplateRequest.Query(), "delete formation template")

	t.Logf("Check if formation template with name: %q and ID: %q was deleted", formationTemplateName, formationTemplateReq.ID)

	getFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(output.ID)
	formationTemplateOutput := graphql.FormationTemplate{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, getFormationTemplateRequest, &formationTemplateOutput)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	example.SaveExample(t, getFormationTemplateRequest.Query(), "query formation template")
}

func TestUpdateFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	createdFormationTemplateName := "created-formation-template-name"
	createdFormationTemplateInput := fixtures.FixFormationTemplateInput(createdFormationTemplateName)

	updatedFormationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(updatedFormationTemplateInput)
	require.NoError(t, err)

	var formationTemplateReq graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplateReq)
	formationTemplateReq = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, createdFormationTemplateInput)

	updateFormationTemplateRequest := fixtures.FixUpdateFormationTemplateRequest(formationTemplateReq.ID, updatedFormationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Update formation template with name: %q and ID: %q", createdFormationTemplateName, formationTemplateReq.ID)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	example.SaveExample(t, updateFormationTemplateRequest.Query(), "update formation template")

	t.Logf("Check if formation template with ID: %q and old name: %q was successully updated to: %q", formationTemplateReq.ID, createdFormationTemplateName, updatedFormationTemplateInput.Name)
	formationTemplateID := output.ID

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateID)

	assertions.AssertFormationTemplate(t, &updatedFormationTemplateInput, formationTemplateOutput)
}

func TestUpdateAppOnlyFormationTemplate(t *testing.T) {
	ctx := context.Background()

	appOnlyFormationTemplateName := "app-only-formation-template"
	t.Logf("Create formation template with name: %q", appOnlyFormationTemplateName)

	appOnlyFormationTemplateInput := fixtures.FixAppOnlyFormationTemplateInput(appOnlyFormationTemplateName)
	var output graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &output)
	output = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, appOnlyFormationTemplateInput)
	formationTemplateID := output.ID

	t.Logf("Check if formation template with name %q was created", appOnlyFormationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateID)

	assertions.AssertAppOnlyFormationTemplate(t, &appOnlyFormationTemplateInput, formationTemplateOutput)

	t.Log("Should fail to update formation template by adding runtime artifact kind only")

	invalidFormationTemplateWithArtifactKindInput := fixtures.FixInvalidFormationTemplateInputWithRuntimeArtifactKind(appOnlyFormationTemplateName)
	fixtures.UpdateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, formationTemplateID, invalidFormationTemplateWithArtifactKindInput)

	t.Log("Should fail to update formation template by adding runtime type display name only")

	invalidFormationTemplateWithDisplayNameInput := fixtures.FixInvalidFormationTemplateInputWithRuntimeTypeDisplayName(appOnlyFormationTemplateName)
	fixtures.UpdateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, formationTemplateID, invalidFormationTemplateWithDisplayNameInput)

	t.Log("Should fail to update formation template by adding runtime types only")

	invalidFormationTemplateWithRuntimeTypesInput := fixtures.FixInvalidFormationTemplateInputWithRuntimeTypes(appOnlyFormationTemplateName, runtimeType)
	fixtures.UpdateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, formationTemplateID, invalidFormationTemplateWithRuntimeTypesInput)

	t.Log("Should fail to update formation template by adding runtime artifact kind and runtime types only")

	invalidFormationTemplateWithoutArtifactKindInput := fixtures.FixInvalidFormationTemplateInputWithoutDisplayName(appOnlyFormationTemplateName, runtimeType)
	fixtures.UpdateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, formationTemplateID, invalidFormationTemplateWithoutArtifactKindInput)

	t.Log("Should fail to update formation template by adding runtime display name and runtime types only")

	invalidFormationTemplateWithoutDisplayNameInput := fixtures.FixInvalidFormationTemplateInputWithoutArtifactKind(appOnlyFormationTemplateName, runtimeType)
	fixtures.UpdateFormationTemplateExpectError(t, ctx, certSecuredGraphQLClient, formationTemplateID, invalidFormationTemplateWithoutDisplayNameInput)
}

func TestModifyFormationTemplateWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "create-formation-template-with-webhook-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	output := graphql.FormationTemplate{}

	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &output)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &output)
	require.NoError(t, err)

	// Add formation template webhook
	webhookSyncMode := graphql.WebhookModeSync

	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"
	url := "http://new-webhook.url"
	webhookInput := &graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeFormationLifecycle,
		Mode:           &webhookSyncMode,
		OutputTemplate: &outputTemplate,
	}
	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(webhookInput)

	require.NoError(t, err)
	addReq := fixtures.FixAddWebhookToFormationTemplateRequest(output.ID, webhookInStr)
	example.SaveExampleInCustomDir(t, addReq.Query(), example.AddWebhookCategory, "add formation template webhook")

	actualWebhook := graphql.Webhook{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addReq, &actualWebhook)
	require.Error(t, err)

	expectedErrorMessage := fmt.Sprintf("does not have access to the parent resource formationTemplate with ID %s]", output.ID)
	require.Contains(t, err.Error(), expectedErrorMessage)

	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, addReq, &actualWebhook)
	require.NoError(t, err)

	assert.NotNil(t, actualWebhook.URL)
	id := actualWebhook.ID
	require.NotNil(t, id)
	assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
	assert.Equal(t, graphql.WebhookTypeFormationLifecycle, actualWebhook.Type)

	t.Run("Get formation template webhooks", func(t *testing.T) {
		updatedFormationTemplate := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)
		assert.Len(t, updatedFormationTemplate.Webhooks, 1)
		assertions.AssertWebhooks(t, []*graphql.WebhookInput{webhookInput}, []graphql.Webhook{*updatedFormationTemplate.Webhooks[0]})
	})

	urlUpdated := "https://test.com"
	t.Run("Update formation template webhook globally", func(t *testing.T) {
		webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: &urlUpdated, Type: graphql.WebhookTypeFormationLifecycle})

		require.NoError(t, err)
		updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)

		var updatedWebhook graphql.Webhook
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateReq, &updatedWebhook)
		require.NoError(t, err)
		assert.NotNil(t, updatedWebhook.URL)
		assert.Equal(t, urlUpdated, *updatedWebhook.URL)
	})

	t.Run("Delete formation template webhook", func(t *testing.T) {
		//GIVEN
		deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
		example.SaveExampleInCustomDir(t, deleteReq.Query(), example.DeleteWebhookCategory, "delete webhook")

		//WHEN
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, urlUpdated, *actualWebhook.URL)
	})
}

func TestQueryFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	formationTemplateName := "query-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	var createdFormationRequest graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &createdFormationRequest)
	createdFormationRequest = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	queryFormationTemplateRequest := fixtures.FixQueryFormationTemplateRequest(createdFormationRequest.ID)
	output := graphql.FormationTemplate{}

	// WHEN
	t.Logf("Query formation template with name %q and id %q", formationTemplateName, createdFormationRequest.ID)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryFormationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	example.SaveExample(t, queryFormationTemplateRequest.Query(), "query formation template")

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
		RuntimeTypeDisplayName: str.Ptr("test-display-name-2"),
		RuntimeArtifactKind:    &serviceInstanceArtifactType,
	}

	// Get current state
	first := 100
	currentFormationTemplatePage := fixtures.QueryFormationTemplatesWithPageSize(t, ctx, certSecuredGraphQLClient, first)

	var createdFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &createdFormationTemplate)
	createdFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)
	var secondCreatedFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &secondCreatedFormationTemplate)
	secondCreatedFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, secondFormationInput)

	var output graphql.FormationTemplatePage
	queryFormationTemplatesRequest := fixtures.FixQueryFormationTemplatesRequestWithPageSize(first)

	// WHEN
	t.Log("Query formation templates")
	err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryFormationTemplatesRequest, &output)

	//THEN
	require.NotEmpty(t, output)
	require.NoError(t, err)
	assert.Equal(t, currentFormationTemplatePage.TotalCount+2, output.TotalCount)

	example.SaveExample(t, queryFormationTemplatesRequest.Query(), "query formation templates")

	t.Log("Check if formation templates are in received slice")

	assert.Subset(t, output.Data, []*graphql.FormationTemplate{
		&createdFormationTemplate,
		&secondCreatedFormationTemplate,
	})

	queryFormationTemplatesRequest = fixtures.FixQueryFormationTemplatesRequestWithNameAndPageSize("test-formation-template-2", first)

	// WHEN
	t.Log("Query formation templates with name filter")
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryFormationTemplatesRequest, &output)

	//THEN
	require.NotEmpty(t, output)
	require.NoError(t, err)
	assert.Equal(t, 1, output.TotalCount)

	example.SaveExample(t, queryFormationTemplatesRequest.Query(), "query formation templates by name")

	t.Log("Check if formation templates are in received slice")

	assert.Subset(t, output.Data, []*graphql.FormationTemplate{
		&secondCreatedFormationTemplate,
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
	var scopedFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, directorCertSecuredClient, &scopedFormationTemplate)
	scopedFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, directorCertSecuredClient, scopedFormationTemplateInput) // tenant_id is extracted from the subject of the cert

	assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, &scopedFormationTemplate)

	t.Logf("List all formation templates for the tenant in which formation template with name: %q was created and verify that it is visible there", scopedFormationTemplateName)
	formationTemplatePage := fixtures.QueryFormationTemplatesWithPageSize(t, ctx, directorCertSecuredClient, first)

	assert.Greater(t, len(formationTemplatePage.Data), 1) // assert that both tenant scoped and global formation templates are visible
	assert.Subset(t, formationTemplatePage.Data, []*graphql.FormationTemplate{
		&scopedFormationTemplate,
	})

	t.Logf("List all formation templates for some other tenant in which formation template with name: %q was NOT created and verify that it is NOT visible there", scopedFormationTemplateName)
	formationTemplatePageForOtherTenant := fixtures.QueryFormationTemplatesWithPageSizeAndTenant(t, ctx, directorCertSecuredClient, first, tenant.TestTenants.GetDefaultTenantID())

	assert.NotEmpty(t, formationTemplatePageForOtherTenant.Data)
	assert.NotContains(t, formationTemplatePageForOtherTenant.Data, []*graphql.FormationTemplate{
		&scopedFormationTemplate,
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

func TestResourceGroupScopedFormationTemplates(t *testing.T) {
	ctx := context.Background()
	first := 100

	resourceGroup := tenant.TestTenants.GetIDByName(t, tenant.TestAtomResourceGroup)

	scopedFormationTemplateName := "resource-group-scoped-formation-template-test"
	scopedFormationTemplateInput := fixtures.FixFormationTemplateInput(scopedFormationTemplateName)

	t.Logf("Create resource group scoped formation template with name: %q", scopedFormationTemplateName)
	var scopedFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplateWithTenant(t, ctx, certSecuredGraphQLClient, resourceGroup, &scopedFormationTemplate)
	scopedFormationTemplate = fixtures.CreateFormationTemplateWithTenant(t, ctx, certSecuredGraphQLClient, resourceGroup, scopedFormationTemplateInput)

	assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, &scopedFormationTemplate)

	t.Logf("List all formation templates for the tenant in which formation template with name: %q was created and verify that it is visible there", scopedFormationTemplateName)
	formationTemplatePage := fixtures.QueryFormationTemplatesWithPageSizeAndTenant(t, ctx, certSecuredGraphQLClient, first, resourceGroup)

	assert.Greater(t, len(formationTemplatePage.Data), 1) // assert that both tenant scoped and global formation templates are visible
	assert.Subset(t, formationTemplatePage.Data, []*graphql.FormationTemplate{
		&scopedFormationTemplate,
	})

	t.Logf("List all formation templates for some other tenant in which formation template with name: %q was NOT created and verify that it is NOT visible there", scopedFormationTemplateName)
	formationTemplatePageForOtherTenant := fixtures.QueryFormationTemplatesWithPageSizeAndTenant(t, ctx, certSecuredGraphQLClient, first, tenant.TestTenants.GetDefaultTenantID())

	assert.NotEmpty(t, formationTemplatePageForOtherTenant.Data)
	assert.NotContains(t, formationTemplatePageForOtherTenant.Data, []*graphql.FormationTemplate{
		&scopedFormationTemplate,
	})

	var globalFormationTemplateID string
	for _, ft := range formationTemplatePage.Data {
		if ft.Name == globalFormationTemplateName {
			globalFormationTemplateID = ft.ID
			break
		}
	}
	require.NotEmpty(t, globalFormationTemplateID)
}

func TestTenantScopedFormationTemplatesWithWebhooks(t *testing.T) {
	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	scopedFormationTemplateName := "tenant-scoped-formation-template-with-webhook-test"
	scopedFormationTemplateInput := fixtures.FixFormationTemplateInput(scopedFormationTemplateName)
	webhookSyncMode := graphql.WebhookModeSync

	t.Logf("Create tenant scoped formation template with name: %q", scopedFormationTemplateName)
	var scopedFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, directorCertSecuredClient, &scopedFormationTemplate)
	scopedFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, directorCertSecuredClient, scopedFormationTemplateInput) // tenant_id is extracted from the subject of the cert

	assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, &scopedFormationTemplate)
	urlUpdated := "http://updated.url"
	webhookInput := &graphql.WebhookInput{
		Type: graphql.WebhookTypeFormationLifecycle,
		Mode: &webhookSyncMode,
		URL:  str.Ptr("http://localhost:6439/"),
	}
	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(webhookInput)
	addReq := fixtures.FixAddWebhookToFormationTemplateRequest(scopedFormationTemplate.ID, webhookInStr)
	example.SaveExampleInCustomDir(t, addReq.Query(), example.AddWebhookCategory, "add formation template webhook")
	t.Run("Add formation template webhook with other tenant should result in unauthorized", func(t *testing.T) {
		actualWebhook := graphql.Webhook{}
		customTenant := tenant.TestTenants.GetIDByName(t, tenant.TestConsumerSubaccount)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, customTenant, addReq, &actualWebhook)
		require.Error(t, err)

		expectedErrorMessage := fmt.Sprintf("does not have access to the parent resource formationTemplate with ID %s]", scopedFormationTemplate.ID)
		require.Contains(t, err.Error(), expectedErrorMessage)
	})
	actualWebhook := graphql.Webhook{}
	t.Run("Add formation template webhook with correct tenant should succeed", func(t *testing.T) {
		err = testctx.Tc.RunOperation(ctx, directorCertSecuredClient, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		id := actualWebhook.ID
		require.NotNil(t, id)
		assert.Equal(t, "http://localhost:6439/", *actualWebhook.URL)
		assert.Equal(t, graphql.WebhookTypeFormationLifecycle, actualWebhook.Type)
	})

	t.Run("Update formation template webhook", func(t *testing.T) {
		webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: &urlUpdated, Type: graphql.WebhookTypeFormationLifecycle})

		require.NoError(t, err)
		updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)

		var updatedWebhook graphql.Webhook
		err = testctx.Tc.RunOperation(ctx, directorCertSecuredClient, updateReq, &updatedWebhook)
		require.NoError(t, err)
		assert.NotNil(t, updatedWebhook.URL)
		assert.Equal(t, urlUpdated, *updatedWebhook.URL)
	})
	t.Run("Delete formation template webhook", func(t *testing.T) {
		deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)

		var deletedWebhook graphql.Webhook
		err = testctx.Tc.RunOperation(ctx, directorCertSecuredClient, deleteReq, &deletedWebhook)
		require.NoError(t, err)

		assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, &scopedFormationTemplate)
	})
}
