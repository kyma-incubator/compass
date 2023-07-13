package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ScenariosLabel          = "scenarios"
	IsNormalizedLabel       = "isNormalized"
	QueryRuntimesCategory   = "query runtimes"
	RegisterRuntimeCategory = "register runtime"
	GlobalSubaccountIdKey   = "global_subaccount_id"
)

func TestRuntimeRegisterUpdateAndUnregister(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	givenInput := fixRuntimeInput("runtime-create-update-delete")
	givenInput.Description = ptr.String("runtime-1-description")
	givenInput.Labels["ggg"] = []interface{}{"hhh"}

	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	saveExampleInCustomDir(t, registerReq.Query(), RegisterRuntimeCategory, "register runtime")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerReq, &actualRuntime)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertions.AssertRuntime(t, givenInput, actualRuntime)

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, "new_label", []string{"bbb"})
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new_label", actualLabel.Key)
	assert.Len(t, actualLabel.Value, 1)
	assert.Contains(t, actualLabel.Value, "bbb")

	// get runtime and validate runtimes
	getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	assert.Len(t, actualRuntime.Labels, 3)

	// add agent auth
	// GIVEN
	in := fixtures.FixSampleApplicationRegisterInputWithWebhooks("app")

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createAppReq := fixtures.FixRegisterApplicationRequest(appInputGQL)

	//WHEN
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAppReq, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	// update runtime, check if only simple values are updated
	//GIVEN
	givenUpdateInput := fixRuntimeUpdateInput("updated-name")
	givenUpdateInput.Description = ptr.String("updated-description")
	givenUpdateInput.Labels["key"] = []interface{}{"values", "aabbcc"}

	runtimeStatusCond := graphql.RuntimeStatusConditionConnected
	givenUpdateInput.StatusCondition = &runtimeStatusCond

	runtimeUpdateInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(givenUpdateInput)
	require.NoError(t, err)
	updateRuntimeReq := fixtures.FixUpdateRuntimeRequest(actualRuntime.ID, runtimeUpdateInGQL)
	saveExample(t, updateRuntimeReq.Query(), "update runtime")
	//WHEN
	actualRuntime = graphql.RuntimeExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateRuntimeReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, givenUpdateInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenUpdateInput.Description, *actualRuntime.Description)
	assert.Equal(t, len(actualRuntime.Labels), 2)
	assert.Equal(t, runtimeStatusCond, actualRuntime.Status.Condition)

	// delete runtime

	// WHEN
	delReq := fixtures.FixUnregisterRuntimeRequest(actualRuntime.ID)
	saveExample(t, delReq.Query(), "unregister runtime")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, delReq, nil)

	//THEN
	require.NoError(t, err)
}

func TestRuntimeRegisterWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	url := "http://mywordpress.com/webhooks1"

	in := fixRuntimeInput("runtime-with-webhooks")
	in.Description = ptr.String("runtime-1-description")
	in.Webhooks = []*graphql.WebhookInput{
		{
			Type: graphql.WebhookTypeConfigurationChanged,
			Auth: fixtures.FixBasicAuth(t),
			URL:  &url,
		},
	}

	runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(in)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	request := fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)
	saveExampleInCustomDir(t, request.Query(), RegisterRuntimeCategory, "register Runtime with webhooks")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualRuntime)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertions.AssertRuntime(t, in, actualRuntime)
}

