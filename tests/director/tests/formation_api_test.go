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

	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	assignFormationCategory   = "assign formation"
	unassignFormationCategory = "unassign formation"
	assignOperation           = "assign"
	unassignOperation         = "unassign"
	createFormationOperation  = "createFormation"
	deleteFormationOperation  = "deleteFormation"
	emptyParentCustomerID     = "" // in the respective tests, the used GA tenant does not have customer parent, thus we assert that it is empty
)

func TestGetFormation(t *testing.T) {
	ctx := context.Background()

	t.Logf("Should create formation: %q", testScenario)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(testScenario)
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, testScenario, formation.Name)

	saveExample(t, createReq.Query(), "create formation")

	t.Logf("Should get formation %q by id %q", testScenario, formation.ID)
	var gotFormation graphql.Formation
	getFormationReq := fixtures.FixGetFormationRequest(formation.ID)
	saveExample(t, getFormationReq.Query(), "query formation")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationReq, &gotFormation)
	require.NoError(t, err)
	require.Equal(t, formation, gotFormation)

	t.Logf("Should get formation %q by name %q", testScenario, formation.Name)
	getFormationByNameReq := fixtures.FixGetFormationByNameRequest(formation.Name)
	saveExample(t, getFormationByNameReq.Query(), "query formation by name")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationByNameReq, &gotFormation)
	require.NoError(t, err)
	require.Equal(t, formation, gotFormation)

	t.Logf("Should delete formation %q", testScenario)
	deleteFormation := fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	assert.Equal(t, formation, *deleteFormation)

	t.Logf("Should NOT get formation %q by id %q because it is already deleted", testScenario, formation.ID)
	var nonexistentFormation *graphql.Formation
	getNonexistentFormationReq := fixtures.FixGetFormationRequest(formation.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getNonexistentFormationReq, nonexistentFormation)
	require.Error(t, err)
	require.Nil(t, nonexistentFormation)
}

func TestListFormations(t *testing.T) {
	ctx := context.Background()

	firstFormationName := "formation1"
	secondFormationName := "formation2"

	first := 100

	expectedFormations := 0
	t.Logf("List should return %d formations", expectedFormations)
	listFormationsReq := fixtures.FixListFormationsRequestWithPageSize(first)
	saveExample(t, listFormationsReq.Query(), "query formations")
	formationPage1 := fixtures.ListFormations(t, ctx, certSecuredGraphQLClient, listFormationsReq, expectedFormations)
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
	require.ElementsMatch(t, formationPage2.Data, []*graphql.Formation{
		&firstFormation,
		&secondFormation,
	})
}

func TestApplicationFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	newFormation := "ADDITIONAL"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application")
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "app", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign application to formation %s, as it is not in the label definition", nonExistingFormation)
	failAssignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), nonExistingFormation)
	var failAssignFormation *graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, failAssignReq, failAssignFormation)
	require.Error(t, err)
	require.Nil(t, failAssignFormation)

	t.Logf("Assign application to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), newFormation)
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

	scenariosLabel, ok := app.Labels[ScenariosLabel].([]interface{})
	require.True(t, ok)

	formations := []string{newFormation}

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
	assert.Contains(t, err.Error(), "are not valid against empty schema")
	assert.Nil(t, nilFormation)

	t.Logf("Unassign Application from formation %s", newFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), newFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign application from formation")

	t.Log("Should be able to delete formation after application is unassigned")
	deleteRequest = fixtures.FixDeleteFormationRequest(newFormation)
	var deleteFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, &deleteFormation)
	assert.NoError(t, err)
	assert.Equal(t, newFormation, deleteFormation.Name)

	saveExample(t, deleteRequest.Query(), "delete formation")
}

func TestRuntimeFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
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
	defer func() {
		t.Log("Should be able to delete ASA formation")
		deleteASAFormationRequest := fixtures.FixDeleteFormationRequest(asaFormation)
		var deleteASAFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteASAFormationRequest, &deleteASAFormation)
		assert.NoError(t, err)
		assert.Equal(t, asaFormation, deleteASAFormation.Name)
	}()
	require.NoError(t, err)
	require.Equal(t, asaFormation, asaGqlFormation.Name)

	formationInput := graphql.FormationInput{Name: asaFormation}
	t.Log("Creating ASA")
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)

	rtmName := "rt"
	rtmDesc := "rt-description"
	rtmInput := fixRuntimeInput(rtmName)
	rtmInput.Description = &rtmDesc
	rtmInput.Labels[selectorKey] = subaccountID

	var rtm graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)
	rtm = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, rtmInput, conf.GatewayOauth)

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
	failAssignReq := fixtures.FixAssignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), nonExistingFormation)
	var failAssignFormation *graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, failAssignReq, failAssignFormation)
	require.Error(t, err)
	require.Nil(t, failAssignFormation)

	t.Logf("Assign runtime to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign runtime to formation")

	t.Log("Check if new scenario label value was set correctly")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation, newFormation})

	t.Logf("Assign runtime to formation %s which was already assigned by ASA should succeed", asaFormation)
	assignReq = fixtures.FixAssignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), asaFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)

	t.Log("Check if the formation label value is still assigned")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation, newFormation})

	t.Logf("Try to unassign runtime from formation %q which was assigned by ASA", asaFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), asaFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, unassignFormation.Name)

	t.Log("Check that the formation label value is still assigned")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation, newFormation})

	t.Log("Should not delete formation while runtime is assigned")
	deleteRequest := fixtures.FixDeleteFormationRequest(newFormation)
	var nilFormation *graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, nilFormation)
	assert.Error(t, err)
	assert.Nil(t, nilFormation)

	t.Logf("Unassign Runtime from formation %s", newFormation)
	unassignReq = fixtures.FixUnassignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign runtime from formation")

	t.Log("Check that the formation label value is unassigned")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation})

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
	newFormation := "ADDITIONAL"
	asaFormation := "ASA"
	asaFormation2 := "ASA2"
	selectorKey := "global_subaccount_id"

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	t.Logf("Should create formation: %s", asaFormation)
	createAsaFormationReq := fixtures.FixCreateFormationRequest(asaFormation)
	var asaGqlFormation graphql.Formation
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAsaFormationReq, &asaGqlFormation)
	defer func() {
		t.Log("Should be able to delete ASA formation")
		deleteASAFormationRequest := fixtures.FixDeleteFormationRequest(asaFormation)
		var deleteASAFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteASAFormationRequest, &deleteASAFormation)
		assert.NoError(t, err)
		assert.Equal(t, asaFormation, deleteASAFormation.Name)
	}()
	require.NoError(t, err)
	require.Equal(t, asaFormation, asaGqlFormation.Name)

	formationInput := graphql.FormationInput{Name: asaFormation}
	t.Log("Creating ASA")
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)

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
	var rtm graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)
	rtm = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, rtmInput, conf.GatewayOauth)

	t.Log("Create runtimeContext")
	runtimeContext := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, subaccountID, rtm.ID, "ASATest", "ASATest")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtimeContext.ID)

	t.Log("RuntimeContext should be assigned to formation coming from ASA")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation})

	t.Logf("Should create formation: %s", asaFormation2)
	createAsaFormationReq2 := fixtures.FixCreateFormationRequest(asaFormation2)
	var asaGqlFormation2 graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAsaFormationReq2, &asaGqlFormation2)
	defer func() {
		t.Log("Should be able to delete ASA formation")
		deleteASAFormationRequest2 := fixtures.FixDeleteFormationRequest(asaFormation2)
		var deleteASAFormation2 graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteASAFormationRequest2, &deleteASAFormation2)
		assert.NoError(t, err)
		assert.Equal(t, asaFormation2, deleteASAFormation2.Name)
	}()
	require.NoError(t, err)
	require.Equal(t, asaFormation2, asaGqlFormation2.Name)

	formationInput2 := graphql.FormationInput{Name: asaFormation2}
	t.Log("Creating second ASA")
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput2.Name, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput2, subaccountID, tenantId)

	t.Log("RuntimeContext should be assigned to the new formation coming from ASA as well")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation, asaFormation2})

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign runtime context to formation %s, as it is not in the label definition", nonExistingFormation)
	failAssignReq := fixtures.FixAssignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntimeContext), nonExistingFormation)
	var failAssignFormation *graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, failAssignReq, failAssignFormation)
	require.Error(t, err)
	require.Nil(t, failAssignFormation)

	t.Logf("Assign runtime context to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(runtimeContext.ID, string(graphql.FormationObjectTypeRuntimeContext), newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign runtime context to formation")

	t.Log("Check if new scenario label value was set correctly")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation, asaFormation2, newFormation})

	t.Logf("Assign runtime context to formation %s which was already assigned by ASA should fail with conflict", asaFormation)
	assignReq = fixtures.FixAssignFormationRequest(runtimeContext.ID, string(graphql.FormationObjectTypeRuntimeContext), asaFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)

	t.Log("Check if the formation label value is still assigned")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation, asaFormation2, newFormation})

	t.Logf("Try to unassign runtime context from formation %q which was assigned by ASA", asaFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(runtimeContext.ID, string(graphql.FormationObjectTypeRuntimeContext), asaFormation)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, asaFormation, unassignFormation.Name)

	t.Log("Check that the formation label value is still assigned")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation, asaFormation2, newFormation})

	t.Log("Should not delete formation while runtime context is assigned")
	deleteRequest := fixtures.FixDeleteFormationRequest(newFormation)
	var nilFormation *graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, nilFormation)
	assert.Error(t, err)
	assert.Nil(t, nilFormation)

	t.Logf("Unassign Runtime Context from formation %s", newFormation)
	unassignReq = fixtures.FixUnassignFormationRequest(runtimeContext.ID, string(graphql.FormationObjectTypeRuntimeContext), newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	saveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign runtime context from formation")

	t.Log("Check that the formation label value is unassigned")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation, asaFormation2})

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
	scenarioSelector := fixtures.FixLabelSelector("global_subaccount_id", subaccountID)

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
	assignReq := fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), firstFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, firstFormation, assignFormation.Name)

	saveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign tenant to formation")

	t.Log("Should match expected ASA")
	asaPage := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 1, len(asaPage.Data))
	assertions.AssertAutomaticScenarioAssignment(t, firstFormation, &scenarioSelector, *asaPage.Data[0])

	t.Logf("Unassign tenant %s from formation %s", subaccountID, firstFormation)
	unassignReq := fixtures.FixUnassignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), firstFormation)
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

