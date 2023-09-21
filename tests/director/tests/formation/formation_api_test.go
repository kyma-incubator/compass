package formation

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ScenariosLabel            = "scenarios"
	assignFormationCategory   = "assign formation"
	unassignFormationCategory = "unassign formation"
	resourceSubtypeANY        = "ANY"
	exceptionSystemType       = "exception-type"
	testScenario              = "testScenario"
	reset                     = true
	dontReset                 = false
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

	example.SaveExample(t, createReq.Query(), "create formation")

	t.Logf("Should get formation with name: %q by ID: %q", testScenario, formation.ID)
	var gotFormation graphql.Formation
	getFormationReq := fixtures.FixGetFormationRequest(formation.ID)
	example.SaveExample(t, getFormationReq.Query(), "query formation")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationReq, &gotFormation)
	require.NoError(t, err)
	require.Equal(t, formation, gotFormation)

	t.Logf("Should get formation by name: %q", formation.Name)
	getFormationByNameReq := fixtures.FixGetFormationByNameRequest(formation.Name)
	example.SaveExample(t, getFormationByNameReq.Query(), "query formation by name")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getFormationByNameReq, &gotFormation)
	require.NoError(t, err)
	require.Equal(t, formation, gotFormation)

	t.Logf("Should delete formation with name: %q", testScenario)
	deleteFormation := fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	assert.Equal(t, formation, *deleteFormation)

	t.Logf("Should NOT get formation with name: %q by ID: %q because it is already deleted", testScenario, formation.ID)
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
	example.SaveExample(t, listFormationsReq.Query(), "query formations")
	formationPage1 := fixtures.ListFormations(t, ctx, certSecuredGraphQLClient, listFormationsReq)
	require.Equal(t, expectedFormations, formationPage1.TotalCount)
	require.Empty(t, formationPage1.Data)

	t.Logf("Should create formation: %q", firstFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, firstFormationName)
	firstFormation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, firstFormationName)

	t.Logf("Should create formation: %q", secondFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, secondFormationName)
	secondFormation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, secondFormationName)

	expectedFormations = 2
	t.Logf("List should return %d formations", expectedFormations)
	formationPage2 := fixtures.ListFormations(t, ctx, certSecuredGraphQLClient, listFormationsReq)
	require.Equal(t, expectedFormations, formationPage2.TotalCount)
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
	app := graphql.ApplicationExt{} // needed so the 'defer' can be above the application creation
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "app", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	t.Logf("Should create formation: %s", newFormation)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign application to formation %s, as it is not in the label definition", nonExistingFormation)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: nonExistingFormation}, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: nonExistingFormation}, app.ID, tenantId)

	t.Logf("Assign application to formation %s", newFormation)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: nonExistingFormation}, app.ID, tenantId)
	assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	example.SaveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign application to formation")

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

	example.SaveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign application from formation")

	t.Log("Should be able to delete formation after application is unassigned")
	fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)

	example.SaveExample(t, deleteRequest.Query(), "delete formation")
}

func TestApplicationOnlyFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	newFormation := "ADDITIONAL"

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	t.Log("Create formation template")
	input := graphql.FormationTemplateInput{Name: "application-only-formation-template", ApplicationTypes: []string{string(util.ApplicationTypeC4C)}}
	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, input)

	t.Logf("Should create formation: %s", newFormation)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, newFormation)
	fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation, &formationTemplate.Name)

	t.Log("Create application")
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "app", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	formationInput := graphql.FormationInput{Name: newFormation, TemplateName: &formationTemplate.Name}

	t.Logf("Assign application to formation %s", newFormation)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app.ID, tenantId)

	t.Logf("Create runtime")
	rtmName := "rt"
	rtmInput := fixtures.FixRuntimeRegisterInputWithoutLabels(rtmName)

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, rtmInput, conf.GatewayOauth)

	t.Logf("Should fail to assign runtime to formation %s", newFormation)
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formationInput, runtime.ID, tenantId)
	fixtures.AssignFormationWithRuntimeObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, formationInput, runtime.ID, tenantId)

	t.Log("Create runtimeContext")
	runtimeContext := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, subaccountID, runtime.ID, "AppOnlyFormationsTest", "AppOnlyFormationTestTest")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtimeContext.ID)

	t.Logf("Should fail to assign runtime context to formation %s", newFormation)
	defer fixtures.UnassignFormationWithRuntimeContextObjectType(t, ctx, certSecuredGraphQLClient, formationInput, runtimeContext.ID, tenantId)
	fixtures.AssignFormationWithRuntimeContextObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, formationInput, runtimeContext.ID, tenantId)

	t.Logf("Should fail to assign tenant to formation %s", newFormation)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, formationInput, subaccountID, tenantId)
}

func TestRuntimeFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	newFormation := "ADDITIONAL"
	asaFormation := "ASA"
	unusedFormationName := "UNUSED"

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	t.Logf("Should create formation: %s", asaFormation)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, asaFormation)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, asaFormation)

	asaFormationInput := graphql.FormationInput{Name: asaFormation}
	t.Log("Creating ASA")
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, asaFormationInput.Name, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, asaFormationInput, subaccountID, tenantId)

	rtmName := "rt"
	rtmDesc := "rt-description"
	rtmInput := fixtures.FixRuntimeRegisterInputWithoutLabels(rtmName)
	rtmInput.Description = &rtmDesc
	rtmInput.Labels[conf.GlobalSubaccountIDLabelKey] = subaccountID

	var rtm graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &rtm)
	rtm = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, rtmInput, conf.GatewayOauth)

	t.Logf("Should create formation: %s", unusedFormationName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, unusedFormationName)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, unusedFormationName)

	t.Logf("Should create formation: %s", newFormation)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)

	nonExistingFormation := "nonExistingFormation"
	t.Logf("Shoud not assign runtime to formation %s, as it is not in the label definition", nonExistingFormation)
	nonExistingFormationInput := graphql.FormationInput{Name: nonExistingFormation}
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, nonExistingFormationInput, rtm.ID, tenantId)
	fixtures.AssignFormationWithRuntimeObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, nonExistingFormationInput, rtm.ID, tenantId)

	t.Logf("Assign runtime to formation %s", newFormation)
	newFormationInput := graphql.FormationInput{Name: newFormation}
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, rtm.ID, tenantId)
	assignReq := fixtures.FixAssignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), newFormation)
	var assignFormation graphql.Formation
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	example.SaveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign runtime to formation")

	t.Log("Check if new scenario label value was set correctly")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation, newFormation})

	t.Logf("Assign runtime to formation %s which was already assigned by ASA should succeed", asaFormation)
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, asaFormationInput, rtm.ID, tenantId)
	fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, asaFormationInput, rtm.ID, tenantId)

	t.Log("Check if the formation label value is still assigned")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation, newFormation})

	t.Logf("Try to unassign runtime from formation %q which was assigned by ASA", asaFormation)
	fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, asaFormationInput, rtm.ID, tenantId)

	t.Log("Check that the formation label value is still assigned")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation, newFormation})

	t.Log("Should not delete formation while runtime is assigned")
	fixtures.DeleteFormationWithinTenantExpectError(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)

	t.Logf("Unassign Runtime from formation %s", newFormation)
	unassignFormation := graphql.Formation{}
	unassignReq := fixtures.FixUnassignFormationRequest(rtm.ID, string(graphql.FormationObjectTypeRuntime), newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unassignReq, &unassignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, unassignFormation.Name)

	example.SaveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign runtime from formation")

	t.Log("Check that the formation label value is unassigned")
	checkRuntimeFormationLabelsExists(t, ctx, tenantId, rtm.ID, ScenariosLabel, []string{asaFormation})

	t.Log("Should be able to delete formation after runtime is unassigned")
	fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, newFormation)

	t.Log("Should be able to delete formation")
	fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, unusedFormationName)

	fixtures.DeleteFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, asaFormationInput.Name, subaccountID, tenantId)
}

