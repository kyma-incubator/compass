package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const assignFormationCategory = "assign formation"
const unassignFormationCategory = "unassign formation"

func TestGetFormation(t *testing.T) {
	ctx := context.Background()
	formationName := "formation1"

	t.Logf("Should create formation: %q", formationName)
	formation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, formationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formationName)

	t.Logf("Should get formation %q by id %q", formationName, formation.ID)
	var gotFormation graphql.Formation
	getFormationReq := fixtures.FixGetFormationRequest(formation.ID)
	saveExample(t, getFormationReq.Query(), "query formation")
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationReq, &gotFormation)
	require.NoError(t, err)
	require.Equal(t, formation, gotFormation)

	t.Logf("Should delete formation %q", formationName)
	deleteFormation := fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formationName)
	assert.Equal(t, formation, *deleteFormation)

	t.Logf("Should NOT get formation %q by id %q because it is already deleted", formationName, formation.ID)
	var nonexistentFormation *graphql.Formation
	getNonexistentFormationReq := fixtures.FixGetFormationRequest(formation.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getNonexistentFormationReq, nonexistentFormation)
	require.Error(t, err)
	require.Nil(t, nonexistentFormation)
}

func TestListFormations(t *testing.T) {
	// Pre-cleanup because the formations table may be dirty by previous tests in the director,
	// which delete their created formations after the end of all director tests.
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	tenant.TestTenants.CleanupTenant(tenantID)

	ctx := context.Background()

	firstFormationName := "formation1"
	secondFormationName := "formation2"

	first := 100

	expectedFormations := 0
	t.Logf("List should return %d formations", expectedFormations)
	listFormationsReq := fixtures.FixListFormationsRequestWithPageSize(first)
	saveExample(t, listFormationsReq.Query(), "query formations")
	formationPage1 := fixtures.ListFormations(t, ctx, certSecuredGraphQLClient, listFormationsReq, expectedFormations)
	require.NotNil(t, formationPage1)
	require.Equal(t, expectedFormations, formationPage1.TotalCount)
	require.Empty(t, formationPage1.Data)

	t.Logf("Should create formation: %q", firstFormationName)
	firstFormation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, firstFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, firstFormationName)

	t.Logf("Should create formation: %q", secondFormationName)
	secondFormation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, secondFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, secondFormationName)

	expectedFormations = 2
	t.Logf("List should return %d formations", expectedFormations)
	formationPage2 := fixtures.ListFormations(t, ctx, certSecuredGraphQLClient, listFormationsReq, expectedFormations)
	require.NotNil(t, formationPage2)
	require.Equal(t, expectedFormations, formationPage2.TotalCount)
	require.ElementsMatch(t, formationPage2.Data, []*graphql.Formation{
		&firstFormation,
		&secondFormation,
	})
}

func TestApplicationFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	labelKey := "scenarios"
	defaultValue := conf.DefaultScenario
	newFormation := "ADDITIONAL"
	unusedFormationName := "UNUSED"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application")
	app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	t.Logf("Should create formation: %s", unusedFormationName)
	var unusedFormation graphql.Formation
	createUnusedReq := fixtures.FixCreateFormationRequest(unusedFormationName)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createUnusedReq, &unusedFormation)
	require.NoError(t, err)
	require.Equal(t, unusedFormationName, unusedFormation.Name)

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	saveExample(t, createReq.Query(), "create formation")

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign application to formation %s, as it is not in the label definition", nonExistingFormation)
	failAssignReq := fixtures.FixAssignFormationRequest(app.ID, "APPLICATION", nonExistingFormation)
	var failAssignFormation *graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, failAssignReq, failAssignFormation)
	require.Error(t, err)
	require.Nil(t, failAssignFormation)

	t.Logf("Assign application to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(app.ID, "APPLICATION", newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign application to formation")

	t.Log("Check if new scenario label value was set correctly")
	appRequest := fixtures.FixGetApplicationRequest(app.ID)
	app = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, appRequest, &app)
	require.NoError(t, err)

	scenariosLabel, ok := app.Labels[labelKey].([]interface{})
	require.True(t, ok)

	formations := []string{newFormation}
	if conf.DefaultScenarioEnabled {
		formations = []string{defaultValue, newFormation}
	}

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.Equal(t, formations, actualScenariosEnum)

	t.Log("Should not delete formation while application is assigned")
	deleteRequest := fixtures.FixDeleteFormationRequest(newFormation)
	var nilFormation *graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, nilFormation)
	assert.Error(t, err)
	assert.Nil(t, nilFormation)

	t.Logf("Unassign Application from formation %s", newFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(app.ID, "APPLICATION", newFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign application from formation")

	if conf.DefaultScenarioEnabled {
		unassignDefaultReq := fixtures.FixUnassignFormationRequest(app.ID, "APPLICATION", defaultValue)
		var unassignDefaultFormation graphql.Formation
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignDefaultReq, &unassignDefaultFormation)
		require.NoError(t, err)
		require.Equal(t, defaultValue, unassignDefaultFormation.Name)
	}

	t.Log("Should be able to delete formation after application is unassigned")
	deleteRequest = fixtures.FixDeleteFormationRequest(newFormation)
	var deleteFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, &deleteFormation)
	assert.NoError(t, err)
	assert.Equal(t, newFormation, deleteFormation.Name)

	saveExample(t, deleteRequest.Query(), "delete formation")

	t.Log("Should be able to delete formation")
	deleteUnusedRequest := fixtures.FixDeleteFormationRequest(unusedFormationName)
	var deleteUnusedFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteUnusedRequest, &deleteUnusedFormation)
	assert.NoError(t, err)
	assert.Equal(t, unusedFormationName, deleteUnusedFormation.Name)
}

func TestRuntimeFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	labelKey := "scenarios"
	newFormation := "ADDITIONAL"
	asaFormation := "ASA"
	unusedFormationName := "UNUSED"
	selectorKey := "global_subaccount_id"

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	t.Logf("Should create formation: %s", asaFormation)
	createAsaFormationReq := fixtures.FixCreateFormationRequest(asaFormation)
	var asaGqlFormation graphql.Formation
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAsaFormationReq, &asaGqlFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, asaGqlFormation.Name)

	defer func() {
		t.Log("Should be able to delete ASA formation")
		deleteASAFormationRequest := fixtures.FixDeleteFormationRequest(asaFormation)
		var deleteASAFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteASAFormationRequest, &deleteASAFormation)
		assert.NoError(t, err)
		assert.Equal(t, asaFormation, deleteASAFormation.Name)
	}()

	formationInput := graphql.FormationInput{Name: asaFormation}
	t.Log("Creating ASA")
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)

	rtmName := "rt"
	rtmDesc := "rt-description"
	rtmInput := fixRuntimeInput(rtmName)
	rtmInput.Description = &rtmDesc
	rtmInput.Labels[selectorKey] = subaccountID

	rtm := registerKymaRuntime(t, ctx, subaccountID, rtmInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)

	t.Logf("Should create formation: %s", unusedFormationName)
	var unusedFormation graphql.Formation
	createUnusedReq := fixtures.FixCreateFormationRequest(unusedFormationName)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createUnusedReq, &unusedFormation)
	require.NoError(t, err)
	require.Equal(t, unusedFormationName, unusedFormation.Name)

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign runtime to formation %s, as it is not in the label definition", nonExistingFormation)
	failAssignReq := fixtures.FixAssignFormationRequest(rtm.ID, "RUNTIME", nonExistingFormation)
	var failAssignFormation *graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, failAssignReq, failAssignFormation)
	require.Error(t, err)
	require.Nil(t, failAssignFormation)

	t.Logf("Assign runtime to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(rtm.ID, "RUNTIME", newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign runtime to formation")

	t.Log("Check if new scenario label value was set correctly")
	checkRuntimeFormationLabels(t, ctx, rtm.ID, labelKey, []string{asaFormation, newFormation})

	t.Logf("Assign runtime to formation %s which was already assigned by ASA", asaFormation)
	assignReq = fixtures.FixAssignFormationRequest(rtm.ID, "RUNTIME", asaFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, assignFormation.Name)

	t.Log("Check if the formation label value is still assigned")
	checkRuntimeFormationLabels(t, ctx, rtm.ID, labelKey, []string{asaFormation, newFormation})

	t.Logf("Try to unassign runtime from formation %q which was assigned by ASA", asaFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(rtm.ID, "RUNTIME", asaFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, unassignFormation.Name)

	t.Log("Check that the formation label value is still assigned")
	checkRuntimeFormationLabels(t, ctx, rtm.ID, labelKey, []string{asaFormation, newFormation})

	t.Log("Should not delete formation while runtime is assigned")
	deleteRequest := fixtures.FixDeleteFormationRequest(newFormation)
	var nilFormation *graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, nilFormation)
	assert.Error(t, err)
	assert.Nil(t, nilFormation)

	t.Logf("Unassign Runtime from formation %s", newFormation)
	unassignReq = fixtures.FixUnassignFormationRequest(rtm.ID, "RUNTIME", newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign runtime from formation")

	t.Log("Check that the formation label value is unassigned")
	checkRuntimeFormationLabels(t, ctx, rtm.ID, labelKey, []string{asaFormation})

	t.Log("Should be able to delete formation after runtime is unassigned")
	deleteRequest = fixtures.FixDeleteFormationRequest(newFormation)
	var deleteFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, &deleteFormation)
	assert.NoError(t, err)
	assert.Equal(t, newFormation, deleteFormation.Name)

	t.Log("Should be able to delete formation")
	deleteUnusedRequest := fixtures.FixDeleteFormationRequest(unusedFormationName)
	var deleteUnusedFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteUnusedRequest, &deleteUnusedFormation)
	assert.NoError(t, err)
	assert.Equal(t, unusedFormationName, deleteUnusedFormation.Name)
}

func TestRuntimeContextFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	labelKey := "scenarios"
	newFormation := "ADDITIONAL"
	asaFormation := "ASA"
	asaFormation2 := "ASA2"
	selectorKey := "global_subaccount_id"

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)
	fmt.Println(tenantId, " l ", subaccountID)

	t.Logf("Should create formation: %s", asaFormation)
	createAsaFormationReq := fixtures.FixCreateFormationRequest(asaFormation)
	var asaGqlFormation graphql.Formation
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAsaFormationReq, &asaGqlFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, asaGqlFormation.Name)

	defer func() {
		t.Log("Should be able to delete ASA formation")
		deleteASAFormationRequest := fixtures.FixDeleteFormationRequest(asaFormation)
		var deleteASAFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteASAFormationRequest, &deleteASAFormation)
		assert.NoError(t, err)
		assert.Equal(t, asaFormation, deleteASAFormation.Name)
	}()

	formationInput := graphql.FormationInput{Name: asaFormation}
	t.Log("Creating ASA")
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)

	rtmName := "rt"
	rtmDesc := "rt-description"
	rtmInput := graphql.RuntimeRegisterInput{
		Name:        rtmName,
		Description: &rtmDesc,
		Labels: graphql.Labels{
			selectorKey: subaccountID,
		},
	}

	t.Log("Create runtime")
	rtm := registerKymaRuntime(t, ctx, subaccountID, rtmInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)

	t.Log("Create runtimeContext")
	runtimeContext := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, subaccountID, rtm.ID, "ASATest", "ASATest")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtimeContext.ID)

	t.Log("RuntimeContext should be assigned to formation coming from ASA")
	checkRuntimeContextFormationLabels(t, ctx, rtm.ID, runtimeContext.ID, labelKey, []string{asaFormation})

	t.Logf("Should create formation: %s", asaFormation2)
	createAsaFormationReq2 := fixtures.FixCreateFormationRequest(asaFormation2)
	var asaGqlFormation2 graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAsaFormationReq2, &asaGqlFormation2)
	require.NoError(t, err)
	require.Equal(t, asaFormation2, asaGqlFormation2.Name)

	defer func() {
		t.Log("Should be able to delete ASA formation")
		deleteASAFormationRequest2 := fixtures.FixDeleteFormationRequest(asaFormation2)
		var deleteASAFormation2 graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteASAFormationRequest2, &deleteASAFormation2)
		assert.NoError(t, err)
		assert.Equal(t, asaFormation2, deleteASAFormation2.Name)
	}()

	formationInput2 := graphql.FormationInput{Name: asaFormation2}
	t.Log("Creating second ASA")
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput2, subaccountID, tenantId)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput2, subaccountID, tenantId)

	t.Log("RuntimeContext should be assigned to the new formation coming from ASA as well")
	checkRuntimeContextFormationLabels(t, ctx, rtm.ID, runtimeContext.ID, labelKey, []string{asaFormation, asaFormation2})

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign runtime context to formation %s, as it is not in the label definition", nonExistingFormation)
	failAssignReq := fixtures.FixAssignFormationRequest(rtm.ID, "RUNTIME_CONTEXT", nonExistingFormation)
	var failAssignFormation *graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, failAssignReq, failAssignFormation)
	require.Error(t, err)
	require.Nil(t, failAssignFormation)

	t.Logf("Assign runtime context to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(runtimeContext.ID, "RUNTIME_CONTEXT", newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign runtime context to formation")

	t.Log("Check if new scenario label value was set correctly")
	checkRuntimeContextFormationLabels(t, ctx, rtm.ID, runtimeContext.ID, labelKey, []string{asaFormation, asaFormation2, newFormation})

	t.Logf("Assign runtime context to formation %s which was already assigned by ASA", asaFormation)
	assignReq = fixtures.FixAssignFormationRequest(runtimeContext.ID, "RUNTIME_CONTEXT", asaFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, assignFormation.Name)

	t.Log("Check if the formation label value is still assigned")
	checkRuntimeContextFormationLabels(t, ctx, rtm.ID, runtimeContext.ID, labelKey, []string{asaFormation, asaFormation2, newFormation})

	t.Logf("Try to unassign runtime context from formation %q which was assigned by ASA", asaFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(runtimeContext.ID, "RUNTIME_CONTEXT", asaFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, unassignFormation.Name)

	t.Log("Check that the formation label value is still assigned")
	checkRuntimeContextFormationLabels(t, ctx, rtm.ID, runtimeContext.ID, labelKey, []string{asaFormation, asaFormation2, newFormation})

	t.Log("Should not delete formation while runtime context is assigned")
	deleteRequest := fixtures.FixDeleteFormationRequest(newFormation)
	var nilFormation *graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, nilFormation)
	assert.Error(t, err)
	assert.Nil(t, nilFormation)

	t.Logf("Unassign Runtime Context from formation %s", newFormation)
	unassignReq = fixtures.FixUnassignFormationRequest(runtimeContext.ID, "RUNTIME_CONTEXT", newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign runtime context from formation")

	t.Log("Check that the formation label value is unassigned")
	checkRuntimeContextFormationLabels(t, ctx, rtm.ID, runtimeContext.ID, labelKey, []string{asaFormation, asaFormation2})

	t.Log("Should be able to delete formation after runtime is unassigned")
	deleteRequest = fixtures.FixDeleteFormationRequest(newFormation)
	var deleteFormation graphql.Formation
	deleteRequest.Header.Add("x-request-id", "TURSISE")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, &deleteFormation)
	assert.NoError(t, err)
	assert.Equal(t, newFormation, deleteFormation.Name)
}