func TestSubaccountInAtMostOneFormationOfType(t *testing.T) {
	ctx := context.Background()
	const (
		firstFormation  = "FIRST"
		secondFormation = "SECOND"
	)

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	formationTemplateName := "create-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	formationTemplateInputGQLString, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTemplateInput)
	require.NoError(t, err)

	createFormationTemplateRequest := fixtures.FixCreateFormationTemplateRequest(formationTemplateInputGQLString)
	formationTemplate := graphql.FormationTemplate{}

	t.Logf("Create formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createFormationTemplateRequest, &formationTemplate)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, formationTemplate.ID)
	require.NotEmpty(t, formationTemplate.Name)

	in := graphql.FormationConstraintInput{
		Name:            "TestSubaccountInAtMostOneFormationOfType",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        "IsNotAssignedToAnyFormationOfType",
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
	require.NotEmpty(t, constraint.ID)

	t.Logf("Attaching constraint to formation template")
	createRequest := fixtures.FixAttachConstraintToFormationTemplateRequest(constraint.ID, formationTemplate.ID)
	saveExample(t, createRequest.Query(), "attach constraint to formation template")

	constraintReference := graphql.ConstraintReference{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, createRequest, &constraintReference))

	t.Logf("Should create formation: %s", firstFormation)

	formationInput := fixtures.FixFormationInput(firstFormation, str.Ptr(formationTemplateName))
	formationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(formationInput)
	require.NoError(t, err)

	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationWithTemplateRequest(formationInputGQL)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, firstFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, firstFormation, formation.Name)

	t.Logf("Should create formation: %s", secondFormation)

	secondFormationInput := fixtures.FixFormationInput(secondFormation, str.Ptr(formationTemplateName))
	secondFormationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(secondFormationInput)
	require.NoError(t, err)

	var unusedFormation graphql.Formation
	createUnusedReq := fixtures.FixCreateFormationWithTemplateRequest(secondFormationInputGQL)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, secondFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createUnusedReq, &unusedFormation)
	require.NoError(t, err)
	require.Equal(t, secondFormation, unusedFormation.Name)

	t.Logf("Assign tenant %s to formation %s", subaccountID, firstFormation)
	assignReq := fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), firstFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, firstFormation, subaccountID, tenantId)
	require.NoError(t, err)
	require.Equal(t, firstFormation, assignFormation.Name)

	t.Log("Should match expected ASA")
	asaPage := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 1, len(asaPage.Data))

	scenarioSelector := fixtures.FixLabelSelector("global_subaccount_id", subaccountID)
	assertions.AssertAutomaticScenarioAssignment(t, firstFormation, &scenarioSelector, *asaPage.Data[0])

	t.Logf("Should fail to assign tenant %s to second formation of type %s", subaccountID, formationTemplateName)
	assignReq = fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), secondFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, secondFormation, subaccountID, tenantId)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Operator \"IsNotAssignedToAnyFormationOfType\" is not satisfied")

	t.Logf("Detaching constraint from formation template")
	deleteRequest := fixtures.FixDetachConstraintFromFormationTemplateRequest(constraint.ID, formationTemplate.ID)
	saveExample(t, deleteRequest.Query(), "detach constraint from formation template")

	constraintReference = graphql.ConstraintReference{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, deleteRequest, &constraintReference)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	t.Logf("Should succeed assigning tenant %s to second formation of type %s after constraint is detached", subaccountID, formationTemplateName)
	assignReq = fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), secondFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, secondFormation, subaccountID, tenantId)
	require.NoError(t, err)
	require.Equal(t, secondFormation, assignFormation.Name)

	t.Log("Should match expected ASAs")
	asaPage = fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 2, len(asaPage.Data))

	assertions.AssertAutomaticScenarioAssignments(t, map[string]*graphql.Label{firstFormation: &scenarioSelector, secondFormation: &scenarioSelector}, asaPage.Data)

	t.Logf("Unassign tenant %s from formation %s", subaccountID, firstFormation)
	var unassignFormation graphql.Formation

	unassignReq := fixtures.FixUnassignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), firstFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, firstFormation, unassignFormation.Name)

	t.Logf("Unassign tenant %s from formation %s", subaccountID, secondFormation)
	unassignReq = fixtures.FixUnassignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), secondFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, secondFormation, unassignFormation.Name)

	t.Log("Should match expected ASA")
	asaPage = fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 0, len(asaPage.Data))
}

func TestRuntimeContextsFormationProcessingFromASA(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Runtime context formation processing from ASA", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant

		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName

		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// Create kyma formation template
		kymaFormationTmplName := "kyma-formation-template-name"
		kymaAppTypes := []string{"kyma-app-type-1", "kyma-app-type-2"}
		var kymaFT graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &kymaFT)
		kymaFT = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, kymaFormationTmplName, conf.KymaRuntimeTypeLabelValue, kymaAppTypes, graphql.ArtifactTypeEnvironmentInstance)

		// Create provider formation template
		providerFormationTmplName := "provider-formation-template-name"
		providerAppTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		runtimeTypes := []string{conf.SubscriptionProviderAppNameValue, "runtime-type-2"}
		var providerFT graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &providerFT)
		providerFT = createFormationTemplateWithMultipleRuntimeTypes(t, ctx, providerFormationTmplName, runtimeTypes, providerAppTypes, graphql.ArtifactTypeSubscription)

		// Create kyma formation
		kymaFormationName := "kyma-formation-name"
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, kymaFormationName)
		fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, kymaFormationName, &kymaFormationTmplName)

		// Create provider formation
		providerFormationName := "provider-formation-name"
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)

		t.Run("Create Automatic Scenario Assignment BEFORE runtime creation", func(t *testing.T) {
			// Create Automatic Scenario Assignment for kyma formation
			defer unassignTenantFromFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, kymaFormationName)
			assignTenantToFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, kymaFormationName)

			// Create Automatic Scenario Assignment for provider formation
			defer unassignTenantFromFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, providerFormationName)
			assignTenantToFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, providerFormationName)

			// Register kyma runtime
			kymaRtmInput := fixtures.FixRuntimeRegisterInput("kyma-runtime")
			var kymaRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
			defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &kymaRuntime)
			kymaRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, kymaRtmInput, conf.GatewayOauth)

			// Register provider runtime
			providerRuntimeInput := graphql.RuntimeRegisterInput{
				Name:        "providerRuntime",
				Description: ptr.String("providerRuntime-description"),
				Labels: graphql.Labels{
					conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
				},
			}

			var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
			defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
			providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
			require.NotEmpty(t, providerRuntime.ID)

			selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
			require.True(t, ok)
			require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

			saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
			require.True(t, ok)
			require.NotEmpty(t, saasAppLbl)

			httpClient := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
				},
			}

			depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
			require.NoError(t, err)
			response, err := httpClient.Do(depConfigureReq)
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Logf("Could not close response body %s", err)
				}
			}()
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)

			apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
			subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
			require.NoError(t, err)
			subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
			subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
			subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
			subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

			// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
			// In case there isn't subscription it will fail-safe without error
			subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

			t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
			resp, err := httpClient.Do(subscribeReq)
			defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Could not close response body %s", err)
				}
			}()
			require.NoError(t, err)
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

			subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
			require.NotEmpty(t, subJobStatusPath)
			subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
			require.Eventually(t, func() bool {
				return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
			}, subscription.EventuallyTimeout, subscription.EventuallyTick)
			t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

			// Validate kyma and provider runtimes scenarios labels
			validateRuntimesScenariosLabels(t, ctx, subscriptionConsumerAccountID, kymaFormationName, providerFormationName, kymaRuntime.ID, providerRuntime.ID)
		})

		t.Run("Create Automatic Scenario Assignment AFTER runtime creation", func(t *testing.T) {
			ctx = context.Background()

			kymaRtmInput := fixtures.FixRuntimeRegisterInput("kyma-runtime")
			var kymaRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
			defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &kymaRuntime)
			kymaRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, kymaRtmInput, conf.GatewayOauth)

			// Register provider runtime
			providerRuntimeInput := graphql.RuntimeRegisterInput{
				Name:        "providerRuntime",
				Description: ptr.String("providerRuntime-description"),
				Labels: graphql.Labels{
					conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
				},
			}

			var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
			defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
			providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
			require.NotEmpty(t, providerRuntime.ID)

			selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
			require.True(t, ok)
			require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

			saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
			require.True(t, ok)
			require.NotEmpty(t, saasAppLbl)

			httpClient := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
				},
			}

			depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
			require.NoError(t, err)
			response, err := httpClient.Do(depConfigureReq)
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Logf("Could not close response body %s", err)
				}
			}()
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)

			apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
			subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
			require.NoError(t, err)
			subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
			subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
			subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
			subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

			// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
			// In case there isn't subscription it will fail-safe without error
			subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

			t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
			resp, err := httpClient.Do(subscribeReq)
			defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Could not close response body %s", err)
				}
			}()
			require.NoError(t, err)
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

			subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
			require.NotEmpty(t, subJobStatusPath)
			subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
			require.Eventually(t, func() bool {
				return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
			}, subscription.EventuallyTimeout, subscription.EventuallyTick)
			t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

			// Create Automatic Scenario Assignment for kyma formation
			defer unassignTenantFromFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, kymaFormationName)
			assignTenantToFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, kymaFormationName)

			// Create Automatic Scenario Assignment for provider formation
			defer unassignTenantFromFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, providerFormationName)
			assignTenantToFormation(t, ctx, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID, providerFormationName)

			// Validate kyma and provider runtimes scenarios labels
			validateRuntimesScenariosLabels(t, ctx, subscriptionConsumerAccountID, kymaFormationName, providerFormationName, kymaRuntime.ID, providerRuntime.ID)
		})
	})
}