func TestRuntimeContextFormationFlow(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	newFormation := "ADDITIONAL"
	asaFormation := "ASA"
	asaFormation2 := "ASA2"

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
			conf.GlobalSubaccountIDLabelKey: subaccountID,
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

	example.SaveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign runtime context to formation")

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

	example.SaveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign runtime context from formation")

	t.Log("Check that the formation label value is unassigned")
	checkRuntimeContextFormationLabels(t, ctx, tenantId, rtm.ID, runtimeContext.ID, ScenariosLabel, []string{asaFormation, asaFormation2})

	t.Log("Should be able to delete formation after runtime is unassigned")
	deleteRequest = fixtures.FixDeleteFormationRequest(newFormation)
	var deleteFormation graphql.Formation
	deleteRequest.Header.Add("x-request-id", "TURSISE")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteRequest, &deleteFormation)
	assert.NoError(t, err)
	assert.Equal(t, newFormation, deleteFormation.Name)

	fixtures.DeleteFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subaccountID, tenantId)
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
	scenarioSelector := fixtures.FixLabelSelector(conf.GlobalSubaccountIDLabelKey, subaccountID)

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

	example.SaveExampleInCustomDir(t, assignReq.Query(), assignFormationCategory, "assign tenant to formation")

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

	example.SaveExampleInCustomDir(t, unassignReq.Query(), unassignFormationCategory, "unassign tenant from formation")

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
		firstFormationName  = "FIRST"
		secondFormationName = "SECOND"
	)

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

	firstFormationInputGql := graphql.FormationInput{Name: firstFormationName}
	secondFormationInputGql := graphql.FormationInput{Name: secondFormationName}

	formationTemplateName := "create-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInput(formationTemplateName)

	t.Logf("Create formation template with name: %q", formationTemplateName)
	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	in := graphql.FormationConstraintInput{
		Name:            "TestSubaccountInAtMostOneFormationOfType",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        graphql.IsNotAssignedToAnyFormationOfType,
		ResourceType:    graphql.ResourceTypeTenant,
		ResourceSubtype: "subaccount",
		InputTemplate:   "{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
	require.NotEmpty(t, constraint.ID)

	t.Logf("Attaching constraint to formation template")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	t.Logf("Should create formation: %s", firstFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, firstFormationName)
	firstFormation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, firstFormationName, &formationTemplate.Name)

	t.Logf("Should create formation: %s", secondFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, secondFormationName)
	secondFormation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, secondFormationName, &formationTemplate.Name)

	t.Logf("Assign tenant %s to formation %s", subaccountID, firstFormation.Name)
	defer fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, subaccountID, tenantId)

	t.Log("Should match expected ASA")
	asaPage := fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 1, len(asaPage.Data))

	scenarioSelector := fixtures.FixLabelSelector(conf.GlobalSubaccountIDLabelKey, subaccountID)
	assertions.AssertAutomaticScenarioAssignment(t, firstFormationName, &scenarioSelector, *asaPage.Data[0])

	t.Logf("Should fail to assign tenant %s to second formation of type %s", subaccountID, formationTemplateName)
	defer fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, subaccountID, tenantId)

	t.Logf("Detaching constraint from formation template")
	fixtures.DetachConstraintFromFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	t.Logf("Should succeed assigning tenant %s to second formation of type %s after constraint is detached", subaccountID, formationTemplateName)
	defer fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, subaccountID, tenantId)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, subaccountID, tenantId)

	t.Log("Should match expected ASAs")
	asaPage = fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 2, len(asaPage.Data))

	assertions.AssertAutomaticScenarioAssignments(t, map[string]*graphql.Label{firstFormationName: &scenarioSelector, secondFormationName: &scenarioSelector}, asaPage.Data)

	t.Logf("Unassign tenant %s from formation %s", subaccountID, firstFormation.Name)
	fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, subaccountID, tenantId)

	t.Logf("Unassign tenant %s from formation %s", subaccountID, secondFormation.Name)
	fixtures.UnassignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, subaccountID, tenantId)

	t.Log("Should match expected ASA")
	asaPage = fixtures.ListAutomaticScenarioAssignmentsWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId)
	require.Equal(t, 0, len(asaPage.Data))
}