func TestTenantFormationFlow(t *testing.T) {
	// GIVEN
	const (
		firstFormation  = "FIRST"
		secondFormation = "SECOND"
	)

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	ctx := context.Background()
	defaultValue := conf.DefaultScenario
	assignment := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: firstFormation,
		Selector: &graphql.LabelSelectorInput{
			Key:   "global_subaccount_id",
			Value: subaccountID,
		},
	}

	expectedFormations := []string{firstFormation, secondFormation}
	if conf.DefaultScenarioEnabled {
		expectedFormations = append(expectedFormations, defaultValue)
	}

	t.Logf("Should create formation: %s", firstFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(firstFormation)
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, firstFormation, formation.Name)

	t.Logf("Should create formation: %s", secondFormation)
	var unusedFormation graphql.Formation
	createUnusedReq := fixtures.FixCreateFormationRequest(secondFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createUnusedReq, &unusedFormation)
	require.NoError(t, err)
	require.Equal(t, secondFormation, unusedFormation.Name)

	t.Logf("Assign tenant %s to formation %s", subaccountID, firstFormation)
	assignReq := fixtures.FixAssignFormationRequest(subaccountID, "TENANT", firstFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, firstFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign tenant to formation")

	t.Log("Should match expected ASA")
	asaPage := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 1, len(asaPage.Data))
	assertions.AssertAutomaticScenarioAssignment(t, assignment, *asaPage.Data[0])

	t.Logf("Unassign tenant %s from formation %s", subaccountID, firstFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(subaccountID, "TENANT", firstFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, firstFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign tenant from formation")

	t.Log("Should match expected ASA")
	asaPage = fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 0, len(asaPage.Data))

	t.Logf("Should be able to delete formation %s", firstFormation)
	deleteRequest := fixtures.FixDeleteFormationRequest(firstFormation)
	var deleteFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, &deleteFormation)
	assert.NoError(t, err)
	assert.Equal(t, firstFormation, deleteFormation.Name)

	t.Logf("Should be able to delete formation %s", secondFormation)
	deleteUnusedRequest := fixtures.FixDeleteFormationRequest(secondFormation)
	var deleteUnusedFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteUnusedRequest, &deleteUnusedFormation)
	assert.NoError(t, err)
	assert.Equal(t, secondFormation, deleteUnusedFormation.Name)
}