func TestFormationAssignmentNotificationsTenantHierarchy(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Assignment Notifications tenant hierarchy", func(t *testing.T) {
		ctx := context.Background()

		subscriptionProviderSubaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)
		subscriptionConsumerTenantID := conf.TestConsumerTenantID // randomly chosen id

		// This test will be executed only on 'local' env and there is a requirement for the subscriptionConsumer tenants - to be related from SA up to CRM. The tenants below are randomly selected
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountIDTenantHierarchy                      // randomly selected child of subscriptionConsumerAccount (tenant.TestTenants.GetDefaultTenantID())
		subscriptionConsumerAccountID := conf.TestConsumerAccountIDTenantHierarchy                            // this global account tenant is selected because it has both subaccount child and customer parent
		subscriptionConsumerCustomerID := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // this is the customer parent of `tenant.TestTenants.GetDefaultTenantID()`

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		mode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		providerRuntimeInput := fixtures.FixProviderRuntimeWithWebhookInput("formation-assignment-tenant-hierarchy-e2e-providerRuntime", graphql.WebhookTypeConfigurationChanged, mode, urlTemplate, inputTemplate, outputTemplate, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaimsTenantHierarchy")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		t.Log("Assert provider runtime is visible in the consumer's subaccount after successful subscription")
		consumerSubaccountRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, providerRuntime.ID)
		require.Equal(t, providerRuntime.ID, consumerSubaccountRuntime.ID)

		t.Log("Assert there is a runtime context(subscription) as part of the provider runtime")
		require.Len(t, consumerSubaccountRuntime.RuntimeContexts.Data, 1)
		require.NotEmpty(t, consumerSubaccountRuntime.RuntimeContexts.Data[0].ID)
		require.Equal(t, conf.SubscriptionLabelKey, consumerSubaccountRuntime.RuntimeContexts.Data[0].Key)
		require.Equal(t, subscriptionConsumerTenantID, consumerSubaccountRuntime.RuntimeContexts.Data[0].Value)
		runtimeContextID := consumerSubaccountRuntime.RuntimeContexts.Data[0].ID

		applicationType := "provider-app-type-1"
		providerFormationTmplName := "provider-formation-template-name"

		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, []string{applicationType}, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template-test")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, appRegion, appNamespace)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType,
			Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "app1-formation-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App 1",
				},
			},
		}

		t.Logf("Create application 1 from template %q", applicationType)
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		app1 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &app1)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app1)
		require.NoError(t, err)
		require.NotEmpty(t, app1.ID)
		t.Logf("app1 ID: %q", app1.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)

		t.Logf("Assign application 1 to formation %s", providerFormationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

		notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, subscriptionConsumerCustomerID)
	})
}

func TestFormationAssignmentNotifications(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Assignment Notifications", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		mode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		providerRuntimeInput := fixtures.FixProviderRuntimeWithWebhookInput("formation-assignment-notifications-e2e-providerRuntime", graphql.WebhookTypeConfigurationChanged, mode, urlTemplate, inputTemplate, outputTemplate, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		rtmRequest := fixtures.FixGetRuntimeContextsRequest(providerRuntime.ID)
		rtm := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, rtmRequest, &rtm)
		require.NoError(t, err)
		require.Len(t, rtm.RuntimeContexts.Data, 1)
		runtimeContextID := rtm.RuntimeContexts.Data[0].ID

		applicationType := "provider-app-type-1"
		providerFormationTmplName := "provider-formation-template-name"

		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, []string{applicationType}, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, appRegion, appNamespace)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType,
			Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "app1-formation-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App 1",
				},
			},
		}

		t.Logf("Create application 1 from template %q", applicationType)
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		app1 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &app1)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app1)
		require.NoError(t, err)
		require.NotEmpty(t, app1.ID)
		t.Logf("app1 ID: %q", app1.ID)

		appFromTmplSrc2 := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "app2-formation-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App 2",
				},
			},
		}

		t.Logf("Create application 2 from template %q", applicationType)
		appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
		require.NoError(t, err)
		createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
		app2 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplSecondRequest, &app2)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app2)
		require.NoError(t, err)
		require.NotEmpty(t, app2.ID)
		t.Logf("app2 ID: %q", app2.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)

		t.Logf("Assign application 1 to formation %s", providerFormationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

		notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		t.Logf("Assign application 2 to formation %s", providerFormationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, app2.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
				app2.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
			},
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
				app2.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
			},
			app2.ID: {
				app2.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
				app1.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 9, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

		notificationForApp2Found := false
		for _, notification := range notificationsForConsumerTenant.Array() {
			appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
			t.Logf("Found notification for app %q", appIDFromNotification)
			if appIDFromNotification == app2.ID {
				notificationForApp2Found = true
				assertFormationNotificationFromAssignments(t, notification, assignOperation, formation.ID, app2.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
			}
		}
		require.True(t, notificationForApp2Found, "notification for assign app2 not found")

		t.Logf("Unassign Application 1 from formation %s", providerFormationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
				app2.ID:          fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
		unassignNotificationFound := false
		for _, notification := range notificationsForConsumerTenant.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationFound = true
				assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
			}
		}
		require.True(t, unassignNotificationFound, "notification for unassign app1 not found")

		t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
		unassignNotificationForApp2Found := false
		for _, notification := range notificationsForConsumerTenant.Array() {
			op := notification.Get("Operation").String()
			appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
			t.Logf("Found %q notification for app %q", op, appIDFromNotification)
			if appIDFromNotification == app2.ID && op == unassignOperation {
				unassignNotificationForApp2Found = true
				assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
			}
		}
		require.True(t, unassignNotificationForApp2Found, "notification for unassign app2 not found")

		t.Logf("Unassign Application 2 from formation %s", providerFormationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
	})
}

func TestAppToAppFormationAssignmentNotifications(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()
	tntParentCustomer := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // parent of `tenant.TestTenants.GetDefaultTenantID()` above

	mode := graphql.WebhookModeSync

	urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
	inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
	outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"
	formationTmplName := "app-to-app-formation-template-name"

	certSecuredHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{
					{
						Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
						PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
					},
				},
				ClientAuth:         tls.RequireAndVerifyClientCert,
				InsecureSkipVerify: conf.SkipSSLValidation,
			},
		},
	}
	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

	// TODO:: Remove runtimeType once we support app only formations
	appTypes := []string{"app-to-app-app-type-1", "app-to-app-app-type-2"}
	var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, "dummy-runtime-type", appTypes, graphql.ArtifactTypeSubscription)

	formationName := "app-to-app-formation-name"
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, "app-template")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	applicationType := "app-to-app-app-type-1"
	appRegion := "test-app-region"
	appNamespace := "compass.test"
	localTenantID := "local-tenant-id"
	t.Logf("Create application template for type %q", applicationType)
	appTemplateInput := fixtures.FixApplicationTemplateWithWebhookNotifications(applicationType, localTenantID, appRegion, appNamespace, graphql.WebhookTypeApplicationTenantMapping, mode, urlTemplate, inputTemplate, outputTemplate)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
	require.NoError(t, err)

	applicationType2 := "app-to-app-app-type-2"
	localTenantID2 := "local-tenant-id2"
	t.Logf("Create application template for type %q", applicationType2)
	appTemplateInput = graphql.ApplicationTemplateInput{
		Name:        applicationType2,
		Description: &applicationType2,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:          "{{name}}",
			ProviderName:  str.Ptr("compass"),
			Description:   ptr.String("test {{display-name}}"),
			LocalTenantID: &localTenantID2,
			Labels: graphql.Labels{
				"applicationType": applicationType2,
				"region":          appRegion,
			},
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name: "name",
			},
			{
				Name: "display-name",
			},
		},
		ApplicationNamespace: &appNamespace,
		AccessLevel:          graphql.ApplicationTemplateAccessLevelGlobal,
	}
	appTmpl, err = fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
	require.NoError(t, err)

	appFromTmplSrc := graphql.ApplicationFromTemplateInput{
		TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app1-formation-notifications-tests",
			},
			{
				Placeholder: "display-name",
				Value:       "App 1",
			},
		},
	}

	t.Logf("Create application 1 from template %q", applicationType)
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	app1 := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplFirstRequest, &app1)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app1)
	require.NoError(t, err)
	require.NotEmpty(t, app1.ID)
	t.Logf("app1 ID: %q", app1.ID)

	appFromTmplSrc2 := graphql.ApplicationFromTemplateInput{
		TemplateName: applicationType2, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app2-formation-notifications-tests",
			},
			{
				Placeholder: "display-name",
				Value:       "App 2",
			},
		},
	}

	t.Logf("Create application 2 from template %q", applicationType2)
	appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
	require.NoError(t, err)
	createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
	app2 := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplSecondRequest, &app2)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app2)
	require.NoError(t, err)
	require.NotEmpty(t, app2.ID)
	t.Logf("app2 ID: %q", app2.ID)

	assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	t.Logf("Assign application 1 to formation %s", formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
	assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	var assignedFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, assignedFormation.Name)

	expectedAssignments := map[string]map[string]fixtures.AssignmentState{
		app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	t.Logf("Assign application 2 to formation %s", formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
	assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, assignedFormation.Name)

	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
		app1.ID: {
			app1.ID: fixtures.AssignmentState{State: "READY", Config: nil},
			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
		},
		app2.ID: {
			app1.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
		},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForTenant(t, body, app1.ID, 1)

	notificationsForApp1 := gjson.GetBytes(body, app1.ID)
	assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
	assertFormationNotificationFromAssignments(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

	t.Logf("Unassign Application 1 from formation %s", formationName)
	unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	var unassignFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, unassignFormation.Name)

	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
		app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForTenant(t, body, app1.ID, 2)

	notificationsForApp1 = gjson.GetBytes(body, app1.ID)
	unassignNotificationFound := false
	for _, notification := range notificationsForApp1.Array() {
		op := notification.Get("Operation").String()
		if op == unassignOperation {
			unassignNotificationFound = true
			assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
		}
	}
	require.True(t, unassignNotificationFound, "notification for unassign app2 not found")

	t.Logf("Assign application 1 to formation %s again", formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
	assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, assignedFormation.Name)

	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
		app1.ID: {
			app1.ID: fixtures.AssignmentState{State: "READY", Config: nil},
			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
		},
		app2.ID: {
			app1.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
		},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForTenant(t, body, app1.ID, 3)

	notificationsForApp1 = gjson.GetBytes(body, app1.ID)
	assignNotificationsFound := 0
	for _, notification := range notificationsForApp1.Array() {
		op := notification.Get("Operation").String()
		if op == assignOperation {
			assignNotificationsFound++
			assertFormationNotificationFromAssignments(t, notification, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
		}
	}
	require.Equal(t, 2, assignNotificationsFound, "two notifications for assign app2 expected")

	t.Logf("Unassign Application 2 from formation %s", formationName)
	unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, unassignFormation.Name)

	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
		app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForTenant(t, body, app1.ID, 4)

	notificationsForApp1 = gjson.GetBytes(body, app1.ID)
	unassignNotificationsFound := 0
	for _, notification := range notificationsForApp1.Array() {
		op := notification.Get("Operation").String()
		if op == unassignOperation {
			unassignNotificationsFound++
			assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
		}
	}
	require.Equal(t, 2, unassignNotificationsFound, "two notifications for unassign app2 expected")

	t.Logf("Unassign Application 1 from formation %s", formationName)
	unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, unassignFormation.Name)

	assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
}

