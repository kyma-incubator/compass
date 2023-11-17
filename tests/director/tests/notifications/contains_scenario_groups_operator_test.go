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
	appTemplateName := "constraintApplicationType"
	otherAppTemplateName := "otherApplicationType"
	appName := "constraint-system-scenario-groups-operator"
	otherAppName := "other-constraint-system-scenario-groups-operator"

	var ft graphql.FormationTemplate
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateAppOnlyFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, []string{appTemplateName, otherAppTemplateName}, nil, supportReset)

	formationName := "app-to-app-formation-name"
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

	scenarioGroups := `{"key": "foo1","description": "bar1"}, {"key": "foo2","description": "bar2"}`

	t.Run("Contains Scenario Groups Constraint should not generate notifications when not satisfied", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Creating application template %q along with application %q", appTemplateName, appName)
		applicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(appTemplateName, graphql.ApplicationStatusConditionInitial)
		applicationTemplate, actualApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplateInput, appName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplate, &actualApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", appTemplateName, applicationTemplate.ID, appName, actualApp.ID)

		t.Logf("Creating application template %q along with application %q", otherAppTemplateName, otherAppName)
		otherApplicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(otherAppTemplateName, graphql.ApplicationStatusConditionInitial)
		otherApplicationTemplate, otherApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, otherApplicationTemplateInput, otherAppName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, otherApplicationTemplate, &otherApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", otherAppTemplateName, otherApplicationTemplate.ID, otherAppName, otherApp.ID)

		t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationAssignFormation, appTemplateName, formationTmplName, ft.ID)
		constraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(appTemplateName, graphql.TargetOperationGenerateFormationAssignmentNotification, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "nonexistent"))
		constraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, constraintsInput, ft.ID, ft.Name)
		defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", constraint.ID, formationTmplName, ft.ID)

		t.Logf("Assign application to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, tnt)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, actualApp.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, actualApp.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		t.Logf("Assign application to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)
		t.Logf("Successfully assigned application %q with ID %q to formation %s", appName, actualApp.ID, formationName)

		assertNoNotificationsAreSent(t, certSecuredHTTPClient, actualApp.ID)
		assertNoNotificationsAreSent(t, certSecuredHTTPClient, otherApp.ID)
	})
	t.Run("Contains Scenario Groups Constraint should not allow assign when not satisfied", func(t *testing.T) {
		t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationAssignFormation, appTemplateName, formationTmplName, ft.ID)
		constraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(appTemplateName, graphql.TargetOperationAssignFormation, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo1"))
		constraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, constraintsInput, ft.ID, ft.Name)
		defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", constraint.ID, formationTmplName, ft.ID)

		t.Logf("Creating application template %q along with application %q", appTemplateName, appName)
		applicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(appTemplateName, graphql.ApplicationStatusConditionInitial)
		applicationTemplate, actualApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplateInput, appName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplate, &actualApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", appTemplateName, applicationTemplate.ID, appName, actualApp.ID)

		t.Logf("Assign application application %q with ID %q to formation %s", appName, actualApp.ID, formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)
	})
	t.Run("Contains Scenario Groups Constraint should not allow assign when there is valid OTT but with wrong scenario groups and valid OTT is not used", func(t *testing.T) {
		t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationAssignFormation, appTemplateName, formationTmplName, ft.ID)
		constraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(appTemplateName, graphql.TargetOperationAssignFormation, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo1"))
		constraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, constraintsInput, ft.ID, ft.Name)
		defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", constraint.ID, formationTmplName, ft.ID)

		t.Logf("Creating application template %q along with application %q", appTemplateName, appName)
		applicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(appTemplateName, graphql.ApplicationStatusConditionConnected)
		applicationTemplate, actualApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplateInput, appName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplate, &actualApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", appTemplateName, applicationTemplate.ID, appName, actualApp.ID)

		t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
		t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)
		fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, actualApp.ID, map[string]string{"scenario_groups": scenarioGroups})
		t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
		token := fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, actualApp.ID, map[string]string{"scenario_groups": `{"key": "someOtherGroup","description": "someOtherGroupDesc"}`})
		headers := map[string][]string{
			oathkeeper.ConnectorTokenHeader: {token.Token},
		}
		hydratorClient.ResolveToken(t, headers)
		t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)

		t.Logf("Assign application application %q with ID %q to formation %s", appName, actualApp.ID, formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)
	})
	t.Run("Contains Scenario Groups Constraint should allow assign when satisfied", func(t *testing.T) {
		t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationAssignFormation, appTemplateName, formationTmplName, ft.ID)
		constraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(appTemplateName, graphql.TargetOperationAssignFormation, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo2"))
		constraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, constraintsInput, ft.ID, ft.Name)
		defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", constraint.ID, formationTmplName, ft.ID)

		t.Logf("Creating application template %q along with application %q", appTemplateName, appName)
		applicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(appTemplateName, graphql.ApplicationStatusConditionConnected)
		applicationTemplate, actualApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplateInput, appName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplate, &actualApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", appTemplateName, applicationTemplate.ID, appName, actualApp.ID)

		t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
		token := fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, actualApp.ID, map[string]string{"scenario_groups": scenarioGroups})
		headers := map[string][]string{
			oathkeeper.ConnectorTokenHeader: {token.Token},
		}
		hydratorClient.ResolveToken(t, headers)
		t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)

		t.Logf("Assign application application %q with ID %q to formation %s", appName, actualApp.ID, formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)
		t.Logf("Successfully assigned application %q with ID %q to formation %s", appName, actualApp.ID, formationName)
	})
	t.Run("Contains Scenario Groups Constraint should generate notifications when satisfied", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Creating constraint with target operation %q for application type %q and attach it to formation template %q with ID %q", graphql.TargetOperationGenerateFormationAssignmentNotification, appTemplateName, formationTmplName, ft.ID)
		constraintsInput := fixtures.FixFormationConstraintInputContainsScenarioGroups(appTemplateName, graphql.TargetOperationGenerateFormationAssignmentNotification, fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo2"))
		constraint := fixtures.CreateFormationConstraintAndAttach(t, ctx, certSecuredGraphQLClient, constraintsInput, ft.ID, ft.Name)
		defer fixtures.CleanupFormationConstraintAndDetach(t, ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		t.Logf("Successfully created and attached constraint with ID %q to formation template %q with ID %q", constraint.ID, formationTmplName, ft.ID)

		t.Logf("Creating application template %q along with application %q", appTemplateName, appName)
		applicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(appTemplateName, graphql.ApplicationStatusConditionConnected)
		applicationTemplate, actualApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplateInput, appName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, applicationTemplate, &actualApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", appTemplateName, applicationTemplate.ID, appName, actualApp.ID)

		t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
		token := fixtures.GenerateOneTimeTokenForApplicationWithCustomHeaders(t, ctx, certSecuredGraphQLClient, tnt, actualApp.ID, map[string]string{"scenario_groups": scenarioGroups})
		headers := map[string][]string{
			oathkeeper.ConnectorTokenHeader: {token.Token},
		}
		hydratorClient.ResolveToken(t, headers)
		t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)

		t.Logf("Creating application template %q along with application %q", otherAppTemplateName, otherAppName)
		otherApplicationTemplateInput := fixtures.FixApplicationTemplateWithStatusAndType(otherAppTemplateName, graphql.ApplicationStatusConditionInitial)
		otherApplicationTemplate, otherApp := fixtures.CreateApplicationTemplateFromInputWithApplication(t, ctx, certSecuredGraphQLClient, tnt, otherApplicationTemplateInput, otherAppName)
		defer fixtures.CleanupApplicationTemplateWithApplication(t, ctx, certSecuredGraphQLClient, tnt, otherApplicationTemplate, &otherApp)
		t.Logf("Successfully created application template %q with ID %q along with application %q with ID %q", otherAppTemplateName, otherApplicationTemplate.ID, otherAppName, otherApp.ID)

		t.Logf("Assign application application %q with ID %q to formation %s", otherAppName, otherApp.ID, formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, tnt)
		t.Logf("Successfully assigned application %q with ID %q to formation %s", otherAppName, otherApp.ID, formationName)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, actualApp.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, actualApp.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)
		t.Logf("Successfully added webhook to application with ID %q", actualApp.ID)

		t.Logf("Assign application application %q with ID %q to formation %s", appName, actualApp.ID, formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, tnt)
		t.Logf("Successfully assigned application %q with ID %q to formation %s", appName, actualApp.ID, formationName)

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, actualApp.ID, 1)
		assertNoNotificationsAreSent(t, certSecuredHTTPClient, otherApp.ID)
	})
}