func TestRuntimeContextsFormationProcessingFromASA(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Runtime contexts formation processing from ASA", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderAccountID := conf.TestProviderAccountID
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant

		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName

		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		providerRuntimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey: conf.SubscriptionConfig.SelfRegRegion},
		}

		providerRuntime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		require.NoError(t, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(subscription.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(subscription.ContentTypeHeader, subscription.ContentTypeApplicationJson)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		// Register kyma runtime
		kymaRtmInput := fixtures.FixRuntimeRegisterInput("kyma-runtime")
		kymaRuntime := registerKymaRuntime(t, ctx, subscriptionConsumerSubaccountID, kymaRtmInput)

		// Register kyma formation template
		kymaFormationTmplName := "kyma-formation-template-name"
		createFormationTemplate(t, ctx, subscriptionConsumerSubaccountID, "kyma", kymaFormationTmplName, conf.KymaRuntimeTypeLabelValue, graphql.ArtifactTypeEnvironmentInstance)

		// Register provider formation template
		providerFormationTmplName := "provider-formation-template-name"
		createFormationTemplate(t, ctx, subscriptionProviderSubaccountID, "provider", providerFormationTmplName, conf.KymaRuntimeTypeLabelValue, graphql.ArtifactTypeSubscription)

		// Register kyma formation
		kymaFormationName := "kyma-formation-name"
		t.Logf("Creating formation with name: %q", kymaFormationName)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, kymaFormationName, &kymaFormationTmplName) // todo:: tenant?
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, kymaFormationName)

		// Register provider formation
		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q", providerFormationName)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionProviderSubaccountID, providerFormationName, &providerFormationTmplName) // todo:: tenant?
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionProviderSubaccountID, providerFormationName)

		t.Run("Assign kyma runtime to formation and validate scenarios labels", func(t *testing.T) {
			t.Logf("Assert there is no kyma runtime scenarios before assigning tenant with ID: %q to formation", subscriptionConsumerSubaccountID)
			checkRuntimeFormationLabels(t, ctx, kymaRuntime.ID, "scenarios", []string{})
			assignTenantToFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionProviderAccountID, kymaFormationName) // todo:: tenants?
			defer unassignTenantFromFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionProviderAccountID, kymaFormationName)
			t.Logf("Assert scenarios label after assigning tenant with ID: %q to formation", subscriptionConsumerSubaccountID)
			checkRuntimeFormationLabels(t, ctx, kymaRuntime.ID, "scenarios", []string{kymaFormationName})
			t.Log("Assert provider runtime has NOT scenarios label")
			checkRuntimeFormationLabels(t, ctx, providerRuntime.ID, "scenarios", []string{})
		})

		t.Run("Assign provider runtime to formation and validate scenarios labels", func(t *testing.T) {
			t.Logf("Assert there is no provider runtime scenarios before assigning tenant with ID: %q to formation", subscriptionProviderSubaccountID)
			checkRuntimeFormationLabels(t, ctx, providerRuntime.ID, "scenarios", []string{})
			assignTenantToFormation(t, ctx, subscriptionProviderSubaccountID, subscriptionConsumerAccountID, providerFormationName) // todo:: tenants?
			defer unassignTenantFromFormation(t, ctx, subscriptionProviderSubaccountID, subscriptionConsumerAccountID, providerFormationName)
			t.Logf("Assert scenarios label after assigning tenant with ID: %q to formation", subscriptionProviderSubaccountID)
			checkRuntimeFormationLabels(t, ctx, providerRuntime.ID, "scenarios", []string{providerFormationName})
			t.Log("Assert kyma runtime has NOT scenarios label")
			checkRuntimeFormationLabels(t, ctx, kymaRuntime.ID, "scenarios", []string{})
		})
	})
}