func TestUseApplicationTemplateWebhookIfAppDoesNotHaveOneForNotifications(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()
	tntParentCustomer := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // parent of `tenant.TestTenants.GetDefaultTenantID()` above

	mode := graphql.WebhookModeSync

	urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
	inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
	outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"
	formationTmplName := "app-to-app-formation-template-name"

	certSecuredHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{
					{
						Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
						PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
					},
				},
				ClientAuth:         tls.RequireAndVerifyClientCert,
				InsecureSkipVerify: conf.SkipSSLValidation,
			},
		},
	}
	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

	// TODO: Remove runtimeType once we support app only formations
	appTypes := []string{"app-to-app-app-type-1", "app-to-app-app-type-2"}
	var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, "dummy-runtime-type", appTypes, graphql.ArtifactTypeSubscription)

	formationName := "app-to-app-formation-name"
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, "app-template")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	applicationType := "app-to-app-app-type-1"
	appRegion := "test-app-region"
	appNamespace := "compass.test"
	localTenantID := "local-tenant-id"
	t.Logf("Create application template for type %q", applicationType)
	appTemplateInput := fixtures.FixApplicationTemplateWithWebhookNotifications(applicationType, localTenantID, appRegion, appNamespace, graphql.WebhookTypeApplicationTenantMapping, mode, urlTemplate, inputTemplate, outputTemplate)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
	require.NoError(t, err)

	applicationType2 := "app-to-app-app-type-2"
	localTenantID2 := "local-tenant-id2"
	t.Logf("Create application template for type %q", applicationType2)
	appTemplateInput = graphql.ApplicationTemplateInput{
		Name:        applicationType2,
		Description: &applicationType2,
		Webhooks: []*graphql.WebhookInput{
			{
				Type: graphql.WebhookTypeApplicationTenantMapping,
				Auth: &graphql.AuthInput{
					AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
				},
				Mode:           &mode,
				URLTemplate:    &urlTemplate,
				InputTemplate:  &inputTemplate,
				OutputTemplate: &outputTemplate,
			},
		},
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:          "{{name}}",
			ProviderName:  str.Ptr("compass"),
			Description:   ptr.String("test {{display-name}}"),
			LocalTenantID: &localTenantID2,
			Labels: graphql.Labels{
				"applicationType": applicationType2,
				"region":          appRegion,
			},
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name: "name",
			},
			{
				Name: "display-name",
			},
		},
		ApplicationNamespace: &appNamespace,
		AccessLevel:          graphql.ApplicationTemplateAccessLevelGlobal,
	}
	appTmpl, err = fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
	require.NoError(t, err)

	t.Logf("Create application 1 from template %q", applicationType)
	app1 := fixtures.RegisterApplicationFromTemplate(t, ctx, certSecuredGraphQLClient, applicationType, "app1-formation-notifications-tests", "App 1", tnt)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app1)
	t.Logf("app1 ID: %q", app1.ID)

	t.Logf("Create application 2 from template %q", applicationType2)
	app2 := fixtures.RegisterApplicationFromTemplate(t, ctx, certSecuredGraphQLClient, applicationType2, "app2-formation-notifications-tests", "App 2", tnt)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app2)
	t.Logf("app2 ID: %q", app2.ID)

	assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	t.Logf("Assign application 1 to formation %s", formationName)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)

	expectedAssignments := map[string]map[string]fixtures.AssignmentState{
		app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	t.Logf("Assign application 2 to formation %s", formationName)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)

	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
		app1.ID: {
			app1.ID: fixtures.AssignmentState{State: "READY", Config: nil},
			app2.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
		},
		app2.ID: {
			app1.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
		},
	}
	assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

	body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForTenant(t, body, app1.ID, 1)

	notificationsForApp1 := gjson.GetBytes(body, app1.ID)
	assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
	assertFormationNotificationFromAssignments(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

	assertNotificationsCountForTenant(t, body, app2.ID, 1)

	notificationsForApp2 := gjson.GetBytes(body, app2.ID)
	assignNotificationAboutApp1 := notificationsForApp2.Array()[0]
	assertFormationNotificationFromAssignments(t, assignNotificationAboutApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer)
}

func TestRuntimeContextToApplicationFormationAssignmentNotifications(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Runtime Context to Application formation assignment notifications", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// Register provider runtime
		providerRuntimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels: graphql.Labels{
				conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
			},
		}

		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		regionLbl, ok := providerRuntime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, regionLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		t.Log("Assert provider runtime is visible in the consumer's subaccount after successful subscription")
		consumerSubaccountRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, providerRuntime.ID)
		require.Equal(t, providerRuntime.ID, consumerSubaccountRuntime.ID)

		t.Log("Assert there is a runtime context(subscription) as part of the provider runtime")
		require.Len(t, consumerSubaccountRuntime.RuntimeContexts.Data, 1)
		require.NotEmpty(t, consumerSubaccountRuntime.RuntimeContexts.Data[0].ID)
		require.Equal(t, conf.SubscriptionLabelKey, consumerSubaccountRuntime.RuntimeContexts.Data[0].Key)
		require.Equal(t, subscriptionConsumerTenantID, consumerSubaccountRuntime.RuntimeContexts.Data[0].Value)
		rtCtx := consumerSubaccountRuntime.RuntimeContexts.Data[0]

		appTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		providerFormationTmplName := "provider-formation-template-name"
		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, appTypes, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)

		mode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		applicationType := "provider-app-type-1"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id-test"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithWebhookNotifications(applicationType, localTenantID, appRegion, appNamespace, graphql.WebhookTypeConfigurationChanged, mode, urlTemplate, inputTemplate, outputTemplate)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "rtctx-to-app-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App test",
				},
			},
		}

		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		t.Logf("Create application from template %q", applicationType)
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		app := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &app)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)

		t.Logf("Assign application to formation %s", providerFormationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, app.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app.ID: {app.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app.ID: {
				app.ID:   fixtures.AssignmentState{State: "READY", Config: nil},
				rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
			rtCtx.ID: {
				rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				app.ID:   fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
		}

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)

		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, localTenantID, 1)

		notificationsForConsumerTenant := gjson.GetBytes(body, localTenantID)
		assignNotificationForApp := notificationsForConsumerTenant.Array()[0]
		err = verifyFormationNotificationForApplication(assignNotificationForApp, assignOperation, formation.ID, rtCtx.ID, rtCtx.Value, regionLbl, "", subscriptionConsumerAccountID, emptyParentCustomerID)
		assert.NoError(t, err)

		t.Logf("Unassign Application from formation %s", providerFormationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			rtCtx.ID: {rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, localTenantID, 2)

		notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
		assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 1)

		t.Logf("Assign application to formation %s", providerFormationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, app.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		assignReq = fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		var secondAssignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &secondAssignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app.ID: {
				app.ID:   fixtures.AssignmentState{State: "READY", Config: nil},
				rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
			rtCtx.ID: {
				rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				app.ID:   fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
		}

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, localTenantID, 3)

		notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
		assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, assignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 2)

		t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app.ID: {app.ID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, localTenantID, 4)

		notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
		assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 2)

		t.Logf("Unassign Application from formation %s", providerFormationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
	})
}

func TestFormationAssignments(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Assignments", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		urlTemplateRuntime := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateRuntime := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplateRuntime := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\"}]}"
		outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, false)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async API call,
		// that's why in the function we don't delete it and add explicit defer deletion here
		// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
		defer func() {
			k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
			require.NoError(t, err)
			k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
		}()

		runtimeWebhookMode := graphql.WebhookModeAsyncCallback

		providerRuntimeInput := fixtures.FixProviderRuntimeWithWebhookInput("formation-assignment-e2e-providerRuntime", graphql.WebhookTypeConfigurationChanged, runtimeWebhookMode, urlTemplateRuntime, inputTemplateRuntime, outputTemplateRuntime, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		regionLbl, ok := providerRuntime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, regionLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		t.Log("Assert provider runtime is visible in the consumer's subaccount after successful subscription")
		consumerSubaccountRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, providerRuntime.ID)
		require.Equal(t, providerRuntime.ID, consumerSubaccountRuntime.ID)

		t.Log("Assert there is a runtime context(subscription) as part of the provider runtime")
		require.Len(t, consumerSubaccountRuntime.RuntimeContexts.Data, 1)
		require.NotEmpty(t, consumerSubaccountRuntime.RuntimeContexts.Data[0].ID)
		require.Equal(t, conf.SubscriptionLabelKey, consumerSubaccountRuntime.RuntimeContexts.Data[0].Key)
		require.Equal(t, subscriptionConsumerTenantID, consumerSubaccountRuntime.RuntimeContexts.Data[0].Value)
		rtCtx := consumerSubaccountRuntime.RuntimeContexts.Data[0]

		providerFormationTmplName := "provider-formation-template-name"
		providerAppTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, providerAppTypes, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		applicationType := "provider-app-type-1"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		appWebhookMode := graphql.WebhookModeSync
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithWebhookNotifications(applicationType, localTenantID, appRegion, appNamespace, graphql.WebhookTypeConfigurationChanged, appWebhookMode, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "app1-formation-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App 1",
				},
			},
		}

		t.Log("Create application")
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		actualApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)
		t.Logf("actualApp ID: %q", actualApp.ID)

		expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
			rtCtx.ID: {
				actualApp.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
				rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				actualApp.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}")},
			},
		}

		t.Run("Normal case notifications are sent and formation assignments are correct", func(t *testing.T) {
			var assignedFormation graphql.Formation

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			t.Logf("Assign application to formation %s", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
			assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			// rtCtx <- App notifications
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
			assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			// rtCtx -> App notifications
			assertNotificationsCountForTenant(t, body, localTenantID, 2)
			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertExpectationsForApplicationNotifications(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
				{
					op:                 assignOperation,
					formationID:        formation.ID,
					objectID:           rtCtx.ID,
					subscribedTenantID: rtCtx.Value,
					objectRegion:       regionLbl,
					configuration:      "",
					tenant:             subscriptionConsumerAccountID,
					customerID:         emptyParentCustomerID,
				},
				{
					op:                 assignOperation,
					formationID:        formation.ID,
					objectID:           rtCtx.ID,
					subscribedTenantID: rtCtx.Value,
					objectRegion:       regionLbl,
					configuration:      "{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}",
					tenant:             subscriptionConsumerAccountID,
					customerID:         emptyParentCustomerID,
				},
			})

			var unassignFormation graphql.Formation
			t.Logf("Unassign application from formation %s", formation.Name)
			unassignReq := fixtures.FixUnassignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formation.Name, assignedFormation.Name)

			application := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios := application.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			// rtCtx <- App notifications
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)
			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

			unassignNotificationFound := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
					t.Logf("Found notification for app %q", appIDFromNotification)
					if appIDFromNotification == actualApp.ID {
						unassignNotificationFound = true
						assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					}
				}
			}
			require.True(t, unassignNotificationFound)

			// rtCtx -> App notifications
			assertNotificationsCountForTenant(t, body, localTenantID, 3)
			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 1)

			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Check that application with ID %q is unassigned from formation %s", actualApp.ID, providerFormationName)
			app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.False(t, hasScenarios)

			t.Logf("Check that runtime context with ID %q is still assigned to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, unassignFormation.Name)

			t.Logf("Check that runtime context with ID %q is actually unassigned from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			assert.False(t, hasScenarios)

		})

		t.Run("Consecutive participants unassignment are still in formation before the formation assignments are processed by the async API call and removed afterwards", func(t *testing.T) {
			var assignedFormation graphql.Formation

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			t.Logf("Assign application with ID: %s to formation %s", actualApp.ID, formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
			assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)

			t.Logf("Check that the runtime context with ID: %s is assigned to formation: %s", rtCtx.ID, providerFormationName)
			actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios := actualRtmCtx.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			t.Logf("Check that the application with ID: %q is assigned to formation: %s", actualApp.ID, providerFormationName)
			app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			var unassignFormation graphql.Formation

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			unassignReq := fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, unassignFormation.Name)

			t.Logf("Unassign application with ID: %s from formation %s", actualApp.ID, formation.Name)
			unassignReq = fixtures.FixUnassignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formation.Name, assignedFormation.Name)

			t.Logf("Check that the runtime context with ID: %s is still assigned to formation: %s", rtCtx.ID, providerFormationName)
			actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			t.Logf("Check that the application with ID: %q is still assigned to formation: %s", actualApp.ID, providerFormationName)
			app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)

			actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			assert.False(t, hasScenarios)

			app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.False(t, hasScenarios)
		})

		t.Run("Application is not unassigned when only tenant is unassigned", func(t *testing.T) {
			var assignedFormation graphql.Formation

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			t.Logf("Assign application to formation %s", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)

			t.Logf("Check that runtime context with ID %q is assigned from formation %s", subscriptionConsumerAccountID, providerFormationName)
			actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios := actualRtmCtx.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			t.Logf("Check that application with ID %q is assigned from formation %s", actualApp.ID, providerFormationName)
			app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			var unassignFormation graphql.Formation

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			unassignReq := fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, unassignFormation.Name)

			t.Logf("Check that runtime context with ID %q is still assigned from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			t.Logf("Check that application with ID %q is still assigned to formation %s", actualApp.ID, providerFormationName)
			app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				actualApp.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)

			actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			assert.False(t, hasScenarios)

			t.Logf("Check that application with ID %q is still assigned to formation %s", actualApp.ID, providerFormationName)
			app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)
			assert.Contains(t, scenarios, providerFormationName)
		})
	})
}

