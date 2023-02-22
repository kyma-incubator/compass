package tests

import (
	"context"
	"fmt"
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
	globalFormationTemplateName     = "Side-by-side extensibility with Kyma"
)

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

	appType := "async-app-type-1"
	formationTemplateName := "create-formation-template-name"
	leadingProductID := "leading-product-id"
	leadingProductIDs := []*string{&leadingProductID}

	formationTemplateInput := fixtures.FixFormationTemplateInputWithLeadingProductIDs(formationTemplateName, "runtimeTypeTest", []string{appType}, graphql.ArtifactTypeEnvironmentInstance, leadingProductIDs)

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

	saveExampleInCustomDir(t, createFormationTemplateRequest.Query(), CreateFormationTemplateCategory, "create formation template")

	t.Logf("Check if formation template with name %q was created", formationTemplateName)

	formationTemplateOutput := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)

	assertions.AssertFormationTemplate(t, &formationTemplateInput, formationTemplateOutput)
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
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)
	require.NoError(t, err)

	//THEN
	require.NotEmpty(t, output.ID)
	require.NotEmpty(t, output.Name)
	assertions.AssertFormationTemplate(t, &formationTemplateInput, &output)

	saveExampleInCustomDir(t, createFormationTemplateRequest.Query(), CreateFormationTemplateCategory, "create formation template with webhooks")

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
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, output.ID)
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
	saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add formation template webhook")

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

	t.Run("Delete formation template webhook", func(t *testing.T) {
		//GIVEN
		deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
		saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete webhook")

		//WHEN
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, *webhookInput.URL, *actualWebhook.URL)
	})
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

	assert.Greater(t, len(formationTemplatePage.Data), 1) // assert that both tenant scoped and global formation templates are visible
	assert.Subset(t, formationTemplatePage.Data, []*graphql.FormationTemplate{
		scopedFormationTemplate,
	})

	t.Logf("List all formation templates for some other tenant in which formation template with name: %q was NOT created and verify that it is NOT visible there", scopedFormationTemplateName)
	formationTemplatePageForOtherTenant := fixtures.QueryFormationTemplatesWithPageSizeAndTenant(t, ctx, directorCertSecuredClient, first, tenant.TestTenants.GetDefaultTenantID())

	assert.NotEmpty(t, formationTemplatePageForOtherTenant.Data)
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

func TestTenantScopedFormationTemplatesWithWebhooks(t *testing.T) {
	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	scopedFormationTemplateName := "tenant-scoped-formation-template-with-webhook-test"
	scopedFormationTemplateInput := fixtures.FixFormationTemplateInput(scopedFormationTemplateName)
	webhookSyncMode := graphql.WebhookModeSync

	t.Logf("Create tenant scoped formation template with name: %q", scopedFormationTemplateName)
	scopedFormationTemplate := fixtures.CreateFormationTemplate(t, ctx, directorCertSecuredClient, scopedFormationTemplateInput) // tenant_id is extracted from the subject of the cert
	defer fixtures.CleanupFormationTemplate(t, ctx, directorCertSecuredClient, scopedFormationTemplate.ID)

	assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, scopedFormationTemplate)
	urlUpdated := "http://updated.url"
	webhookInput := &graphql.WebhookInput{
		Type: graphql.WebhookTypeFormationLifecycle,
		Mode: &webhookSyncMode,
		URL:  str.Ptr("http://localhost:6439/"),
	}
	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(webhookInput)
	addReq := fixtures.FixAddWebhookToFormationTemplateRequest(scopedFormationTemplate.ID, webhookInStr)
	saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add formation template webhook")
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

		assertions.AssertFormationTemplate(t, &scopedFormationTemplateInput, scopedFormationTemplate)
	})
}