func assignTenantToFormation(t *testing.T, ctx context.Context, objectID, tenantID, formationName string) {
	t.Logf("Assign tenant %s to formation %s...", objectID, formationName)
	assignReq := fixtures.FixAssignFormationRequest(objectID, "TENANT", formationName)
	var formation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, assignReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
	t.Logf("Successfully assigned tenant %s to formation %s", objectID, formationName)
}

func unassignTenantFromFormation(t *testing.T, ctx context.Context, objectID, tenantID, formationName string) {
	t.Logf("Unassign tenant %s from formation %s...", objectID, formationName)
	unassignReq := fixtures.FixUnassignFormationRequest(objectID, "TENANT", formationName)
	var formation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, unassignReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
	t.Logf("Successfully unassigned tenant %s from formation %s", objectID, formationName)
}

func createFormationTemplate(t *testing.T, ctx context.Context, tenantID, prefix, formationTemplateName, runtimeType string, runtimeArtifactKind graphql.ArtifactType) {
	formationTmplInput := graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       []string{prefix + "-app-type-1", prefix + "-app-type-2"},
		RuntimeType:            runtimeType,
		RuntimeTypeDisplayName: prefix + "-formation-template-display-name",
		RuntimeArtifactKind:    runtimeArtifactKind,
	}

	formationTmplGQLInput, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTmplInput)
	require.NoError(t, err)
	formationTmplRequest := fixtures.FixCreateFormationTemplateRequest(formationTmplGQLInput)

	ft := graphql.FormationTemplate{}
	t.Logf("Creating formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, formationTmplRequest, &ft)
	require.NoError(t, err)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)
}

func checkRuntimeFormationLabels(t *testing.T, ctx context.Context, rtmID, formationLabelKey string, expectedFormations []string) {
	appRequest := fixtures.FixGetRuntimeRequest(rtmID)
	rtm := graphql.RuntimeExt{}
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, appRequest, &rtm)
	require.NoError(t, err)

	scenariosLabel, ok := rtm.Labels[formationLabelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.ElementsMatch(t, expectedFormations, actualScenariosEnum)
}

func checkRuntimeContextFormationLabels(t *testing.T, ctx context.Context, rtmID, rtmCtxID, formationLabelKey string, expectedFormations []string) {
	rtmRequest := fixtures.FixRuntimeContextRequest(rtmID, rtmCtxID)
	rtm := graphql.RuntimeExt{}
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, rtmRequest, &rtm)
	require.NoError(t, err)

	scenariosLabel, ok := rtm.RuntimeContext.Labels[formationLabelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.ElementsMatch(t, expectedFormations, actualScenariosEnum)
}