func TestFormationApplicationTypeWhileAssigning(t *testing.T) {
	ctx := context.TODO()

	formationName := "test-formation"
	applicationName := "test-application"
	invalidApplicationType := "Not in the template"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	formation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, formationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formation.Name)

	formationTemplate := fixtures.QueryFormationTemplate(t, ctx, certSecuredGraphQLClient, formation.FormationTemplateID)

	actualApplication, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, applicationName, conf.ApplicationTypeLabelKey, invalidApplicationType, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApplication)
	require.NoError(t, err)
	require.Equal(t, invalidApplicationType, actualApplication.Labels[conf.ApplicationTypeLabelKey])

	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApplication.ID, graphql.FormationObjectTypeApplication, tenantId)
	createRequest := fixtures.FixAssignFormationRequest(actualApplication.ID, string(graphql.FormationObjectTypeApplication), formationName)
	formationResultFormation := graphql.Formation{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createRequest, &formationResultFormation)
	require.Empty(t, formationResultFormation)
	require.EqualError(t, err, fmt.Sprintf("graphql: The operation is not allowed [reason=unsupported applicationType %q for formation template %q, allowing only %q]", invalidApplicationType, formationTemplate.Name, formationTemplate.ApplicationTypes))
}

func TestFailProcessingFormationAssignmentsWhileAssigningToFormation(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Notifications", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{if eq .Operation \\\"assign\\\"}}fail-once/{{end}}{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		mode := graphql.WebhookModeSync

		providerRuntimeInput := fixtures.FixProviderRuntimeWithWebhookInput("fail-processing-while-assigning-e2e-providerRuntime", graphql.WebhookTypeConfigurationChanged, mode, urlTemplate, inputTemplate, outputTemplate, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		rtmRequest := fixtures.FixGetRuntimeContextsRequest(providerRuntime.ID)
		rtm := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, rtmRequest, &rtm)
		require.NoError(t, err)
		require.Len(t, rtm.RuntimeContexts.Data, 1)
		runtimeContextID := rtm.RuntimeContexts.Data[0].ID

		providerFormationTmplName := "provider-formation-template-name"
		providerAppTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, providerAppTypes, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		applicationType := "provider-app-type-1"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, appRegion, appNamespace)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "app1-formation-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App 1",
				},
			},
		}

		t.Log("Create application")
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		actualApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)
		t.Logf("actualApp ID: %q", actualApp.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		var assignedFormation graphql.Formation

		t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		// notification mock API should return error
		t.Logf("Assign application to formation %s should fail", formation.Name)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
		assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		// target:source:state
		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

		notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]

		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		t.Logf("Assign application to formation %s should succeed on retry", formation.Name)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
		assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]

		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		var unassignFormation graphql.Formation
		t.Logf("Unassign application from formation %s", formation.Name)
		unassignReq := fixtures.FixUnassignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formation.Name, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

		unassignNotificationFound := false
		for _, notification := range notificationsForConsumerTenant.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
				t.Logf("Found notification for app %q", appIDFromNotification)
				if appIDFromNotification == actualApp.ID {
					unassignNotificationFound = true
					assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
				}
			}
		}
		require.True(t, unassignNotificationFound)

		t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})
}

func TestFailProcessingFormationAssignmentsWhileUnassigningFromFormation(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Notifications", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{if eq .Operation \\\"unassign\\\"}}fail-once/{{end}}{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		mode := graphql.WebhookModeSync

		providerRuntimeInput := fixtures.FixProviderRuntimeWithWebhookInput("fail-processing-while-unassigning-e2e-providerRuntime", graphql.WebhookTypeConfigurationChanged, mode, urlTemplate, inputTemplate, outputTemplate, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		rtmRequest := fixtures.FixGetRuntimeContextsRequest(providerRuntime.ID)
		rtm := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, rtmRequest, &rtm)
		require.NoError(t, err)
		require.Len(t, rtm.RuntimeContexts.Data, 1)
		runtimeContextID := rtm.RuntimeContexts.Data[0].ID

		providerFormationTmplName := "provider-formation-template-name"
		providerAppTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, providerAppTypes, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		applicationType := "provider-app-type-1"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, appRegion, appNamespace)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "app1-formation-notifications-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "App 1",
				},
			},
		}

		t.Log("Create application")
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		actualApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)
		t.Logf("actualApp ID: %q", actualApp.ID)

		var assignedFormation graphql.Formation

		// Expect no formation assignments to be created
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		// Expect one formation assignment to be created
		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application to formation %s", formation.Name)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
		assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
		}

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

		notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]

		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		var unassignFormation graphql.Formation
		// notification mock api should return error
		t.Logf("Unassign application from formation %s should fail.", formation.Name)
		unassignReq := fixtures.FixUnassignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.Error(t, err)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			actualApp.ID: {
				runtimeContextID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")},
			},
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]

		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, unassignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		t.Logf("Unassign application from formation %s should succeed on retry", formation.Name)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formation.Name, assignedFormation.Name)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

		unassignNotificationFound := false
		for _, notification := range notificationsForConsumerTenant.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
				t.Logf("Found notification for app %q", appIDFromNotification)
				if appIDFromNotification == actualApp.ID {
					unassignNotificationFound = true
					assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
				}
			}
		}
		require.True(t, unassignNotificationFound)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		// Expect formation assignments to be cleared
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})
}

func TestFormationLifecycleNotifications(t *testing.T) {
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	parentTenantID := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // this is the customer parent of `tenant.TestTenants.GetDefaultTenantID()`

	urlTemplateFormation := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.Formation.ID}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"createFormation\\\"}}POST{{else}}DELETE{{end}}\\\"}"
	inputTemplateFormation := "{\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"details\\\":{\\\"id\\\":\\\"{{.Formation.ID}}\\\",\\\"name\\\":\\\"{{.Formation.Name}}\\\"}}"
	outputTemplateFormation := "{\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"

	certSecuredHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{
					{
						Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
						PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
					},
				},
				ClientAuth:         tls.RequireAndVerifyClientCert,
				InsecureSkipVerify: conf.SkipSSLValidation,
			},
		},
	}

	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

	formationTmplName := "formation-template-with-webhook-name"
	appTypes := []string{"app-type-1", "app-type-2"}
	webhookSyncMode := graphql.WebhookModeSync
	webhookFormationLifecycleType := graphql.WebhookTypeFormationLifecycle
	formationTmplInput := prepareFormationTemplateInputWithWebhook(formationTmplName, appTypes, []string{conf.KymaRuntimeTypeLabelValue}, graphql.ArtifactTypeEnvironmentInstance, webhookFormationLifecycleType, webhookSyncMode, urlTemplateFormation, inputTemplateFormation, outputTemplateFormation)

	formationTmplGQLInput, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTmplInput)
	require.NoError(t, err)
	formationTmplRequest := fixtures.FixCreateFormationTemplateRequest(formationTmplGQLInput)

	ft := graphql.FormationTemplate{}
	t.Logf("Creating formation template with name: %q with %q webhook of type %q", formationTmplName, webhookSyncMode, webhookFormationLifecycleType)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, formationTmplRequest, &ft)
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	require.NoError(t, err)

	formationName := "formation-name-from-template-with-webhook"
	t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, webhookFormationLifecycleType)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, formationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, formationName, &formationTmplName)
	require.NotEmpty(t, formation.ID)

	body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertFormationNotificationFromCreationOrDeletion(t, body, formation.ID, formation.Name, createFormationOperation, tenantID, parentTenantID)

	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

	delFormation := fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, formationName)
	require.NotEmpty(t, delFormation.ID)

	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForFormationID(t, body, formation.ID, 1)
	assertFormationNotificationFromCreationOrDeletion(t, body, formation.ID, formation.Name, deleteFormationOperation, tenantID, parentTenantID)
}

