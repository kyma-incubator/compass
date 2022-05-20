package tests

import (
	"context"
	"fmt"
	"testing"

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
	RegionLabel             = "region"
	IsNormalizedLabel       = "isNormalized"
	QueryRuntimesCategory   = "query runtimes"
	RegisterRuntimeCategory = "register runtime"
)

func TestRuntimeRegisterUpdateAndUnregister(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	givenInput := graphql.RuntimeRegisterInput{
		Name:        "runtime-create-update-delete",
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{"ggg": []interface{}{"hhh"}},
	}
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
	givenUpdateInput := graphql.RuntimeUpdateInput{
		Name:        "updated-name",
		Description: ptr.String("updated-description"),
		Labels: graphql.Labels{
			"key": []interface{}{"values", "aabbcc"},
		},
	}
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

	in := graphql.RuntimeRegisterInput{
		Name:        "runtime-with-webhooks",
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{"ggg": []interface{}{"hhh"}},
		Webhooks: []*graphql.WebhookInput{
			{
				Type: graphql.WebhookTypeConfigurationChanged,
				Auth: fixtures.FixBasicAuth(t),
				URL:  &url,
			},
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
	in := fixtures.FixRuntimeRegisterInput(placeholder)

	runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(in)
	require.NoError(t, err)

	createReq := fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)
	actualRuntime := graphql.RuntimeExt{}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &actualRuntime)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &actualRuntime)

	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)

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

	givenInput := graphql.RuntimeRegisterInput{
		Name:        "runtime-with-scenario-assignments",
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{"global_subaccount_id": []interface{}{subaccount}},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	// WHEN
	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerReq, &actualRuntime)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assertions.AssertRuntime(t, givenInput, actualRuntime, conf.DefaultScenarioEnabled, true)

	// update label definition
	_ = fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), []string{testFormation, "DEFAULT"})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), []string{"DEFAULT"})

	// assign to formation
	givenFormation := graphql.FormationInput{Name: testFormation}

	actualFormation := graphql.Formation{}

	// WHEN
	assignFormationReq := fixtures.FixAssignFormationRequest(subaccount, string(graphql.FormationObjectTypeTenant), givenFormation.Name)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, assignFormationReq, &actualFormation)

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
		givenInput := graphql.RuntimeRegisterInput{
			Name:        rtm.Name,
			Description: rtm.Description,
		}
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

	givenInput := graphql.RuntimeRegisterInput{
		Name: "runtime-specific-runtime",
	}
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
		runtimeInput := graphql.RuntimeRegisterInput{
			Name: fmt.Sprintf("runtime-%d", i),
		}
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
	runtimeInput := graphql.RuntimeRegisterInput{Name: name}

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &runtimeInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertRuntime(t, runtimeInput, fetchedRuntime, conf.DefaultScenarioEnabled, false)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeUpdateInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{ScenariosLabel: []interface{}{"DEFAULT"}},
	}
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
	runtimeInput := graphql.RuntimeRegisterInput{
		Name:   name,
		Labels: graphql.Labels{IsNormalizedLabel: "false"},
	}

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &runtimeInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	//WHEN
	fetchedRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	//THEN
	require.Equal(t, runtime.ID, fetchedRuntime.ID)
	assertions.AssertRuntime(t, runtimeInput, fetchedRuntime, conf.DefaultScenarioEnabled, false)

	//GIVEN
	secondRuntime := graphql.RuntimeExt{}
	secondInput := graphql.RuntimeUpdateInput{
		Name:        name,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{IsNormalizedLabel: "true", ScenariosLabel: []interface{}{"DEFAULT"}},
	}
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
		distinguishLabelValue := conf.SelfRegDistinguishLabelValue

		protectedConsumerSubaccountIdsLabel := "consumer_subaccount_ids"

		runtimeInput := &graphql.RuntimeRegisterInput{
			Name:        "register-runtime-with-protected-labels",
			Description: ptr.String("register-runtime-with-protected-labels-description"),
			Labels:      graphql.Labels{protectedConsumerSubaccountIdsLabel: []string{"subaccountID-1", "subaccountID-2"}},
		}

		t.Log("Successfully register runtime using certificate with protected labels and validate that they are excluded")
		actualRtm := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, certSecuredGraphQLClient, runtimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &actualRtm)

		//THEN
		require.NotEmpty(t, actualRtm.ID)
		require.Equal(t, runtimeInput.Name, actualRtm.Name)
		require.Equal(t, runtimeInput.Description, actualRtm.Description)
		require.Empty(t, actualRtm.Labels[protectedConsumerSubaccountIdsLabel])

		t.Log("Successfully register runtime with certificate")
		// GIVEN
		runtimeInput = &graphql.RuntimeRegisterInput{
			Name:        "runtime-create-update-delete",
			Description: ptr.String("runtime-create-update-delete-description"),
			Labels:      graphql.Labels{conf.SelfRegDistinguishLabelKey: []interface{}{distinguishLabelValue}, RegionLabel: conf.SelfRegRegion},
		}

		actualRuntime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, certSecuredGraphQLClient, runtimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &actualRuntime)

		//THEN
		require.NotEmpty(t, actualRuntime.ID)
		assertions.AssertRuntime(t, *runtimeInput, actualRuntime, conf.DefaultScenarioEnabled, true)

		t.Log("Successfully set regular runtime label using certificate")
		// GIVEN
		actualLabel := graphql.Label{}

		// WHEN
		addLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, "regular_label", []string{"labelValue"})
		err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, addLabelReq, &actualLabel)

		//THEN
		require.NoError(t, err)
		require.Equal(t, "regular_label", actualLabel.Key)
		require.Len(t, actualLabel.Value, 1)
		require.Contains(t, actualLabel.Value, "labelValue")

		t.Log("Fail setting protected label on runtime")
		// GIVEN
		protectedLabel := graphql.Label{}

		// WHEN
		pLabelReq := fixtures.FixSetRuntimeLabelRequest(actualRuntime.ID, protectedConsumerSubaccountIdsLabel, []string{"subaccountID-1", "subaccountID-2"})
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, pLabelReq, &protectedLabel)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "could not set unmodifiable label with key consumer_subaccount_ids")
		require.Empty(t, protectedLabel)

		t.Log("Successfully get runtime")
		getRuntimeReq := fixtures.FixGetRuntimeRequest(actualRuntime.ID)
		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, getRuntimeReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		assert.Len(t, actualRuntime.Labels, 5) // three labels from the different runtime inputs plus two additional during runtime registration - isNormalized and "self register" label

		t.Log("Successfully update runtime and validate the protected labels are excluded")
		//GIVEN
		runtimeUpdateInput := graphql.RuntimeUpdateInput{
			Name:        "updated-runtime",
			Description: ptr.String("updated-runtime-description"),
			Labels: graphql.Labels{
				conf.SelfRegDistinguishLabelKey: []interface{}{distinguishLabelValue}, RegionLabel: conf.SelfRegRegion, protectedConsumerSubaccountIdsLabel: []interface{}{"subaccountID-1", "subaccountID-2"},
			},
		}

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
		labelValues, ok := actualRuntime.Labels[protectedConsumerSubaccountIdsLabel]
		require.False(t, ok)
		require.Empty(t, labelValues)

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
			givenInput := graphql.RuntimeRegisterInput{
				Name:        rtm.Name,
				Description: rtm.Description,
			}
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

		runtimeInput := graphql.RuntimeRegisterInput{
			Name: "runtime-specific-runtime",
		}
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
