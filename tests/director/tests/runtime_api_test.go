package tests

import (
	"context"
	"fmt"
	"testing"

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
	assertions.AssertRuntime(t, givenInput, actualRuntime, conf.DefaultScenarioEnabled, false)

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
	if conf.DefaultScenarioEnabled {
		assert.Len(t, actualRuntime.Labels, 4)
	} else {
		assert.Len(t, actualRuntime.Labels, 3)
	}

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
	assertions.AssertRuntime(t, in, actualRuntime, conf.DefaultScenarioEnabled, false)
}

func TestModifyRuntimeWebhooks(t *testing.T) {
	ctx := context.Background()
	placeholder := "runtime"
	in := fixRuntimeInput(placeholder)

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	actualRuntime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, in, conf.GatewayOauth)

	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &actualRuntime)

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
	const (
		testFormation = "test-scenario"
	)
	// GIVEN
	ctx := context.Background()
	subaccount := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)
	tenantID := tenant.TestTenants.GetDefaultTenantID()

	givenInput := fixRuntimeInput("runtime-with-scenario-assignments")
	givenInput.Description = ptr.String("runtime-1-description")
	givenInput.Labels["global_subaccount_id"] = []interface{}{subaccount}

	// WHEN
	actualRuntime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccount, givenInput, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount, &actualRuntime)

	//THEN
	assertions.AssertRuntime(t, givenInput, actualRuntime, conf.DefaultScenarioEnabled, true)

	// update label definition
	_ = fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), []string{testFormation, "DEFAULT"})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), []string{"DEFAULT"})

	// assign to formation
	givenFormation := graphql.FormationInput{Name: testFormation}

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
	assert.Contains(t, scenarios, testFormation)

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
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				fixtures.UnregisterRuntime(t, ctx, certSecuredGraphQLClient, tenantId, id)
			}
		}
	}()

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

	name := "test-create-runtime-without-labels"
	runtimeInput := fixRuntimeInput(name)

	runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtimeInput, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertRuntime(t, runtimeInput, fetchedRuntime, conf.DefaultScenarioEnabled, false)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := fixRuntimeUpdateInput(name)
	secondInput.Labels[ScenariosLabel] = []interface{}{"DEFAULT"}
	secondInput.Description = ptr.String("runtime-1-description")
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixtures.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertions.AssertUpdatedRuntime(t, secondInput, secondRuntime, conf.DefaultScenarioEnabled, false)
}

func TestRegisterUpdateRuntimeWithIsNormalizedLabel(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := "test-create-runtime-without-labels"
	runtimeInput := fixRuntimeInput(name)
	runtimeInput.Labels[IsNormalizedLabel] = "false"

	runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtimeInput, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertRuntime(t, runtimeInput, fetchedRuntime, conf.DefaultScenarioEnabled, false)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := fixRuntimeUpdateInput(name)
	secondInput.Description = ptr.String("runtime-1-description")
	secondInput.Labels[ScenariosLabel] = []interface{}{"DEFAULT"}
	secondInput.Labels[IsNormalizedLabel] = "true"

	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeUpdateInputToGQL(secondInput)
	require.NoError(t, err)
	updateReq := fixtures.FixUpdateRuntimeRequest(fetchedRuntime.ID, runtimeInGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &secondRuntime)

	//THEN
	require.NoError(t, err)
	assertions.AssertUpdatedRuntime(t, secondInput, secondRuntime, conf.DefaultScenarioEnabled, false)
}

func TestRuntimeRegisterUpdateAndUnregisterWithCertificate(t *testing.T) {
	t.Run("Test runtime operations(CUD) with externally issued certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		runtimeTypeLabelKey := conf.RuntimeTypeLabelKey
		runtimeTypeLabelValue := conf.SubscriptionProviderAppNameValue

		runtimeInput := fixRuntimeWithSelfRegLabelsInput("runtime-create-update-delete")
		runtimeInput.Description = ptr.String("runtime-create-update-delete-description")

		t.Log("Successfully register runtime with certificate")
		actualRuntime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, certSecuredGraphQLClient, &runtimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &actualRuntime)

		//THEN
		require.NotEmpty(t, actualRuntime.ID)
		assertions.AssertRuntime(t, runtimeInput, actualRuntime, conf.DefaultScenarioEnabled, true)

		t.Log("Successfully set regular runtime label using certificate")
		// GIVEN
		actualLabel := graphql.Label{}

		// WHEN
		addLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, "regularLabel", "regularLabelValue")
		err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, addLabelReq, &actualLabel)

		//THEN
		require.NoError(t, err)
		require.Equal(t, "regularLabel", actualLabel.Key)
		require.Equal(t, "regularLabelValue", actualLabel.Value)

		t.Log("Fail setting immutable label on runtime")
		// GIVEN
		immutableLabel := graphql.Label{}

		// WHEN
		iLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, runtimeTypeLabelKey, runtimeTypeLabelValue)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, iLabelReq, &immutableLabel)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("could not set unmodifiable label with key %s", runtimeTypeLabelKey))
		require.Empty(t, immutableLabel)

		t.Log("Successfully get runtime")
		getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, getRuntimeReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		assert.Len(t, actualRuntime.Labels, 5) // three labels from the different runtime inputs plus two additional during runtime registration - isNormalized and "self register" label

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
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, updateRuntimeReq, &actualRuntime)

		//THEN
		require.NoError(t, err)
		require.Equal(t, runtimeUpdateInput.Name, actualRuntime.Name)
		require.Equal(t, *runtimeUpdateInput.Description, *actualRuntime.Description)
		require.Equal(t, runtimeStatusCond, actualRuntime.Status.Condition)
		require.Equal(t, len(actualRuntime.Labels), 3) // two labels from the runtime input plus one additional label, added during runtime update(isNormalized)

		t.Log("Successfully delete runtime using certificate")
		// WHEN
		delReq := fixtures.FixUnregisterRuntimeRequest(actualRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, delReq, nil)

		//THEN
		require.NoError(t, err)
	})
}