func TestFormationNotificationResynchronizationSync(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Notification Synchronous Resynchronization", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/fail-once/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		mode := graphql.WebhookModeSync

		providerRuntimeInput := fixtures.FixProviderRuntimeWithWebhookInput("formation-resynchronization-sync-e2e-providerRuntime", graphql.WebhookTypeConfigurationChanged, mode, urlTemplate, inputTemplate, outputTemplate, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)
		require.Len(t, providerRuntime.Webhooks, 1)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		rtmRequest := fixtures.FixGetRuntimeContextsRequest(providerRuntime.ID)
		rtm := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, rtmRequest, &rtm)
		require.NoError(t, err)
		require.Len(t, rtm.RuntimeContexts.Data, 1)
		runtimeContextID := rtm.RuntimeContexts.Data[0].ID

		providerFormationTmplName := "provider-formation-template-name"
		providerAppTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, providerAppTypes, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		applicationType := "provider-app-type-1"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, appRegion, appNamespace)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		t.Log("Create application")
		actualApp := fixtures.RegisterApplicationFromTemplate(t, ctx, certSecuredGraphQLClient, applicationType, "app1-formation-notifications-tests", "App 1", subscriptionConsumerAccountID)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &actualApp)
		t.Logf("actualApp ID: %q", actualApp.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, providerFormationName, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		assignedFormation := fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil}},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		// notification mock API should return error
		t.Logf("Assign application to formation %s should fail", formation.Name)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, subscriptionConsumerAccountID)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

		notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]

		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
		resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				actualApp.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
		assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]

		assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		var unassignFormation graphql.Formation
		t.Logf("Unassign application from formation %s should fail", formation.Name)
		unassignReq := fixtures.FixUnassignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), providerFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
		require.Error(t, err)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
			actualApp.ID: {
				runtimeContextID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")},
			},
		}
		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

		notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

		unassignNotificationFound := false
		for _, notification := range notificationsForConsumerTenant.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
				t.Logf("Found notification for app %q", appIDFromNotification)
				if appIDFromNotification == actualApp.ID {
					unassignNotificationFound = true
					assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
				}
			}
		}
		require.True(t, unassignNotificationFound)

		t.Logf("Check that the application with ID: %q is still assigned to formation: %s", actualApp.ID, providerFormationName)
		app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
		scenarios, hasScenarios := app.Labels["scenarios"]
		assert.True(t, hasScenarios)
		assert.Len(t, scenarios, 1)

		t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
		resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, providerFormationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			runtimeContextID: {
				runtimeContextID: fixtures.AssignmentState{State: "READY", Config: nil},
			},
		}

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)

		t.Logf("Check that the application with ID: %q is unassigned from formation %s from formation after resyonchronization", actualApp.ID, providerFormationName)
		assert.Contains(t, scenarios, providerFormationName)
		app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
		scenarios, hasScenarios = app.Labels["scenarios"]
		assert.False(t, hasScenarios)

		t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
		unassignedFormation := fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerAccountID, subscriptionConsumerSubaccountID, graphql.FormationObjectTypeTenant)
		require.Equal(t, formation.ID, unassignedFormation.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, expectedAssignments)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})
}

func TestFormationNotificationResynchronizationAsync(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("Formation Notification Asynchronous Resynchronization", func(t *testing.T) {
		ctx := context.Background()
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		urlTemplateRuntime := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-fail-once/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateRuntime := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
		outputTemplateRuntime := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\"}]}"
		outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		certSecuredHTTPClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: cc.Get()[conf.ExternalClientCertSecretName].Certificate,
							PrivateKey:  cc.Get()[conf.ExternalClientCertSecretName].PrivateKey,
						},
					},
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: conf.SkipSSLValidation,
				},
			},
		}
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, false)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async API call,
		// that's why in the function we don't delete it and add explicit defer deletion here
		// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
		defer func() {
			k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
			require.NoError(t, err)
			k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
		}()

		runtimeWebhookMode := graphql.WebhookModeAsyncCallback
		webhookThatFailsOnce := &graphql.WebhookInput{
			Type: graphql.WebhookTypeConfigurationChanged,
			Auth: &graphql.AuthInput{
				AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
			},
			Mode:           &runtimeWebhookMode,
			URLTemplate:    &urlTemplateRuntime,
			InputTemplate:  &inputTemplateRuntime,
			OutputTemplate: &outputTemplateRuntime,
		}
		providerRuntimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels: graphql.Labels{
				conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
			},
			Webhooks: []*graphql.WebhookInput{webhookThatFailsOnce},
		}

		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		regionLbl, ok := providerRuntime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, regionLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		t.Log("Assert provider runtime is visible in the consumer's subaccount after successful subscription")
		consumerSubaccountRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, providerRuntime.ID)
		require.Equal(t, providerRuntime.ID, consumerSubaccountRuntime.ID)

		t.Log("Assert there is a runtime context(subscription) as part of the provider runtime")
		require.Len(t, consumerSubaccountRuntime.RuntimeContexts.Data, 1)
		require.NotEmpty(t, consumerSubaccountRuntime.RuntimeContexts.Data[0].ID)
		require.Equal(t, conf.SubscriptionLabelKey, consumerSubaccountRuntime.RuntimeContexts.Data[0].Key)
		require.Equal(t, subscriptionConsumerTenantID, consumerSubaccountRuntime.RuntimeContexts.Data[0].Value)
		rtCtx := consumerSubaccountRuntime.RuntimeContexts.Data[0]

		providerFormationTmplName := "provider-formation-template-name"
		providerAppTypes := []string{"provider-app-type-1", "provider-app-type-2"}
		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, providerAppTypes, graphql.ArtifactTypeSubscription)

		providerFormationName := "provider-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", providerFormationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerFormationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		applicationType := "provider-app-type-1"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		appWebhookMode := graphql.WebhookModeSync
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := fixtures.FixApplicationTemplateWithWebhookNotifications(applicationType, localTenantID, appRegion, appNamespace, graphql.WebhookTypeConfigurationChanged, appWebhookMode, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		t.Log("Create application")
		actualApp := fixtures.RegisterApplicationFromTemplate(t, ctx, certSecuredGraphQLClient, applicationType, "app1-formation-notifications-tests", "App 1", subscriptionConsumerAccountID)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &actualApp)
		t.Logf("actualApp ID: %q", actualApp.ID)

		t.Run("Resynchronize when in CREATE_ERROR and DELETE_ERROR should resend notifications and succeed", func(t *testing.T) {
			resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
			defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, providerFormationName, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			assignedFormation := fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

			t.Logf("Assign application to formation %s", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, subscriptionConsumerAccountID)

			expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: nil},
					rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				},
				actualApp.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: nil},
					rtCtx.ID:     fixtures.AssignmentState{State: "CREATE_ERROR", Config: str.Ptr(`{"error":{"message":"test error","errorCode":2}}`)},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{{
					Message:   "test error",
					ErrorCode: 2,
				}},
			})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			// rtCtx <- App notifications
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
			assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			// rtCtx -> App notifications
			assertNotificationsCountForTenant(t, body, localTenantID, 1)
			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertExpectationsForApplicationNotifications(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
				{
					op:                 assignOperation,
					formationID:        formation.ID,
					objectID:           rtCtx.ID,
					subscribedTenantID: rtCtx.Value,
					objectRegion:       regionLbl,
					configuration:      "",
					tenant:             subscriptionConsumerAccountID,
					customerID:         emptyParentCustomerID,
				},
			})

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
					rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				},
				actualApp.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: nil},
					rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: str.Ptr(`{"asyncKey":"asyncValue","asyncKey2":{"asyncNestedKey":"asyncNestedValue"}}`)},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

			resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
			defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Unassign application from formation %s", formation.Name)
			unassignedFormation := fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerAccountID, actualApp.ID, graphql.FormationObjectTypeApplication)
			require.Equal(t, formation.ID, unassignedFormation.ID)

			application := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios := application.Labels["scenarios"]
			require.True(t, hasScenarios)
			require.Len(t, scenarios, 1)
			require.Contains(t, scenarios, providerFormationName)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			// rtCtx <- App notifications
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)
			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

			unassignNotificationFound := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
					t.Logf("Found notification for app %q", appIDFromNotification)
					if appIDFromNotification == actualApp.ID {
						unassignNotificationFound = true
						assertFormationNotificationFromAssignments(t, notification, unassignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					}
				}
			}
			require.True(t, unassignNotificationFound)

			// rtCtx -> App notifications
			assertNotificationsCountForTenant(t, body, localTenantID, 3)
			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 1)

			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
				actualApp.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: str.Ptr("{\"error\":{\"message\":\"test error\",\"errorCode\":2}}")},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{{
					Message:   "test error",
					ErrorCode: 2,
				}},
			})

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)

			t.Logf("Check that application with ID %q is unassigned from formation %s", actualApp.ID, providerFormationName)
			app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApp.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			require.False(t, hasScenarios)

			t.Logf("Check that runtime context with ID %q is still assigned to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			require.True(t, hasScenarios)
			require.Len(t, scenarios, 1)
			require.Contains(t, scenarios, providerFormationName)

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			unassignedFormation = fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerAccountID, subscriptionConsumerSubaccountID, graphql.FormationObjectTypeTenant)
			require.Equal(t, formation.ID, unassignedFormation.ID)

			t.Logf("Check that runtime context with ID %q is actually unassigned from formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
			scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
			require.False(t, hasScenarios)
		})
		t.Run("Resynchronize when in INITIAL and DELETING should resend notifications and succeed", func(t *testing.T) {

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookID := providerRuntime.Webhooks[0].ID
			require.NotNil(t, webhookID)

			webhookThatNeverResponds := &graphql.WebhookInput{
				Type: graphql.WebhookTypeConfigurationChanged,
				Auth: &graphql.AuthInput{
					AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
				},
				Mode:           &runtimeWebhookMode,
				URLTemplate:    str.Ptr("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-no-response/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"),
				InputTemplate:  &inputTemplateRuntime,
				OutputTemplate: &outputTemplateRuntime,
			}

			webhookThatNeverRespondsInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(webhookThatNeverResponds)
			require.NoError(t, err)

			t.Logf("Update webhook with ID %q to have URLTemlate that points to endpoint which never responds", webhookID)
			updateWebhookRequest := fixtures.FixUpdateWebhookRequest(webhookID, webhookThatNeverRespondsInStr)
			actualWebhook := graphql.Webhook{}
			err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateWebhookRequest, &actualWebhook)
			require.NoError(t, err)
			require.Equal(t, actualWebhook.ID, webhookID)

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, providerFormationName)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, providerFormationName, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			assignedFormation := fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

			t.Logf("Assign application to formation %s", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, actualApp.ID, subscriptionConsumerAccountID)

			expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: nil},
					rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				},
				actualApp.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: nil},
					rtCtx.ID:     fixtures.AssignmentState{State: "INITIAL", Config: nil},
				},
			}

			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			// rtCtx <- App notifications
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
			assertFormationNotificationFromAssignments(t, assignNotificationForApp1, assignOperation, formation.ID, actualApp.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			// rtCtx -> App notifications
			assertNotificationsCountForTenant(t, body, localTenantID, 1)
			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertExpectationsForApplicationNotifications(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
				{
					op:                 assignOperation,
					formationID:        formation.ID,
					objectID:           rtCtx.ID,
					subscribedTenantID: rtCtx.Value,
					objectRegion:       regionLbl,
					configuration:      "",
					tenant:             subscriptionConsumerAccountID,
					customerID:         emptyParentCustomerID,
				},
			})

			webhookThatSucceeds := &graphql.WebhookInput{
				Type: graphql.WebhookTypeConfigurationChanged,
				Auth: &graphql.AuthInput{
					AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
				},
				Mode:           &runtimeWebhookMode,
				URLTemplate:    str.Ptr("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"),
				InputTemplate:  &inputTemplateRuntime,
				OutputTemplate: &outputTemplateRuntime,
			}

			webhookThatSucceedsInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(webhookThatSucceeds)
			require.NoError(t, err)

			t.Logf("Update webhook with ID %q to have URLTemlate that responds with success", webhookID)
			updateWebhookRequest = fixtures.FixUpdateWebhookRequest(webhookID, webhookThatSucceedsInStr)
			err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateWebhookRequest, &actualWebhook)
			require.NoError(t, err)
			require.Equal(t, actualWebhook.ID, webhookID)
			require.NotNil(t, actualWebhook.URLTemplate)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")},
					rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: nil},
				},
				actualApp.ID: {
					actualApp.ID: fixtures.AssignmentState{State: "READY", Config: nil},
					rtCtx.ID:     fixtures.AssignmentState{State: "READY", Config: str.Ptr(`{"asyncKey":"asyncValue","asyncKey2":{"asyncNestedKey":"asyncNestedValue"}}`)},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

			t.Logf("Update webhook with ID %q to have URLTemlate that points to endpoint which never responds", webhookID)
			updateWebhookRequest = fixtures.FixUpdateWebhookRequest(webhookID, webhookThatNeverRespondsInStr)
			err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateWebhookRequest, &actualWebhook)
			require.NoError(t, err)
			require.Equal(t, actualWebhook.ID, webhookID)
			require.NotNil(t, actualWebhook.URLTemplate)

			t.Logf("Unassign application from formation %s", formation.Name)
			unassignedFormation := fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerFormationName}, subscriptionConsumerAccountID, actualApp.ID, graphql.FormationObjectTypeApplication)
			require.Equal(t, formation.ID, unassignedFormation.ID)

			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
				actualApp.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "DELETING", Config: str.Ptr(`{"asyncKey":"asyncValue","asyncKey2":{"asyncNestedKey":"asyncNestedValue"}}`)},
				},
			}

			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

			t.Logf("Update webhook with ID %q to have URLTemlate that responds with success", webhookID)
			updateWebhookRequest = fixtures.FixUpdateWebhookRequest(webhookID, webhookThatSucceedsInStr)
			err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, updateWebhookRequest, &actualWebhook)
			require.NoError(t, err)
			require.Equal(t, actualWebhook.ID, webhookID)
			require.NotNil(t, actualWebhook.URLTemplate)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, providerFormationName, assignedFormation.Name)

			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)
		})
	})
}