func TestModifyRuntimeWebhooks(t *testing.T) {
	ctx := context.Background()
	placeholder := "runtime"
	in := fixRuntimeInput(placeholder)
	tenantId := tenant.TestTenants.GetDefaultTenantID()
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(in)
	require.NoError(t, err)
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	actualRuntime := graphql.RuntimeExt{}
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &actualRuntime)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerReq, &actualRuntime)
	assert.NoError(t, err)

	// add
	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"
	url := "http://new-webhook.url"
	urlUpdated := "http://updated-webhook.url"
	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeConfigurationChanged,
		OutputTemplate: &outputTemplate,
	})

	require.NoError(t, err)
	addReq := fixtures.FixAddWebhookToRuntimeRequest(actualRuntime.ID, webhookInStr)
	saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add runtime webhook")

	actualWebhook := graphql.Webhook{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addReq, &actualWebhook)
	require.NoError(t, err)

	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
	assert.Equal(t, graphql.WebhookTypeConfigurationChanged, actualWebhook.Type)
	id := actualWebhook.ID
	require.NotNil(t, id)

	// get all webhooks
	updatedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, actualRuntime.ID)
	assert.Len(t, updatedRuntime.Webhooks, 1)

	// update
	webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL: &urlUpdated, Type: graphql.WebhookTypeConfigurationChanged, OutputTemplate: &outputTemplate})

	require.NoError(t, err)
	updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
	saveExampleInCustomDir(t, updateReq.Query(), updateWebhookCategory, "update webhook")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)

	// delete

	//GIVEN
	deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
	saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete webhook")

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteReq, &actualWebhook)

	//THEN
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)
}

func TestRuntimeUnregisterDeletesScenarioAssignments(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	subaccount := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)
	tenantID := tenant.TestTenants.GetDefaultTenantID()

	givenInput := fixRuntimeInput("runtime-with-scenario-assignments")
	givenInput.Description = ptr.String("runtime-1-description")
	givenInput.Labels[GlobalSubaccountIdKey] = []interface{}{subaccount}

	// WHEN
	var actualRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount, &actualRuntime)
	actualRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccount, givenInput, conf.GatewayOauth)

	//THEN
	assertions.AssertRuntime(t, givenInput, actualRuntime)

	// update label definition
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)

	// assign to formation
	givenFormation := graphql.FormationInput{Name: testScenario}

	actualFormation := graphql.Formation{}

	// WHEN
	assignFormationReq := fixtures.FixAssignFormationRequest(subaccount, string(graphql.FormationObjectTypeTenant), givenFormation.Name)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, assignFormationReq, &actualFormation)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, givenFormation.Name, actualFormation.Name)

	// get runtime - verify it is in scenario
	getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getRuntimeReq, &actualRuntime)

	require.NoError(t, err)
	scenarios, hasScenarios := actualRuntime.Labels["scenarios"]
	assert.True(t, hasScenarios)
	assert.Len(t, scenarios, 1)
	assert.Contains(t, scenarios, testScenario)

	// delete runtime

	// WHEN
	delReq := fixtures.FixUnregisterRuntimeRequest(actualRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, delReq, nil)

	//THEN
	require.NoError(t, err)

	// get automatic scenario assignment - see that it's deleted
	actualScenarioAssignments := graphql.AutomaticScenarioAssignmentPage{}
	getScenarioAssignmentsReq := fixtures.FixAutomaticScenarioAssignmentsRequest()
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getScenarioAssignmentsReq, &actualScenarioAssignments)
	require.NoError(t, err)
	assert.Equal(t, actualScenarioAssignments.TotalCount, 0)
}

func TestQueryRuntimes(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	idsToRemove := make([]string, 0)

	inputRuntimes := []*graphql.Runtime{
		{Name: "runtime-query-1", Description: ptr.String("test description")},
		{Name: "runtime-query-2", Description: ptr.String("another description")},
		{Name: "runtime-query-3"},
	}

	for _, rtm := range inputRuntimes {
		givenInput := fixRuntimeInput(rtm.Name)
		givenInput.Description = rtm.Description

		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(givenInput)
		require.NoError(t, err)
		createReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
		actualRuntime := graphql.Runtime{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		rtm.ID = actualRuntime.ID
		idsToRemove = append(idsToRemove, actualRuntime.ID)
	}
	for _, id := range idsToRemove {
		if id != "" {
			defer fixtures.UnregisterRuntime(t, ctx, certSecuredGraphQLClient, tenantId, id)
		}
	}
	actualPage := graphql.RuntimePage{}

	// WHEN
	queryReq := fixtures.FixGetRuntimesRequestWithPagination()
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryReq, &actualPage)
	saveExampleInCustomDir(t, queryReq.Query(), QueryRuntimesCategory, "query runtimes")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualPage.Data, len(inputRuntimes))
	assert.Equal(t, len(inputRuntimes), actualPage.TotalCount)

	for _, inputRtm := range inputRuntimes {
		found := false
		for _, actualRtm := range actualPage.Data {
			if inputRtm.ID == actualRtm.ID {
				found = true
				assert.Equal(t, inputRtm.Name, actualRtm.Name)
				assert.Equal(t, inputRtm.Description, actualRtm.Description)
				break
			}
		}
		assert.True(t, found)
	}
}

func TestQuerySpecificRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	givenInput := fixRuntimeInput("runtime-specific-runtime")
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(givenInput)
	require.NoError(t, err)
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	createdRuntime := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerReq, &createdRuntime)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &createdRuntime)

	require.NoError(t, err)
	require.NotEmpty(t, createdRuntime.ID)

	// WHEN
	queriedRuntime := graphql.Runtime{}
	queryReq := fixtures.FixGetRuntimeRequest(createdRuntime.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryReq, &queriedRuntime)
	saveExample(t, queryReq.Query(), "query runtime")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
	assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
	assert.Equal(t, createdRuntime.Description, queriedRuntime.Description)
}

func TestQueryRuntimesWithPagination(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimes := make(map[string]*graphql.Runtime)
	runtimesAmount := 10
	for i := 0; i < runtimesAmount; i++ {
		runtimeInput := fixRuntimeInput(fmt.Sprintf("runtime-%d", i))
		runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)
		require.NoError(t, err)

		registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.Runtime{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerReq, &runtime)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &graphql.RuntimeExt{Runtime: runtime})

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		runtimes[runtime.ID] = &runtime
	}

	after := 3
	cursor := ""
	queriesForFullPage := int(runtimesAmount / after)

	for i := 0; i < queriesForFullPage; i++ {
		runtimesRequest := fixtures.FixRuntimeRequestWithPaginationRequest(after, cursor)

		//WHEN
		runtimePage := graphql.RuntimePage{}
		err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, runtimesRequest, &runtimePage)
		require.NoError(t, err)

		//THEN
		assert.Equal(t, cursor, string(runtimePage.PageInfo.StartCursor))
		assert.True(t, runtimePage.PageInfo.HasNextPage)
		assert.Len(t, runtimePage.Data, after)
		assert.Equal(t, runtimesAmount, runtimePage.TotalCount)
		for _, runtime := range runtimePage.Data {
			assert.Equal(t, runtime, runtimes[runtime.ID])
			delete(runtimes, runtime.ID)
		}
		cursor = string(runtimePage.PageInfo.EndCursor)
	}

	//WHEN get last page with last runtime
	runtimesRequest := fixtures.FixRuntimeRequestWithPaginationRequest(after, cursor)
	lastRuntimePage := graphql.RuntimePage{}
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, runtimesRequest, &lastRuntimePage)
	require.NoError(t, err)
	saveExampleInCustomDir(t, runtimesRequest.Query(), QueryRuntimesCategory, "query runtimes with pagination")

	//THEN
	assert.False(t, lastRuntimePage.PageInfo.HasNextPage)
	assert.Empty(t, lastRuntimePage.PageInfo.EndCursor)
	require.Len(t, lastRuntimePage.Data, 1)
	assert.Equal(t, lastRuntimePage.Data[0], runtimes[lastRuntimePage.Data[0].ID])
	delete(runtimes, lastRuntimePage.Data[0].ID)
	assert.Len(t, runtimes, 0)
}

func TestRegisterUpdateRuntimeWithoutLabels(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)

	name := "test-create-runtime-without-labels"
	runtimeInput := fixRuntimeInput(name)

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtimeInput, conf.GatewayOauth)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertKymaRuntime(t, runtimeInput, fetchedRuntime)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := fixRuntimeUpdateInput(name)
	secondInput.Labels[ScenariosLabel] = []interface{}{testScenario}
	secondInput.Description = ptr.String("runtime-1-description")
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixtures.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertions.AssertUpdatedRuntime(t, secondInput, secondRuntime)
}

