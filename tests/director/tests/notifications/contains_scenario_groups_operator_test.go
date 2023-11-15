package notifications

import (
	"context"
	"fmt"
	"testing"

	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainsScenarioGroupsOperator(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()

	formationTmplName := "e2e-tests-contains-scenario-groups-operator"
	appTemplateName := fixtures.CreateAppTemplateName("S/4HANA Cloud")
	otherAppTemplateName := fixtures.CreateAppTemplateName("Other Application Type")

	var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateAppOnlyFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, []string{appTemplateName, otherAppTemplateName}, nil, supportReset)

	formationName := "app-to-app-formation-name"
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
	_ = fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	t.Run("Contains Scenario Groups Constraint should not generate notifications when not satisfied", func(t *testing.T) {

		actualApp, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "test-app", conf.ApplicationTypeLabelKey, appTemplateName, tnt)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &actualApp)
		require.NoError(t, err)

		otherApp, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "otherSystem", conf.ApplicationTypeLabelKey, otherAppTemplateName, tnt)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &otherApp)
		require.NoError(t, err)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		in := graphql.FormationConstraintInput{
			Name:            "TestContainsScenarioGroupsNotifications",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationGenerateFormationAssignmentNotification,
			Operator:        formationconstraintpkg.ContainsScenarioGroups,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: appTemplateName,
			InputTemplate:   fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"formation_template_id\\\":\\\"{{.FormationTemplateID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "nonexistent"),
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}
		constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
		require.NotEmpty(t, constraint.ID)

		defer fixtures.DetachConstraintFromFormationTemplateNoCheckError(ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, constraint.Name, ft.ID, ft.Name)

		t.Logf("Assign application to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
		var assignedFormation graphql.Formation
		assignReq := fixtures.FixAssignFormationRequest(otherApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/with-state/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, actualApp.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, actualApp.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		t.Logf("Assign application to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		assertNoNotificationsAreSent(t, certSecuredHTTPClient, actualApp.ID)
		assertNoNotificationsAreSent(t, certSecuredHTTPClient, otherApp.ID)
	})
	t.Run("Contains Scenario Groups Constraint should not allow assign when not satisfied", func(t *testing.T) {
		constraintsIn := graphql.FormationConstraintInput{
			Name:            "TestContainsScenarioGroupsAssign",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationAssignFormation,
			Operator:        formationconstraintpkg.ContainsScenarioGroups,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: appTemplateName,
			InputTemplate:   fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"formation_template_id\\\":\\\"{{.FormationTemplateID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo1"),
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}
		constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintsIn)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
		require.NotEmpty(t, constraint.ID)

		defer fixtures.DetachConstraintFromFormationTemplateNoCheckError(ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, constraint.Name, ft.ID, ft.Name)

		actualApp, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "test-app", conf.ApplicationTypeLabelKey, appTemplateName, tnt)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		t.Logf("Assign application to formation %s", formationName)
		var assignedFormation graphql.Formation
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.Error(t, err)
	})
	t.Run("Contains Scenario Groups Constraint should allow assign when satisfied", func(t *testing.T) {
		constraintsIn := graphql.FormationConstraintInput{
			Name:            "TestContainsScenarioGroupsAssign",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationAssignFormation,
			Operator:        formationconstraintpkg.ContainsScenarioGroups,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: appTemplateName,
			InputTemplate:   fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"formation_template_id\\\":\\\"{{.FormationTemplateID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo2"),
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}
		constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintsIn)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
		require.NotEmpty(t, constraint.ID)

		defer fixtures.DetachConstraintFromFormationTemplateNoCheckError(ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, constraint.Name, ft.ID, ft.Name)

		statusCond := graphql.ApplicationStatusConditionConnected
		in := graphql.ApplicationRegisterInput{
			Name: "test-s4-system",
			Labels: graphql.Labels{
				conf.ApplicationTypeLabelKey: appTemplateName,
			},
			StatusCondition: &statusCond,
		}

		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)

		request := fixtures.FixRegisterApplicationRequest(appInputGQL)

		actualApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &actualApp)

		t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
		tokenRequest := fixtures.FixRequestOneTimeTokenForApplication(actualApp.ID)
		token := graphql.OneTimeTokenForApplicationExt{}
		scenarioGroups := `{"key": "foo1","description": "bar1"}, {"key": "foo2","description": "bar2"}`
		tokenRequest.Header.Add("scenario_groups", scenarioGroups)
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, tokenRequest, &token)

		require.NoError(t, err)
		require.NotEmpty(t, token.Token)
		require.NotEmpty(t, token.ConnectorURL)
		require.NotEmpty(t, token.LegacyConnectorURL)
		t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)

		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		assertions.AssertApplication(t, in, actualApp)
		assert.Equal(t, statusCond, actualApp.Status.Condition)

		t.Logf("Assign application to formation %s", formationName)
		var assignedFormation graphql.Formation
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)
	})
	t.Run("Contains Scenario Groups Constraint should generate notifications when satisfied", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		constraintsIn := graphql.FormationConstraintInput{
			Name:            "TestContainsScenarioGroupsAssign",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationGenerateFormationAssignmentNotification,
			Operator:        formationconstraintpkg.ContainsScenarioGroups,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: appTemplateName,
			InputTemplate:   fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"formation_template_id\\\":\\\"{{.FormationTemplateID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", "foo2"),
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}
		constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, constraintsIn)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
		require.NotEmpty(t, constraint.ID)

		defer fixtures.DetachConstraintFromFormationTemplateNoCheckError(ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, constraint.Name, ft.ID, ft.Name)

		statusCond := graphql.ApplicationStatusConditionConnected
		in := graphql.ApplicationRegisterInput{
			Name: "test-s4-system",
			Labels: graphql.Labels{
				conf.ApplicationTypeLabelKey: appTemplateName,
			},
			StatusCondition: &statusCond,
		}

		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)

		request := fixtures.FixRegisterApplicationRequest(appInputGQL)

		actualApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &actualApp)

		t.Logf("Getting one time token for application with name: %q and id: %q...", actualApp.Name, actualApp.ID)
		tokenRequest := fixtures.FixRequestOneTimeTokenForApplication(actualApp.ID)
		token := graphql.OneTimeTokenForApplicationExt{}
		scenarioGroups := `{"key": "foo1","description": "bar1"}, {"key": "foo2","description": "bar2"}`
		tokenRequest.Header.Add("scenario_groups", scenarioGroups)
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, tokenRequest, &token)

		require.NoError(t, err)
		require.NotEmpty(t, token.Token)
		require.NotEmpty(t, token.ConnectorURL)
		require.NotEmpty(t, token.LegacyConnectorURL)
		t.Logf("Successfully got one time token for application with name: %q and id: %q", actualApp.Name, actualApp.ID)

		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		assertions.AssertApplication(t, in, actualApp)
		assert.Equal(t, statusCond, actualApp.Status.Condition)

		otherApp, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "otherSystem", conf.ApplicationTypeLabelKey, otherAppTemplateName, tnt)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &otherApp)
		require.NoError(t, err)

		t.Logf("Assign application to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, otherApp.ID, graphql.FormationObjectTypeApplication, tnt)
		var assignedFormation graphql.Formation
		assignReq := fixtures.FixAssignFormationRequest(otherApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/with-state/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, actualApp.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, actualApp.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		t.Logf("Assign application to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, actualApp.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(actualApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, actualApp.ID, 1)
		assertNoNotificationsAreSent(t, certSecuredHTTPClient, otherApp.ID)
	})
}