func prepareFormationTemplateInputWithWebhook(formationTemplateName string, applicationTypes, runtimeTypes []string, runtimeArtifactKind graphql.ArtifactType, webhookType graphql.WebhookType, webhookMode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string) graphql.FormationTemplateInput {
	return graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: formationTemplateName,
		RuntimeArtifactKind:    runtimeArtifactKind,
		Webhooks: []*graphql.WebhookInput{
			{
				Type: webhookType,
				Auth: &graphql.AuthInput{
					AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
				},
				Mode:           &webhookMode,
				URLTemplate:    &urlTemplate,
				InputTemplate:  &inputTemplate,
				OutputTemplate: &outputTemplate,
			},
		},
	}
}

func assertNotificationsCountForTenant(t *testing.T, body []byte, tenantID string, count int) {
	assertNotificationsCount(t, body, tenantID, count)
}

func assertNotificationsCountForFormationID(t *testing.T, body []byte, formationID string, count int) {
	assertNotificationsCount(t, body, formationID, count)
}

func assertNotificationsCount(t *testing.T, body []byte, objectID string, count int) {
	notifications := gjson.GetBytes(body, objectID)
	require.True(t, notifications.Exists())
	require.Len(t, notifications.Array(), count)
}

func cleanupNotificationsFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/formation-callback/cleanup", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func resetShouldFailEndpointFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/formation-callback/reset-should-fail", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func getNotificationsFromExternalSvcMock(t *testing.T, client *http.Client) []byte {
	t.Logf("Getting formation notifications recieved in external services mock")
	resp, err := client.Get(conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusOK, string(body)))
	return body
}

func assertFormationNotificationFromAssignments(t *testing.T, notification gjson.Result, op, formationID, expectedAppID, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string) {
	require.Equal(t, op, notification.Get("Operation").String())
	if op == unassignOperation {
		require.Equal(t, expectedAppID, notification.Get("ApplicationID").String())
	}
	require.Equal(t, formationID, notification.Get("RequestBody.ucl-formation-id").String())
	require.Equal(t, expectedTenant, notification.Get("RequestBody.globalAccountId").String())
	require.Equal(t, expectedCustomerID, notification.Get("RequestBody.crmId").String())

	notificationItems := notification.Get("RequestBody.items")
	require.True(t, notificationItems.Exists())
	require.Equal(t, 1, len(notificationItems.Array()))

	app1FromNotification := notificationItems.Array()[0]
	require.Equal(t, expectedAppID, app1FromNotification.Get("ucl-system-tenant-id").String())
	require.Equal(t, expectedLocalTenantID, app1FromNotification.Get("tenant-id").String())
	require.Equal(t, expectedAppNamespace, app1FromNotification.Get("application-namespace").String())
	require.Equal(t, expectedAppRegion, app1FromNotification.Get("region").String())
}

func assertFormationNotificationFromCreationOrDeletion(t *testing.T, body []byte, formationID, formationName, formationOperation, tenantID, parentTenantID string) {
	notificationsForFormation := gjson.GetBytes(body, formationID)
	require.True(t, notificationsForFormation.Exists())
	require.Len(t, notificationsForFormation.Array(), 1)

	notificationForFormation := notificationsForFormation.Array()[0]
	require.Equal(t, formationOperation, notificationForFormation.Get("Operation").String())
	require.Equal(t, tenantID, notificationForFormation.Get("RequestBody.globalAccountId").String())
	require.Equal(t, parentTenantID, notificationForFormation.Get("RequestBody.crmId").String())

	notificationForFormationDetails := notificationForFormation.Get("RequestBody.details")
	require.True(t, notificationForFormationDetails.Exists())
	require.Equal(t, formationID, notificationForFormationDetails.Get("id").String())
	require.Equal(t, formationName, notificationForFormationDetails.Get("name").String())
}

func assertSeveralFormationAssignmentsNotifications(t *testing.T, notificationsForConsumerTenant gjson.Result, rtCtx *graphql.RuntimeContextExt, formationID, region, operationType, expectedTenant, expectedCustomerID string, expectedNumberOfNotifications int) {
	actualNumberOfNotifications := 0
	for _, notification := range notificationsForConsumerTenant.Array() {
		rtCtxIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
		op := notification.Get("Operation").String()
		t.Logf("Found notification about rtCtx %q", rtCtxIDFromNotification)
		if rtCtxIDFromNotification == rtCtx.ID && op == operationType {
			actualNumberOfNotifications++
			err := verifyFormationNotificationForApplication(notification, operationType, formationID, rtCtx.ID, rtCtx.Value, region, "", expectedTenant, expectedCustomerID)
			assert.NoError(t, err)
		}
	}
	require.Equal(t, expectedNumberOfNotifications, actualNumberOfNotifications)
}

type applicationFormationExpectations struct {
	op                 string
	formationID        string
	objectID           string
	subscribedTenantID string
	objectRegion       string
	configuration      string
	tenant             string
	customerID         string
}

func assertExpectationsForApplicationNotifications(t *testing.T, notifications []gjson.Result, expectations []*applicationFormationExpectations) {
	assert.Equal(t, len(expectations), len(notifications))
	for _, expectation := range expectations {
		found := false
		for _, notification := range notifications {
			err := verifyFormationNotificationForApplication(notification, expectation.op, expectation.formationID, expectation.objectID, expectation.subscribedTenantID, expectation.objectRegion, expectation.configuration, expectation.tenant, expectation.customerID)
			if err == nil {
				found = true
			}
		}
		assert.Truef(t, found, "Did not match expectations for notification %v", expectation)
	}
}

func verifyFormationNotificationForApplication(notification gjson.Result, op, formationID, expectedObjectID, expectedSubscribedTenantID, expectedObjectRegion, expectedConfiguration, expectedTenant, expectedCustomerID string) error {
	actualOp := notification.Get("Operation").String()
	if op != actualOp {
		return errors.Errorf("Operation does not match: expected %q, but got %q", op, actualOp)
	}

	if op == unassignOperation {
		actualObjectID := notification.Get("ApplicationID").String()
		if expectedObjectID != actualObjectID {
			return errors.Errorf("ObjectID does not match: expected %q, but got %q", expectedObjectID, actualObjectID)
		}
	}

	actualFormationID := notification.Get("RequestBody.ucl-formation-id").String()
	if formationID != actualFormationID {
		return errors.Errorf("FormationID does not match: expected %q, but got %q", formationID, actualFormationID)
	}

	actualTenantID := notification.Get("RequestBody.globalAccountId").String()
	if actualTenantID != expectedTenant {
		return errors.Errorf("Global Account does not match: expected %q, but got %q", expectedTenant, actualTenantID)
	}

	actualCustomerID := notification.Get("RequestBody.crmId").String()
	if actualCustomerID != expectedCustomerID {
		return errors.Errorf("Customer ID does not match: expected %q, but got %q", expectedCustomerID, actualCustomerID)
	}

	notificationItems := notification.Get("RequestBody.items")
	if !notificationItems.Exists() {
		return errors.Errorf("NotificationItems do not exist")
	}

	actualItemsLength := len(notificationItems.Array())
	if actualItemsLength != 1 {
		return errors.Errorf("Items count does not match: expected %q, but got %q", 1, actualItemsLength)
	}

	rtCtxFromNotification := notificationItems.Array()[0]

	actualSubscribedTenantID := rtCtxFromNotification.Get("application-tenant-id").String()
	if expectedSubscribedTenantID != actualSubscribedTenantID {
		return errors.Errorf("SubscribeTenantID does not match: expected %q, but got %q", expectedSubscribedTenantID, rtCtxFromNotification.Get("application-tenant-id").String())
	}

	actualObjectRegion := rtCtxFromNotification.Get("region").String()
	if expectedObjectRegion != actualObjectRegion {
		return errors.Errorf("ObjectRegion does not match: expected %q, but got %q", expectedObjectRegion, actualObjectRegion)
	}
	if expectedConfiguration != "" && notification.Get("RequestBody.config").String() != expectedConfiguration {
		return errors.Errorf("config does not match: expected %q, but got %q", expectedConfiguration, notification.Get("RequestBody.config").String())
	}

	return nil
}