func TestRegisterUpdateRuntimeWithIsNormalizedLabel(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)

	name := "test-create-runtime-without-labels"
	runtimeInput := fixRuntimeInput(name)
	runtimeInput.Labels[IsNormalizedLabel] = "false"

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtimeInput, conf.GatewayOauth)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertKymaRuntime(t, runtimeInput, fetchedRuntime)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := fixRuntimeUpdateInput(name)
	secondInput.Description = ptr.String("runtime-1-description")
	secondInput.Labels[ScenariosLabel] = []interface{}{testScenario}
	secondInput.Labels[IsNormalizedLabel] = "true"

	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixtures.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertions.AssertUpdatedRuntime(t, secondInput, secondRuntime)
}

func TestRuntimeRegisterUpdateAndUnregisterWithCertificate(t *testing.T) {
	t.Run("Test runtime operations(CUD) with externally issued certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		runtimeTypeLabelKey := conf.RuntimeTypeLabelKey
		runtimeTypeLabelValue := conf.SubscriptionProviderAppNameValue

		runtimeInput := fixRuntimeWithSelfRegLabelsInput("runtime-create-update-delete")
		runtimeInput.Description = ptr.String("runtime-create-update-delete-description")

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		var actualRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &actualRuntime)
		actualRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
		t.Log("Successfully register runtime with certificate")

		//THEN
		require.NotEmpty(t, actualRuntime.ID)
		runtimeInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

		saasAppLbl, ok := actualRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		assertions.AssertRuntime(t, runtimeInput, actualRuntime)

		t.Log("Successfully set regular runtime label using certificate")
		// GIVEN
		actualLabel := graphql.Label{}

		// WHEN
		addLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, "regularLabel", "regularLabelValue")
		err := testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, addLabelReq, &actualLabel)

		//THEN
		require.NoError(t, err)
		require.Equal(t, "regularLabel", actualLabel.Key)
		require.Equal(t, "regularLabelValue", actualLabel.Value)

		t.Log("Fail setting immutable label on runtime")
		// GIVEN
		immutableLabel := graphql.Label{}

		// WHEN
		iLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, runtimeTypeLabelKey, runtimeTypeLabelValue)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, iLabelReq, &immutableLabel)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("could not set unmodifiable label with key %s", runtimeTypeLabelKey))
		require.Empty(t, immutableLabel)

		t.Log("Successfully get runtime")
		getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getRuntimeReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		assert.Len(t, actualRuntime.Labels, 6) // two labels from the different runtime inputs plus four additional during runtime registration - isNormalized and three "self register" labels - region, "self register" label and SaaS app name label

		t.Log("Successfully update runtime with certificate")
		//GIVEN
		runtimeUpdateInput := fixRuntimeUpdateWithSelfRegLabelsInput("updated-runtime")
		runtimeUpdateInput.Description = ptr.String("updated-runtime-description")

		runtimeStatusCond := graphql.RuntimeStatusConditionConnected
		runtimeUpdateInput.StatusCondition = &runtimeStatusCond

		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(runtimeUpdateInput)
		require.NoError(t, err)
		updateRuntimeReq := fixtures.FixUpdateRuntimeRequest(actualRuntime.ID, runtimeInGQL)

		//WHEN
		actualRuntime = graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateRuntimeReq, &actualRuntime)

		//THEN
		require.NoError(t, err)
		require.Equal(t, runtimeUpdateInput.Name, actualRuntime.Name)
		require.Equal(t, *runtimeUpdateInput.Description, *actualRuntime.Description)
		require.Equal(t, runtimeStatusCond, actualRuntime.Status.Condition)
		require.Len(t, actualRuntime.Labels, 5) // two labels from the runtime input, one additional label, added during runtime update(isNormalized) plus the two "self register" labels

		t.Log("Successfully delete runtime using certificate")
		// WHEN
		delReq := fixtures.FixUnregisterRuntimeRequest(actualRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, delReq, nil)

		//THEN
		require.NoError(t, err)
	})
}