func TestApplicationOfGivenTypeInAtMostOneFormationOfGivenType(t *testing.T) {
	ctx := context.Background()

	const (
		formationName         = "test-formation"
		applicationType       = "app-type"
		applicationNameFirst  = "e2e-tests-app-first"
		applicationNameSecond = "e2e-tests-app-second"
	)

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	formationInputGql := graphql.FormationInput{Name: formationName}

	formationTemplateName := "create-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInputWithApplicationTypes(formationTemplateName, []string{applicationType})

	t.Logf("Create formation template with name: %q", formationTemplateName)
	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	in := graphql.FormationConstraintInput{
		Name:            "SystemOfGivenTypeInAtMostOneFormationOfGivenType",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        graphql.DoesNotContainResourceOfSubtype,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: applicationType,
		InputTemplate:   "{\\\"formation_name\\\": \\\"{{.FormationName}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\"}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
	require.NotEmpty(t, constraint.ID)

	t.Logf("Attaching constraint to formation template")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	t.Logf("Should create formation: %q", formationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, formationName, &formationTemplate.Name)

	t.Logf("Create application with name %q and type %q", applicationNameFirst, applicationType)
	appFirst := graphql.ApplicationExt{} // needed so the 'defer' can be above the application creation
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &appFirst)
	appFirst, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, applicationNameFirst, conf.ApplicationTypeLabelKey, applicationType, tenantId)
	require.NoError(t, err)
	require.NotEmpty(t, appFirst.ID)

	t.Logf("Create application with name %q and type %q", applicationNameSecond, applicationType)
	appSecond := graphql.ApplicationExt{} // needed so the 'defer' can be above the application creation
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &appSecond)
	appSecond, err = fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, applicationNameSecond, conf.ApplicationTypeLabelKey, applicationType, tenantId)
	require.NoError(t, err)
	require.NotEmpty(t, appSecond.ID)

	t.Logf("Assign first application to formation with name: %q", formation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInputGql, appFirst.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInputGql, appFirst.ID, tenantId)

	t.Logf("Should fail to assign second application formation with name: %q", formation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInputGql, appSecond.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, formationInputGql, appSecond.ID, tenantId)

	t.Logf("Detaching constraint from formation template")
	fixtures.DetachConstraintFromFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	t.Logf("Should assign second application formation with name: %q", formation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInputGql, appSecond.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInputGql, appSecond.ID, tenantId)
}