func validateRuntimesScenariosLabels(t *testing.T, ctx context.Context, subscriptionConsumerAccountID, kymaFormationName, providerFormationName, kymaRuntimeID, providerRuntimeID string) {
	t.Log("Assert kyma runtime HAS only kyma scenarios label")
	checkRuntimeFormationLabelsExists(t, ctx, subscriptionConsumerAccountID, kymaRuntimeID, ScenariosLabel, []string{kymaFormationName})

	t.Log("Assert provider runtime is NOT part of any scenarios")
	checkRuntimeFormationLabelIsMissing(t, ctx, subscriptionConsumerAccountID, providerRuntimeID)

	t.Log("Assert runtime context of the provider runtime HAS only provider scenarios label")
	checkRuntimeContextFormationLabelsForRuntime(t, ctx, subscriptionConsumerAccountID, providerRuntimeID, ScenariosLabel, []string{providerFormationName})
}

func TestFormationRuntimeTypeWhileAssigning(t *testing.T) {
	ctx := context.TODO()

	formationTemplateName := "new-formation-template"
	runtimeType := "some-new-runtime-type"
	formationName := "test-formation"
	runtimeName := "test-runtime"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	formationTemplateInput := fixtures.FixFormationTemplateInputWithType(formationTemplateName, runtimeType)
	var actualFormationTemplate *graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, actualFormationTemplate)
	actualFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	formation := fixtures.FixFormationInput(formationName, str.Ptr(formationTemplateName))
	formationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(formation)
	require.NoError(t, err)

	createFormationReq := fixtures.FixCreateFormationWithTemplateRequest(formationInputGQL)
	actualFormation := graphql.Formation{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createFormationReq, &actualFormation)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formation.Name)
	require.NoError(t, err)

	inRuntime := fixtures.FixRuntimeRegisterInput(runtimeName)
	var actualRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &actualRuntime)
	actualRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, inRuntime, conf.GatewayOauth)
	require.Equal(t, conf.KymaRuntimeTypeLabelValue, actualRuntime.Labels[conf.RuntimeTypeLabelKey])

	createRequest := fixtures.FixAssignFormationRequest(actualRuntime.ID, string(graphql.FormationObjectTypeRuntime), formationName)
	formationResultFormation := graphql.Formation{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createRequest, &formationResultFormation)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, formation, actualRuntime.ID, graphql.FormationObjectTypeRuntime, tenantId)
	require.Empty(t, formationResultFormation)
	require.EqualError(t, err, "graphql: The operation is not allowed [reason=unsupported runtimeType \"kyma\" for formation template \"new-formation-template\", allowing only [\"some-new-runtime-type\"]]")

	runtimeCtx := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, actualRuntime.ID, "testRuntimeCtxKey", "testRuntimeCtxValue")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtimeCtx.ID)
	createRuntimeContextAssignRequest := fixtures.FixAssignFormationRequest(runtimeCtx.ID, string(graphql.FormationObjectTypeRuntimeContext), formationName)
	formationResultForContextFormation := graphql.Formation{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createRuntimeContextAssignRequest, &formationResultForContextFormation)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, formation, runtimeCtx.ID, graphql.FormationObjectTypeRuntimeContext, tenantId)
	require.Empty(t, formationResultForContextFormation)
	require.EqualError(t, err, "graphql: The operation is not allowed [reason=unsupported runtimeType \"kyma\" for formation template \"new-formation-template\", allowing only [\"some-new-runtime-type\"]]")
}

func assignTenantToFormation(t *testing.T, ctx context.Context, objectID, tenantID, formationName string) {
	t.Logf("Assign tenant: %q to formation with name: %q...", objectID, formationName)
	assignReq := fixtures.FixAssignFormationRequest(objectID, string(graphql.FormationObjectTypeTenant), formationName)
	var formation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, assignReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
	t.Logf("Successfully assigned tenant %s to formation %s", objectID, formationName)
}

func unassignTenantFromFormation(t *testing.T, ctx context.Context, objectID, tenantID, formationName string) {
	t.Logf("Unassign tenant: %q from formation with name: %q...", objectID, formationName)
	unassignReq := fixtures.FixUnassignFormationRequest(objectID, string(graphql.FormationObjectTypeTenant), formationName)
	var formation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, unassignReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
	t.Logf("Successfully unassigned tenant: %q from formation with name: %q", objectID, formationName)
}

func createFormationTemplateWithMultipleRuntimeTypes(t *testing.T, ctx context.Context, formationTemplateName string, runtimeTypes []string, applicationTypes []string, runtimeArtifactKind graphql.ArtifactType) graphql.FormationTemplate {
	formationTmplInput := graphql.FormationTemplateInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       applicationTypes,
		RuntimeTypes:           runtimeTypes,
		RuntimeTypeDisplayName: formationTemplateName,
		RuntimeArtifactKind:    runtimeArtifactKind,
	}

	formationTmplGQLInput, err := testctx.Tc.Graphqlizer.FormationTemplateInputToGQL(formationTmplInput)
	require.NoError(t, err)
	formationTmplRequest := fixtures.FixCreateFormationTemplateRequest(formationTmplGQLInput)

	ft := graphql.FormationTemplate{}
	t.Logf("Creating formation template with name: %q", formationTemplateName)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, formationTmplRequest, &ft)
	require.NoError(t, err)
	return ft
}

func checkRuntimeFormationLabelsExists(t *testing.T, ctx context.Context, tenantID string, rtmID, formationLabelKey string, expectedFormations []string) {
	runtimeRequest := fixtures.FixGetRuntimeRequest(rtmID)
	rtm := graphql.RuntimeExt{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, runtimeRequest, &rtm)
	require.NoError(t, err)

	scenariosLabel, ok := rtm.Labels[formationLabelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.ElementsMatch(t, expectedFormations, actualScenariosEnum)
}

func checkRuntimeFormationLabelIsMissing(t *testing.T, ctx context.Context, tenantID, rtmID string) {
	rtmRequest := fixtures.FixGetRuntimeRequest(rtmID)
	rtm := graphql.RuntimeExt{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, rtmRequest, &rtm)
	require.NoError(t, err)
	require.Equal(t, rtmID, rtm.ID)

	scenariosLabel, hasScenario := rtm.Labels[ScenariosLabel].([]interface{})
	require.False(t, hasScenario)
	require.Empty(t, scenariosLabel)
}

func checkRuntimeContextFormationLabelsForRuntime(t *testing.T, ctx context.Context, tenantID, rtmID, formationLabelKey string, expectedFormations []string) {
	rtmRequest := fixtures.FixGetRuntimeContextsRequest(rtmID)
	rtm := graphql.RuntimeExt{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, rtmRequest, &rtm)
	require.NoError(t, err)
	require.Equal(t, rtmID, rtm.ID)
	require.NotEmpty(t, rtm.RuntimeContexts)
	require.NotEmpty(t, rtm.RuntimeContexts.Data)

	for _, rtCtx := range rtm.RuntimeContexts.Data {
		scenariosLabel, ok := rtCtx.Labels[formationLabelKey].([]interface{})
		require.True(t, ok)

		var actualScenariosEnum []string
		for _, v := range scenariosLabel {
			actualScenariosEnum = append(actualScenariosEnum, v.(string))
		}
		assert.ElementsMatch(t, expectedFormations, actualScenariosEnum)
	}
}

func checkRuntimeContextFormationLabels(t *testing.T, ctx context.Context, tenantID, rtmID, rtmCtxID, formationLabelKey string, expectedFormations []string) {
	rtmRequest := fixtures.FixRuntimeContextRequest(rtmID, rtmCtxID)
	rtm := graphql.RuntimeExt{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, rtmRequest, &rtm)
	require.NoError(t, err)
	require.Equal(t, rtmID, rtm.ID)
	require.NotEmpty(t, rtm.RuntimeContext)

	scenariosLabel, ok := rtm.RuntimeContext.Labels[formationLabelKey].([]interface{})
	require.True(t, ok)

	var actualScenariosEnum []string
	for _, v := range scenariosLabel {
		actualScenariosEnum = append(actualScenariosEnum, v.(string))
	}
	assert.ElementsMatch(t, expectedFormations, actualScenariosEnum)
}

func assertFormationAssignments(t *testing.T, ctx context.Context, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)

	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q", assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q and target %q", assignment.Source, assignment.Target)

		require.Equal(t, assignmentExpectation.State, assignment.State)
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Config), str.PtrStrToStr(assignment.Value))
	}
}

func assertFormationAssignmentsAsynchronously(t *testing.T, ctx context.Context, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	t.Logf("Sleeping for %d seconds while the async formation assignment status is proccessed...", conf.FormationMappingAsyncResponseDelay+1)
	time.Sleep(time.Second * time.Duration(conf.FormationMappingAsyncResponseDelay+1))
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data

	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q and source %q", assignment.ID, assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q, source %q and target %q", assignment.ID, assignment.Source, assignment.Target)

		require.Equal(t, assignmentExpectation.State, assignment.State, "Assignment with ID: %q has different state than expected", assignment.ID)

		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Config), str.PtrStrToStr(assignment.Value))
	}

	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
}

func assertFormationStatus(t *testing.T, ctx context.Context, tenant, formationID string, expectedFormationStatus graphql.FormationStatus) {
	// Get the formation with its status
	var gotFormation graphql.FormationWithStatus
	getFormationReq := fixtures.FixGetFormationRequest(formationID)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant, getFormationReq, &gotFormation)
	require.NoError(t, err)

	// Assert the status
	require.Equal(t, expectedFormationStatus.Condition, gotFormation.Status.Condition)

	if expectedFormationStatus.Errors == nil {
		require.Nil(t, gotFormation.Status.Errors)
	} else { // assert only the Message and ErrorCode
		for i := range expectedFormationStatus.Errors {
			require.Equal(t, expectedFormationStatus.Errors[i].Message, gotFormation.Status.Errors[i].Message)
			require.Equal(t, expectedFormationStatus.Errors[i].ErrorCode, gotFormation.Status.Errors[i].ErrorCode)
		}
	}
}