func TestQueryRuntimesWithCertificate(t *testing.T) {
	t.Run("Query runtime with externally issued certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		idsToRemove := make([]string, 0)

		inputRuntimes := []*graphql.Runtime{
			{Name: "runtime-query-1", Description: ptr.String("test description")},
			{Name: "runtime-query-2", Description: ptr.String("another description")},
			{Name: "runtime-query-3"},
		}

		for _, rtm := range inputRuntimes {
			givenInput := fixRuntimeInput(rtm.Name)
			givenInput.Description = rtm.Description
			runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(givenInput)
			require.NoError(t, err)
			createReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
			actualRuntime := graphql.Runtime{}
			err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, createReq, &actualRuntime)
			require.NoError(t, err)
			require.NotEmpty(t, actualRuntime.ID)
			rtm.ID = actualRuntime.ID
			idsToRemove = append(idsToRemove, actualRuntime.ID)
		}
		for _, id := range idsToRemove {
			if id != "" {
				defer fixtures.UnregisterRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, id)
			}
		}
		actualPage := graphql.RuntimePage{}

		// WHEN
		queryReq := fixtures.FixGetRuntimesRequestWithPagination()
		err := testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, queryReq, &actualPage)

		//THEN
		require.NoError(t, err)
		assert.Len(t, actualPage.Data, len(inputRuntimes))
		assert.Equal(t, len(inputRuntimes), actualPage.TotalCount)

		for _, inputRtm := range inputRuntimes {
			found := false
			for _, actualRtm := range actualPage.Data {
				if inputRtm.ID == actualRtm.ID {
					found = true
					assert.Equal(t, inputRtm.Name, actualRtm.Name)
					assert.Equal(t, inputRtm.Description, actualRtm.Description)
					break
				}
			}
			assert.True(t, found)
		}
	})
}

func TestQuerySpecificRuntimeWithCertificate(t *testing.T) {
	t.Run("Query specific runtime with externally issued certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		runtimeInput := fixRuntimeInput("runtime-specific-runtime")
		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)
		require.NoError(t, err)
		registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
		createdRuntime := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, registerReq, &createdRuntime)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &createdRuntime)

		require.NoError(t, err)
		require.NotEmpty(t, createdRuntime.ID)

		// WHEN
		queriedRuntime := graphql.Runtime{}
		queryReq := fixtures.FixGetRuntimeRequest(createdRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, queryReq, &queriedRuntime)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
		assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
		assert.Equal(t, createdRuntime.Description, queriedRuntime.Description)
	})
}