func TestSystemInAtMostOneFormationOfType(t *testing.T) {
	ctx := context.Background()
	const (
		firstFormationName  = "FIRST"
		secondFormationName = "SECOND"
	)

	firstFormationInputGql := graphql.FormationInput{Name: firstFormationName}
	secondFormationInputGql := graphql.FormationInput{Name: secondFormationName}

	applicationTypes := []string{string(util.ApplicationTypeC4C), exceptionSystemType}
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	formationTemplateName := "e2e-tests-formation-template-name"
	formationTemplateInput := fixtures.FixFormationTemplateInputWithApplicationTypes(formationTemplateName, applicationTypes)

	t.Logf("Create formation template with name: %q", formationTemplateName)
	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	in := graphql.FormationConstraintInput{
		Name:            "TestSystemInAtMostOneFormationOfType",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationAssignFormation,
		Operator:        graphql.IsNotAssignedToAnyFormationOfType,
		ResourceType:    graphql.ResourceTypeApplication,
		ResourceSubtype: resourceSubtypeANY,
		InputTemplate:   fmt.Sprintf("{\\\"formation_template_id\\\": \\\"{{.FormationTemplateID}}\\\",\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"exceptSystemTypes\\\": [\\\"%s\\\"]}", exceptionSystemType),
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}
	constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
	defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
	require.NotEmpty(t, constraint.ID)

	t.Logf("Attaching constraint to formation template")
	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	t.Logf("Should create first formation: %s", firstFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, firstFormationName)
	firstFormation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, firstFormationName, &formationTemplate.Name)

	t.Logf("Should create second formation: %s", secondFormationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, secondFormationName)
	secondFormation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, secondFormationName, &formationTemplate.Name)

	t.Log("Create first application")
	app := graphql.ApplicationExt{} // needed so the 'defer' can be above the application creation
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "e2e-tests-app", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	t.Log("Create second application with exception type")
	exceptionTypeApp := graphql.ApplicationExt{} // needed so the 'defer' can be above the application creation
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &exceptionTypeApp)
	exceptionTypeApp, err = fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "e2e-tests-exceptionType-app", conf.ApplicationTypeLabelKey, exceptionSystemType, tenantId)
	require.NoError(t, err)
	require.NotEmpty(t, exceptionTypeApp.ID)

	t.Logf("Assign first application to first formation with name: %s", firstFormation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, app.ID, tenantId)

	t.Logf("Should fail to assign first application to another formation with name: %s of the same kind", secondFormation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, app.ID, tenantId)

	t.Logf("Assign second application to first formation with name: %s", firstFormation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, exceptionTypeApp.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, firstFormationInputGql, exceptionTypeApp.ID, tenantId)

	t.Logf("Should succeed assigning second application with exception type to another formation with name: %s of the same kind", secondFormation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, exceptionTypeApp.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, exceptionTypeApp.ID, tenantId)

	t.Logf("Detaching constraint from formation template")
	fixtures.DetachConstraintFromFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, formationTemplate.ID)

	t.Logf("Should succeed assigning first application to another formation with name: %s of the same kind", secondFormation.Name)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, secondFormationInputGql, app.ID, tenantId)
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

			deps, err := json.Marshal([]string{selfRegLabelValue})
			require.NoError(t, err)
			depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
			require.NoError(t, err)
			response, err := httpClient.Do(depConfigureReq)
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Logf("Could not close response body %s", err)
				}
			}()
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)

			subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
			apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
			defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
			subscription.CreateRuntimeSubscription(t, conf.SubscriptionConfig, httpClient, providerRuntime, subscriptionToken, apiPath, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, conf.SubscriptionProviderAppNameValue, true, conf.SubscriptionConfig.StandardFlow)

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

			deps, err := json.Marshal([]string{selfRegLabelValue})
			require.NoError(t, err)
			depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
			require.NoError(t, err)
			response, err := httpClient.Do(depConfigureReq)
			defer func() {
				if err := response.Body.Close(); err != nil {
					t.Logf("Could not close response body %s", err)
				}
			}()
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)

			subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
			apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
			defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
			subscription.CreateRuntimeSubscription(t, conf.SubscriptionConfig, httpClient, providerRuntime, subscriptionToken, apiPath, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, conf.SubscriptionProviderAppNameValue, true, conf.SubscriptionConfig.StandardFlow)

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

func TestFormationApplicationTypeWhileAssigning(t *testing.T) {
	ctx := context.TODO()

	formationName := "test-formation"
	applicationName := "test-application"
	invalidApplicationType := "Not in the template"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formationName)
	formation := fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, formationName)

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

	formationTemplateInput := fixtures.FixFormationTemplateInputWithRuntimeTypes(formationTemplateName, []string{runtimeType})
	var actualFormationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &actualFormationTemplate)
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
		RuntimeTypeDisplayName: &formationTemplateName,
		RuntimeArtifactKind:    &runtimeArtifactKind,
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
