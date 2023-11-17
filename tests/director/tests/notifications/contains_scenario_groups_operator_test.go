package notifications

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
)

func TestContainsScenarioGroupsOperator(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()

	formationTmplName := "e2e-tests-contains-scenario-groups-operator"
	generateFANotificationsApplicationType := "constraintGenerateApplicationType"
	assignApplication := "constraintAssignApplicationType"
	appName := "constraint-system-scenario-groups-operator-notifications"
	otherAppName := "constraint-system-scenario-groups-operator-assign"

	var ft graphql.FormationTemplate
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateAppOnlyFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, []string{generateFANotificationsApplicationType, assignApplication}, nil, supportReset)

	formationName := "containsScenarioGroupsFormationE2ETest"
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
	_ = fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)
	webhookType := graphql.WebhookTypeApplicationTenantMapping
	webhookMode := graphql.WebhookModeSync
	urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/with-state/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
	inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
	outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"

	applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

	scenarioGroup := "testScenarioGroup"
	scenarioGroup2 := "testScenarioGroup2"
	scenarioGroups := fmt.Sprintf(`{"key": "%s","description": "some description for key"}, {"key": "%s","description": "some description for key 2"}`, scenarioGroup, scenarioGroup2)

	t.Logf("Creating application template %q along with application %q", generateFANotificationsApplicationType, appName)
	applicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(generateFANotificationsApplicationType, graphql.ApplicationStatusConditionConnected)
	applicationTemplate, actualApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplateInput, appName)
	defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplate, &actualApp)
	t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", generateFANotificationsApplicationType, applicationTemplate.ID, appName, actualApp.ID)

	t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, actualApp.ID)
	actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, actualApp.ID)
	defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

	t.Logf("Creating application template %q along with application %q", assignApplication, otherAppName)
	otherApplicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(assignApplication, graphql.ApplicationStatusConditionConnected)
	otherApplicationTemplate, otherApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, otherApplicationTemplateInput, otherAppName)
	defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, otherApplicationTemplate, &otherApp)
	t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", assignApplication, otherApplicationTemplate.ID, otherAppName, otherApp.ID)

	t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationAssignFormation, assignApplication, formationTmplName, ft.ID)
	assignConstraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(assignApplication, graphql.TargetOperationAssignFormation, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", scenarioGroup))
	assignConstraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, assignConstraintsInput, ft.ID, ft.Name)
	defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, assignConstraint.ID, ft.ID)
	t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", assignConstraint.ID, formationTmplName, ft.ID)

	t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationGenerateFormationAssignmentNotification, generateFANotificationsApplicationType, formationTmplName, ft.ID)
	generateFANotificationsConstraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(generateFANotificationsApplicationType, graphql.TargetOperationGenerateFormationAssignmentNotification, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", scenarioGroup2))
	generateFANotificationsConstraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, generateFANotificationsConstraintsInput, ft.ID, ft.Name)
	defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, generateFANotificationsConstraint.ID, ft.ID)
	t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", generateFANotificationsConstraint.ID, formationTmplName, ft.ID)

	t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
	tokenForGenerateApp := fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, actualApp.ID, map[string]string{"scenario_groups": scenarioGroups})
	t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)

	t.Logf("Getting one time token for application with name: %q and id: %q...", otherApp.Name, otherApp.ID)
	tokenForAssignApp := fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, otherApp.ID, map[string]string{"scenario_groups": scenarioGroups})
	t.Logf("Successfully got one time token for application with name: %q and id: %q", otherApp.Name, otherApp.ID)

	t.Logf("Assign application application %q with ID %q to formation %s", appName, otherApp.ID, formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
	fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, tnt)

	t.Logf("Getting one time token for application with name: %q and id: %q...", otherApp.Name, otherApp.ID)
	invalidTokenForAssignApp := fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, otherApp.ID, map[string]string{"scenario_groups": `{"key": "someOtherGroup", "description": "someOtherDescription"}`})
	t.Logf("Successfully got one time token for application with name: %q and id: %q", otherApp.Name, otherApp.ID)

	headers := map[string][]string{
		oathkeeper.ConnectorTokenHeader: {invalidTokenForAssignApp.Token},
	}
	hydratorClient.ResolveToken(t, headers)

	t.Logf("Assign application application %q with ID %q to formation %s", appName, otherApp.ID, formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
	fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, tnt)

	headers = map[string][]string{
		oathkeeper.ConnectorTokenHeader: {tokenForAssignApp.Token},
	}
	hydratorClient.ResolveToken(t, headers)

	t.Logf("Assign application application %q with ID %q to formation %s", appName, otherApp.ID, formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, tnt)
	t.Logf("Successfully assigned application %q with ID %q to formation %s", appName, otherApp.ID, formationName)

	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

	t.Logf("Assign application to formation %s", formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)

	assertNoNotificationsAreSent(t, certSecuredHTTPClient, actualApp.ID)
	assertNoNotificationsAreSent(t, certSecuredHTTPClient, otherApp.ID)

	t.Logf("Unassign application to formation %s", formationName)
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)

	headers = map[string][]string{
		oathkeeper.ConnectorTokenHeader: {tokenForGenerateApp.Token},
	}
	hydratorClient.ResolveToken(t, headers)

	t.Logf("Assign application application %q with ID %q to formation %s", appName, actualApp.ID, formationName)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)
	t.Logf("Successfully assigned application %q with ID %q to formation %s", appName, actualApp.ID, formationName)

	body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	assertNotificationsCountForTenant(t, body, actualApp.ID, 1)
	assertNoNotificationsAreSent(t, certSecuredHTTPClient, otherApp.ID)
}