func TestRuntimeTypeAndRegionLabels(t *testing.T) {
	ctx := context.Background()
	runtimeName := "runtime-with-int-sys-creds"
	runtimeNameCert := "runtime-with-cert-creds"

	t.Run(fmt.Sprintf("Validate %q, %q labels and application namespace - they are added when runtime is registered with integration system credentials", conf.RuntimeTypeLabelKey, tenantfetcher.RegionKey), func(t *testing.T) {
		runtimeInput := fixRuntimeInput(runtimeName)
		subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount) // randomly selected subaccount the parent of which is the default tenant used below
		tenantID := tenant.TestTenants.GetDefaultTenantID()
		intSysName := "runtime-integration-system"

		t.Logf("Creating integration system with name: %q", intSysName)
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSysName)
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		t.Logf("Registering runtime with name %q with integration system credentials...", runtimeName)
		runtimeInput.Labels[GlobalSubaccountIdKey] = []interface{}{subaccountID} // so that the region can be set for the runtime based on the region of the subaccount
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenantID, &runtimeInput)
		defer fixtures.CleanupRuntime(t, ctx, oauthGraphQLClient, tenantID, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		t.Logf("Validate the %q label is available...", conf.RuntimeTypeLabelKey)
		runtimeTypeLabelValue, ok := runtime.Labels[conf.RuntimeTypeLabelKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.KymaRuntimeTypeLabelValue, runtimeTypeLabelValue)

		t.Log("Validate that the Application Namespace is available...")
		appNamespace := runtime.ApplicationNamespace
		require.NotNil(t, appNamespace)
		require.Equal(t, conf.KymaApplicationNamespaceValue, *appNamespace)

		t.Log("Validate that the region label of the runtime is available...")
		rtRegionLabelValue, ok := runtime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.DefaultTenantRegion, rtRegionLabelValue)

	})

	t.Run(fmt.Sprintf("Validate %q, %q labels and application namespace - they are missing when runtime is NOT registered with integration system credentials", conf.RuntimeTypeLabelKey, tenantfetcher.RegionKey), func(t *testing.T) {
		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		rtInput := fixRuntimeInput(runtimeNameCert)
		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(rtInput)
		require.NoError(t, err)
		actualRuntime := graphql.RuntimeExt{}

		// WHEN
		registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, registerReq, &actualRuntime)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &actualRuntime)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		assertions.AssertRuntime(t, rtInput, actualRuntime)

		t.Logf("Validate %q label is not added when runtime is registered without integration system credentials...", conf.RuntimeTypeLabelKey)
		runtimeTypeLabelValue, ok := actualRuntime.Labels[conf.RuntimeTypeLabelKey].(string)
		require.False(t, ok)
		require.Empty(t, runtimeTypeLabelValue)

		t.Log("Validate that the Application Namespace is nil...")
		appNamespace := actualRuntime.ApplicationNamespace
		require.Nil(t, appNamespace)

		t.Log("Validate that the region label is not added when runtime is registered without integration system credentials...")
		rtRegionLabelValue, ok := actualRuntime.Labels[tenantfetcher.RegionKey].(string)
		require.False(t, ok)
		require.Empty(t, rtRegionLabelValue)
	})
}

func TestSelfRegMoreThanOneProviderRuntime(t *testing.T) {
	ctx := context.Background()

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	// Self register runtime
	runtimeInput := graphql.RuntimeRegisterInput{
		Name:        "selfRegisterRuntime-1",
		Description: ptr.String("selfRegisterRuntime-1-description"),
		Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
	}

	t.Logf("Self registering runtime with labels %q:%q and %q:%q...", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion)
	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &runtime)
	runtime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
	require.NotEmpty(t, runtime.ID)
	strLbl, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
	require.True(t, ok)
	require.Contains(t, strLbl, runtime.ID)

	regionLbl, ok := runtime.Labels[tenantfetcher.RegionKey].(string)
	require.True(t, ok)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, regionLbl)

	// Self register second runtime with same distinguish label and region labels
	secondRuntimeInput := graphql.RuntimeRegisterInput{
		Name:        "selfRegisterRuntime-2",
		Description: ptr.String("selfRegisterRuntime-2-description"),
		Labels: graphql.Labels{
			conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
		},
	}

	t.Logf("Self registering second runtime with same distinguish label: %q and region: %q and validate it will fail...", conf.SubscriptionConfig.SelfRegDistinguishLabelValue, conf.SubscriptionConfig.SelfRegRegion)
	inputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(secondRuntimeInput)
	require.NoError(t, err)

	registerSecondRuntimeRequest := fixtures.FixRegisterRuntimeRequest(inputGQL)
	var secondRuntimeExt graphql.RuntimeExt

	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, registerSecondRuntimeRequest, &secondRuntimeExt)
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &secondRuntimeExt)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("cannot have more than one runtime with labels %q: %q and %q: %q", tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue))
	require.Empty(t, secondRuntimeExt)
}