func TestQueryRuntimesWithCertificate(t *testing.T) {
	t.Run("Query runtime with externally issued certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		idsToRemove := make([]string, 0)
		defer func() {
			for _, id := range idsToRemove {
				if id != "" {
					fixtures.UnregisterRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, id)
				}
			}
		}()

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
			err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createReq, &actualRuntime)
			require.NoError(t, err)
			require.NotEmpty(t, actualRuntime.ID)
			rtm.ID = actualRuntime.ID
			idsToRemove = append(idsToRemove, actualRuntime.ID)
		}
		actualPage := graphql.RuntimePage{}

		// WHEN
		queryReq := fixtures.FixGetRuntimesRequestWithPagination()
		err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryReq, &actualPage)

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

		runtimeInput := fixRuntimeInput("runtime-specific-runtime")
		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)
		require.NoError(t, err)
		registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
		createdRuntime := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, registerReq, &createdRuntime)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &createdRuntime)

		require.NoError(t, err)
		require.NotEmpty(t, createdRuntime.ID)

		// WHEN
		queriedRuntime := graphql.Runtime{}
		queryReq := fixtures.FixGetRuntimeRequest(createdRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryReq, &queriedRuntime)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
		assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
		assert.Equal(t, createdRuntime.Description, queriedRuntime.Description)
	})
}

func TestRuntimeTypeLabels(t *testing.T) {
	ctx := context.Background()
	runtimeName := "runtime-with-int-sys-creds"
	runtimeInput := fixRuntimeInput(runtimeName)

	t.Run(fmt.Sprintf("Validate runtime type label - %q is added when runtime is registered with integration system credentials", conf.RuntimeTypeLabelKey), func(t *testing.T) {
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
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenantID, &runtimeInput)
		defer fixtures.CleanupRuntime(t, ctx, oauthGraphQLClient, tenantID, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		t.Logf("Validate the %q label is available...", conf.RuntimeTypeLabelKey)
		runtimeTypeLabelValue, ok := runtime.Labels[conf.RuntimeTypeLabelKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.KymaRuntimeTypeLabelValue, runtimeTypeLabelValue)
	})

	t.Run(fmt.Sprintf("Validate runtime type label - %q is missing when runtime is NOT registered with integration system credentials", conf.RuntimeTypeLabelKey), func(t *testing.T) {
		runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)
		require.NoError(t, err)
		actualRuntime := graphql.RuntimeExt{}

		// WHEN
		registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, registerReq, &actualRuntime)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &actualRuntime)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		assertions.AssertRuntime(t, runtimeInput, actualRuntime, conf.DefaultScenarioEnabled, true)

		t.Logf("Validate %q label is not added when runtime is registered without integration system credentials...", conf.RuntimeTypeLabelKey)
		runtimeTypeLabelValue, ok := actualRuntime.Labels[conf.RuntimeTypeLabelKey].(string)
		require.False(t, ok)
		require.Empty(t, runtimeTypeLabelValue)
	})
}

func TestSelfRegMoreThanOneProviderRuntime(t *testing.T) {
	ctx := context.Background()

	// Self register runtime
	runtimeInput := graphql.RuntimeRegisterInput{
		Name:        "selfRegisterRuntime-1",
		Description: ptr.String("selfRegisterRuntime-1-description"),
		Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey: conf.SubscriptionConfig.SelfRegRegion},
	}

	t.Logf("Self registering runtime with labels %q:%q and %q:%q...", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion)
	runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, certSecuredGraphQLClient, &runtimeInput)
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &runtime)
	require.NotEmpty(t, runtime.ID)
	strLbl, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
	require.True(t, ok)
	require.Contains(t, strLbl, runtime.ID)

	// Self register second runtime with same distinguish label and region labels
	secondRuntimeInput := graphql.RuntimeRegisterInput{
		Name:        "selfRegisterRuntime-2",
		Description: ptr.String("selfRegisterRuntime-2-description"),
		Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey: conf.SubscriptionConfig.SelfRegRegion},
	}

	t.Logf("Self registering second runtime with same distinguish label: %q and region: %q and validate it will fail...", conf.SubscriptionConfig.SelfRegDistinguishLabelValue, conf.SubscriptionConfig.SelfRegRegion)
	inputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(secondRuntimeInput)
	require.NoError(t, err)

	registerSecondRuntimeRequest := fixtures.FixRegisterRuntimeRequest(inputGQL)
	var secondRuntimeExt graphql.RuntimeExt

	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, registerSecondRuntimeRequest, &secondRuntimeExt)
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &secondRuntimeExt)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("cannot have more than one runtime with labels %q: %q and %q: %q", tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue))
	require.Empty(t, secondRuntimeExt)
}

func fixRuntimeInput(name string) graphql.RuntimeRegisterInput {
	input := fixtures.FixRuntimeRegisterInput(name)
	delete(input.Labels, "placeholder")

	return input
}

func fixRuntimeWithSelfRegLabelsInput(name string) graphql.RuntimeRegisterInput {
	input := fixtures.FixRuntimeRegisterInput(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = []interface{}{conf.SubscriptionConfig.SelfRegDistinguishLabelValue}
	input.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion
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
