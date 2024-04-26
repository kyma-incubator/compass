package notifications

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	mock_data "github.com/kyma-incubator/compass/tests/pkg/notifications/expectations-builders"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/operations"
	resource_providers "github.com/kyma-incubator/compass/tests/pkg/notifications/resource-providers"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
)

func TestFormationNotificationsForDraftFormationWithInitialConfig(t *testing.T) {
	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

	ctx := context.Background()

	tnt := conf.TestConsumerAccountID

	appName := "testAsyncApp"
	appType := "async-app-type-1"
	appProvider1 := resource_providers.NewApplicationProvider(appType, conf.ApplicationTypeLabelKey, appName, tnt)
	defer appProvider1.Cleanup(t, ctx, certSecuredGraphQLClient)
	app1ID := appProvider1.Provide(t, ctx, certSecuredGraphQLClient)
	t.Logf("Created application 1 with name %q and ID %q", appName, app1ID)

	appName2 := "testAsyncApp2"
	appType2 := "async-app-type-2"
	appProvider2 := resource_providers.NewApplicationProvider(appType2, conf.ApplicationTypeLabelKey, appName2, tnt)
	defer appProvider2.Cleanup(t, ctx, certSecuredGraphQLClient)
	app2ID := appProvider2.Provide(t, ctx, certSecuredGraphQLClient)
	t.Logf("Created application 2 with name %q and ID %q", appName2, app2ID)

	t.Logf("Add webhook with type %q and mode: %q to application with type %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, appType)
	op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app2ID, tnt).
		WithWebhookMode(graphql.WebhookModeSync).
		WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
		WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
		WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}").Operation()
	defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
	op.Execute(t, ctx, certSecuredGraphQLClient)

	formationTmplName := "formation-template-name"
	ftProvider := resource_providers.NewFormationTemplateCreator(formationTmplName)
	defer ftProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	ftplID := ftProvider.WithSupportedResources(appProvider1.GetResource(), appProvider2.GetResource()).
		Provide(t, ctx, certSecuredGraphQLClient)
	t.Logf("Created Formation Template with ID: %q and name: %q", ftplID, formationTmplName)

	draftFormationName := "draft-formation-name"
	t.Logf("Creating formation with name: %q from template with name: %q", draftFormationName, formationTmplName)
	formationProvider := resource_providers.NewFormationProvider(draftFormationName, tnt, &formationTmplName).
		WithState(draftFormationState)
	defer formationProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	formationID := formationProvider.Provide(t, ctx, certSecuredGraphQLClient)

	// Assign both applications when the formation is still in DRAFT state and validate no notifications are sent and formation assignments are in INITIAL state
	t.Logf("Assign application 1 to formation: %s", draftFormationName)
	assignApp1 := operations.NewAssignAppToFormationOperation(app1ID, tnt).
		WithFormationName(draftFormationName).
		Operation()
	defer assignApp1.Cleanup(t, ctx, certSecuredGraphQLClient)
	assignApp1.Execute(t, ctx, certSecuredGraphQLClient)

	t.Logf("Assign application 2 to formation: %s", draftFormationName)
	assignApp2 := operations.NewAssignAppToFormationOperation(app2ID, tnt).
		WithFormationName(draftFormationName).
		Operation()
	defer assignApp2.Cleanup(t, ctx, certSecuredGraphQLClient)
	assignApp2.Execute(t, ctx, certSecuredGraphQLClient)

	t.Logf("Listing formation assignments for formation with ID: %q", formationID)
	listFormationAssignmentsReq := fixtures.FixListFormationAssignmentRequest(formationID, 100)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tnt, listFormationAssignmentsReq)
	require.Len(t, assignmentsPage.Data, 4)
	require.Equal(t, 4, assignmentsPage.TotalCount)
	formationAssignmentID := getFormationAssignmentIDByTargetTypeAndSourceID(t, assignmentsPage, graphql.FormationAssignmentTypeApplication, app1ID)
	t.Logf("Successfully listed FAs for formation ID: %q", formationID)

	accountTokenURL, err := token.ChangeSubdomain(conf.UsernameAuthCfg.Account.TokenURL, conf.UsernameAuthCfg.Account.Subdomain, conf.UsernameAuthCfg.Account.OAuthTokenPath)
	require.NoError(t, err)
	require.NotEmpty(t, accountTokenURL)

	// The accountToken is JWT token containing claim with account ID for tenant. In local setup that's 'ApplicationsForRuntimeTenantName'
	accountToken := token.GetUserToken(t, ctx, accountTokenURL, conf.UsernameAuthCfg.Account.ClientID, conf.UsernameAuthCfg.Account.ClientSecret, conf.BasicUsername, conf.BasicPassword, claims.AccountAuthenticatorClaimKey)

	testConfig := `{"test":"something"}`
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	executeFAStatusUpdateReqWithExternalToken(t, client, accountToken, testConfig, formationID, formationAssignmentID, http.StatusOK)

	expectedAssignments := map[string]map[string]fixtures.Assignment{
		app1ID: {
			app1ID: fixtures.Assignment{AssignmentStatus: fixtures.AssignmentState{State: initialAssignmentState, Config: nil, Value: nil, Error: nil}},
			app2ID: fixtures.Assignment{AssignmentStatus: fixtures.AssignmentState{State: initialAssignmentState, Config: &testConfig, Value: nil, Error: nil}},
		},
		app2ID: {
			app1ID: fixtures.Assignment{AssignmentStatus: fixtures.AssignmentState{State: initialAssignmentState, Config: nil, Value: nil, Error: nil}},
			app2ID: fixtures.Assignment{AssignmentStatus: fixtures.AssignmentState{State: initialAssignmentState, Config: nil, Value: nil, Error: nil}},
		},
	}

	asserters.NewFormationAssignmentAsyncAsserter(expectedAssignments, 4, certSecuredGraphQLClient, tnt).
		WithFormationName(draftFormationName).
		AssertExpectations(t, ctx)

	t.Logf("Cleanup notifications")
	op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
	defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
	op.Execute(t, ctx, certSecuredGraphQLClient)

	t.Logf("Finalize formation")
	expectationsBuilder := mock_data.NewFAExpectationsBuilder().
		WithParticipant(app1ID).
		WithParticipant(app2ID).
		WithNotifications([]*mock_data.NotificationData{
			mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
		})
	faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
		WithFormationName(draftFormationName)
	statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
		WithFormationName(draftFormationName)
	notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app2ID, app1ID, "", "", "", tnt, "", testConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
		WithAssertTrustDetails(true).
		WithFormationName(draftFormationName).
		WithGQLClient(certSecuredGraphQLClient)
	op = operations.NewFinalizeFormationOperation().
		WithTenantID(tnt).
		WithFormationName(draftFormationName).
		WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter)
	defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
	op.Execute(t, ctx, certSecuredGraphQLClient)

	t.Logf("Cleanup notifications")
	op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
	defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
	op.Execute(t, ctx, certSecuredGraphQLClient)
}