func TestRuntimeTypeImmutability(t *testing.T) {
	ctx := context.Background()

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	testRuntimeType := "testRuntimeTypeValue"
	testSaaSAppName := "testSaaSAppNameValue"

	// Runtime with runtimeType label
	runtimeInput := graphql.RuntimeRegisterInput{
		Name:        "immutable-runtime",
		Description: ptr.String("runtime-description"),
		Labels:      graphql.Labels{conf.RuntimeTypeLabelKey: testRuntimeType, conf.SaaSAppNameLabelKey: testSaaSAppName, tenantfetcher.RegionKey: conf.SubscriptionConfig.SelfRegRegion},
	}

	t.Logf("Registering runtime with the following labels: %q:%q, %q:%q and %q:%q", conf.RuntimeTypeLabelKey, testRuntimeType, conf.SaaSAppNameLabelKey, testSaaSAppName, tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion)
	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &runtime)
	runtime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
	require.NotEmpty(t, runtime.ID)

	require.Equal(t, len(runtime.Labels), 2)
	strLbl, ok := runtime.Labels[tenantfetcher.RegionKey].(string)
	require.True(t, ok)
	require.Equal(t, strLbl, conf.SubscriptionConfig.SelfRegRegion)

	strLbl, ok = runtime.Labels[IsNormalizedLabel].(string)
	require.True(t, ok)
	require.Equal(t, strLbl, "true")

	// check immutable labels are not added
	require.NotContains(t, runtime.Labels, conf.RuntimeTypeLabelKey)
	require.NotContains(t, runtime.Labels, conf.SaaSAppNameLabelKey)

	// Update runtime with immutable labels
	updateRuntimeInput := graphql.RuntimeUpdateInput{
		Name:        "immutable-runtime-labels-updated",
		Description: ptr.String("immutable-runtime-labels-description"),
		Labels:      graphql.Labels{conf.RuntimeTypeLabelKey: testRuntimeType, conf.SaaSAppNameLabelKey: testSaaSAppName, tenantfetcher.RegionKey: conf.SubscriptionConfig.SelfRegRegion},
	}
	runtimeUpdateInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(updateRuntimeInput)
	require.NoError(t, err)

	updateRuntimeReq := fixtures.FixUpdateRuntimeRequest(runtime.ID, runtimeUpdateInGQL)
	t.Logf("Updating runtime with the following labels: %q:%q, %q:%q and %q:%q", conf.RuntimeTypeLabelKey, testRuntimeType, conf.SaaSAppNameLabelKey, testSaaSAppName, tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion)
	updatedRuntime := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateRuntimeReq, &updatedRuntime)
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &updatedRuntime)

	require.NoError(t, err)
	require.Equal(t, "immutable-runtime-labels-updated", updatedRuntime.Name)
	require.Equal(t, "immutable-runtime-labels-description", *updatedRuntime.Description)

	require.Equal(t, len(updatedRuntime.Labels), 2)
	strLbl, ok = updatedRuntime.Labels[tenantfetcher.RegionKey].(string)
	require.True(t, ok)
	require.Equal(t, strLbl, conf.SubscriptionConfig.SelfRegRegion)

	strLbl, ok = runtime.Labels[IsNormalizedLabel].(string)
	require.True(t, ok)
	require.Equal(t, strLbl, "true")

	// check immutable labels are not added during update operation
	require.NotContains(t, updatedRuntime.Labels, conf.RuntimeTypeLabelKey)
	require.NotContains(t, updatedRuntime.Labels, conf.SaaSAppNameLabelKey)
}

func fixRuntimeInput(name string) graphql.RuntimeRegisterInput {
	input := fixtures.FixRuntimeRegisterInput(name)
	delete(input.Labels, "placeholder")

	return input
}

func fixRuntimeWithSelfRegLabelsInput(name string) graphql.RuntimeRegisterInput {
	input := fixtures.FixRuntimeRegisterInput(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = []interface{}{conf.SubscriptionConfig.SelfRegDistinguishLabelValue}
	delete(input.Labels, "placeholder")

	return input
}

func fixRuntimeUpdateInput(name string) graphql.RuntimeUpdateInput {
	input := fixtures.FixRuntimeUpdateInput(name)
	delete(input.Labels, "placeholder")

	return input
}

func fixRuntimeUpdateWithSelfRegLabelsInput(name string) graphql.RuntimeUpdateInput {
	input := fixtures.FixRuntimeUpdateInput(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = []interface{}{conf.SubscriptionConfig.SelfRegDistinguishLabelValue}
	input.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion
	delete(input.Labels, "placeholder")

	return input
}
