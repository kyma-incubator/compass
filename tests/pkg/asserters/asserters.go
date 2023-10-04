package asserters

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	unassignOperation = "unassign"
	assignOperation   = "assign"
)

type Asserter interface {
	AssertExpectations(t *testing.T, ctx context.Context)
}

type FormationAssignmentsAsserter struct {
	expectations             map[string]map[string]fixtures.AssignmentState
	expectedAssignmentsCount int
	certSecuredGraphQLClient *graphql.Client
	tenantID                 string
	formationID              string
	delay                    int
}

func NewFormationAssignmentAsserter(expectations map[string]map[string]fixtures.AssignmentState, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string, formationID string) *FormationAssignmentsAsserter {
	return &FormationAssignmentsAsserter{
		expectations:             expectations,
		expectedAssignmentsCount: expectedAssignmentsCount,
		certSecuredGraphQLClient: certSecuredGraphQLClient,
		tenantID:                 tenantID,
		formationID:              formationID,
	}
}

func (a *FormationAssignmentsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	a.assertFormationAssignments(t, ctx, a.certSecuredGraphQLClient, a.tenantID, a.formationID, a.expectedAssignmentsCount, a.expectations)
}

func (a *FormationAssignmentsAsserter) assertFormationAssignments(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
	spew.Dump("Started Asserting assignments")
	spew.Dump("Assignment dumped: ", assignments)
	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q", assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q and target %q", assignment.Source, assignment.Target)

		require.Equal(t, assignmentExpectation.State, assignment.State)
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Config), str.PtrStrToStr(assignment.Configuration))
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Value), str.PtrStrToStr(assignment.Value))
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error))
	}
	spew.Dump("Finished Asserting assignments")
}

type FormationAssignmentsAsyncAsserter struct {
	FormationAssignmentsAsserter
	delay int64
}

func NewFormationAssignmentAsyncAsserter(expectations map[string]map[string]fixtures.AssignmentState, expectedAssignmentsCount int, certSecuredGraphQLClient *graphql.Client, tenantID string, formationID string, delay int64) *FormationAssignmentsAsyncAsserter {
	f := FormationAssignmentsAsyncAsserter{}
	f.expectations = expectations
	f.expectedAssignmentsCount = expectedAssignmentsCount
	f.certSecuredGraphQLClient = certSecuredGraphQLClient
	f.tenantID = tenantID
	f.formationID = formationID
	f.delay = delay
	return &f
}

func (a *FormationAssignmentsAsyncAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	a.assertFormationAssignmentsAsynchronously(t, ctx, a.certSecuredGraphQLClient, a.tenantID, a.formationID, a.expectedAssignmentsCount, a.expectations)
}

func (a *FormationAssignmentsAsyncAsserter) assertFormationAssignmentsAsynchronously(t *testing.T, ctx context.Context, certSecuredGraphQLClient *graphql.Client, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.AssignmentState) {
	t.Logf("Sleeping for %d seconds while the async formation assignment status is proccessed...", a.delay)
	time.Sleep(time.Second * time.Duration(a.delay))
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)

	assignments := assignmentsPage.Data

	spew.Dump("ASSIGNMENTS ::: ", assignments)
	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q and source %q", assignment.ID, assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with ID: %q, source %q and target %q", assignment.ID, assignment.Source, assignment.Target)
		require.Equal(t, assignmentExpectation.State, assignment.State, "Assignment with ID: %q has different state than expected", assignment.ID)

		require.Equal(t, str.PtrStrToStr(assignmentExpectation.Error), str.PtrStrToStr(assignment.Error))

		expectedAssignmentConfigStr := str.PtrStrToStr(assignmentExpectation.Config)
		actualAssignmentConfigStr := str.PtrStrToStr(assignment.Configuration)
		if expectedAssignmentConfigStr != "" && expectedAssignmentConfigStr != "\"\"" && actualAssignmentConfigStr != "" && actualAssignmentConfigStr != "\"\"" {
			require.JSONEq(t, expectedAssignmentConfigStr, actualAssignmentConfigStr)
			require.JSONEq(t, str.PtrStrToStr(assignmentExpectation.Config), actualAssignmentConfigStr)
		} else {
			require.Equal(t, expectedAssignmentConfigStr, actualAssignmentConfigStr)
		}
	}
}

type NotificationsAsserter struct {
	expectedNotificationsCount         int
	op                                 string
	formationID                        string
	targetObjectID                     string
	sourceObjectID                     string
	localTenantID                      string
	appNamespace                       string
	region                             string
	tenant                             string
	tenantParentCustomer               string
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewNotificationsAsserter(expectedNotificationsCount int, op string, formationID string, targetObjectID, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, externalServicesMockMtlsSecuredURL string, client *http.Client) *NotificationsAsserter {
	return &NotificationsAsserter{expectedNotificationsCount: expectedNotificationsCount, op: op, formationID: formationID, targetObjectID: targetObjectID, sourceObjectID: sourceObjectID, localTenantID: localTenantID, appNamespace: appNamespace, region: region, tenant: tenant, tenantParentCustomer: tenantParentCustomer, externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL, client: client}
}

func (a *NotificationsAsserter) AssertExpectations(t *testing.T, _ context.Context) {
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
	assertNotificationsCount(t, body, a.targetObjectID, a.expectedNotificationsCount)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	assignNotificationAboutSource := notificationsForTarget.Array()[0]
	assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, assignNotificationAboutSource, assignOperation, a.formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.tenant, a.tenantParentCustomer, nil)
}

func getNotificationsFromExternalSvcMock(t *testing.T, client *http.Client, ExternalServicesMockMtlsSecuredURL string) []byte {
	t.Logf("Getting formation notifications recieved in external services mock")
	resp, err := client.Get(ExternalServicesMockMtlsSecuredURL + "/formation-callback")
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

func assertNotificationsCount(t *testing.T, body []byte, objectID string, count int) {
	fmt.Println("WHOLE BODY: ", string(body))
	notifications := gjson.GetBytes(body, objectID)
	if count > 0 {
		require.True(t, notifications.Exists())
		require.Len(t, notifications.Array(), count)
	} else {
		require.False(t, notifications.Exists())
	}
}

func assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t *testing.T, notification gjson.Result, op, formationID, expectedAppID, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string, expectedConfig *string) {
	require.Equal(t, op, notification.Get("Operation").String())
	if op == unassignOperation {
		require.Equal(t, expectedAppID, notification.Get("ApplicationID").String())
	}
	require.Equal(t, formationID, notification.Get("RequestBody.ucl-formation-id").String())
	require.Equal(t, expectedTenant, notification.Get("RequestBody.globalAccountId").String())
	require.Equal(t, expectedCustomerID, notification.Get("RequestBody.crmId").String())

	notificationItems := notification.Get("RequestBody.items")
	require.True(t, notificationItems.Exists())
	require.Len(t, notificationItems.Array(), 1)

	app1FromNotification := notificationItems.Array()[0]
	require.Equal(t, expectedAppID, app1FromNotification.Get("ucl-system-tenant-id").String())
	require.Equal(t, expectedLocalTenantID, app1FromNotification.Get("tenant-id").String())
	require.Equal(t, expectedAppNamespace, app1FromNotification.Get("application-namespace").String())
	require.Equal(t, expectedAppRegion, app1FromNotification.Get("region").String())
	if expectedConfig != nil {
		require.Equal(t, *expectedConfig, notification.Get("RequestBody.config").String())
	}
}

type UnassignNotificationsAsserter struct {
	op                                 string
	expectedNotificationsCountForOp    int
	formationID                        string
	targetObjectID                     string
	sourceObjectID                     string
	localTenantID                      string
	appNamespace                       string
	region                             string
	tenant                             string
	tenantParentCustomer               string
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewUnassignNotificationsAsserter(op string, expectedNotificationsCountForOp int, formationID string, targetObjectID string, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, externalServicesMockMtlsSecuredURL string, client *http.Client) *UnassignNotificationsAsserter {
	return &UnassignNotificationsAsserter{op: op, expectedNotificationsCountForOp: expectedNotificationsCountForOp, formationID: formationID, targetObjectID: targetObjectID, sourceObjectID: sourceObjectID, localTenantID: localTenantID, appNamespace: appNamespace, region: region, tenant: tenant, tenantParentCustomer: tenantParentCustomer, externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL, client: client}
}

func (a *UnassignNotificationsAsserter) AssertExpectations(t *testing.T, _ context.Context) {
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	assignNotificationAboutSource := notificationsForTarget.Array()[0]
	notificationsFoundCount := 0
	for _, notification := range notificationsForTarget.Array() {
		op := notification.Get("Operation").String()
		if op == a.op {
			notificationsFoundCount++
			assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, assignNotificationAboutSource, assignOperation, a.formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.tenant, a.tenantParentCustomer, nil)
		}
	}
	require.Equal(t, a.expectedNotificationsCountForOp, notificationsFoundCount, "two notifications for unassign app2 expected")
}

type FormationStatusAsserter struct {
	formationID              string
	tenant                   string
	certSecuredGraphQLClient *graphql.Client
	condition                gql.FormationStatusCondition
	errors                   []*gql.FormationStatusError
}

func NewFormationStatusAsserter(formationID string, tenant string, certSecuredGraphQLClient *graphql.Client) *FormationStatusAsserter {
	return &FormationStatusAsserter{formationID: formationID, tenant: tenant, certSecuredGraphQLClient: certSecuredGraphQLClient, condition: gql.FormationStatusConditionReady, errors: nil}
}

func (a *FormationStatusAsserter) WithCondition(condition gql.FormationStatusCondition) *FormationStatusAsserter {
	a.condition = condition
	return a
}

func (a *FormationStatusAsserter) WithErrors(errors []*gql.FormationStatusError) *FormationStatusAsserter {
	a.errors = errors
	return a
}

func (a *FormationStatusAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	a.assertFormationStatus(t, ctx, a.tenant, a.formationID, gql.FormationStatus{
		Condition: a.condition,
		Errors:    a.errors,
	})
}

func (a *FormationStatusAsserter) assertFormationStatus(t *testing.T, ctx context.Context, tenant, formationID string, expectedFormationStatus gql.FormationStatus) {
	// Get the formation with its status
	t.Logf("Getting formation with ID: %q", formationID)
	var gotFormation gql.FormationExt
	getFormationReq := fixtures.FixGetFormationRequest(formationID)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, a.certSecuredGraphQLClient, tenant, getFormationReq, &gotFormation)
	require.NoError(t, err)

	// Assert the status
	require.Equal(t, expectedFormationStatus.Condition, gotFormation.Status.Condition, "Formation with ID %q is with status %q, but %q was expected", formationID, gotFormation.Status.Condition, expectedFormationStatus.Condition)

	if expectedFormationStatus.Errors == nil {
		require.Nil(t, gotFormation.Status.Errors)
	} else { // assert only the Message and ErrorCode
		require.Len(t, gotFormation.Status.Errors, len(expectedFormationStatus.Errors))
		for _, expectedError := range expectedFormationStatus.Errors {
			found := false
			for _, gotError := range gotFormation.Status.Errors {
				if gotError.ErrorCode == expectedError.ErrorCode && gotError.Message == expectedError.Message {
					found = true
					break
				}
			}
			require.Truef(t, found, "Error %q with error code %d was not found", expectedError.Message, expectedError.ErrorCode)
		}
	}
}

type NotificationsCountAsserter struct {
	expectedNotificationsCount         int
	op                                 string
	targetObjectID                     string
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
}

func NewNotificationsCountAsserter(expectedNotificationsCount int, op string, targetObjectID, externalServicesMockMtlsSecuredURL string, client *http.Client) *NotificationsCountAsserter {
	return &NotificationsCountAsserter{expectedNotificationsCount: expectedNotificationsCount, op: op, targetObjectID: targetObjectID, externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL, client: client}
}

func (a *NotificationsCountAsserter) AssertExpectations(t *testing.T, _ context.Context) {
	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	notificationsAboutSource := notificationsForTarget.Array()
	assertAtLeastNNotificationsOfTypeReceived(t, notificationsAboutSource, a.op, a.expectedNotificationsCount)
}

func assertAtLeastNNotificationsOfTypeReceived(t *testing.T, notifications []gjson.Result, op string, minCount int) {
	notificationsForOperationCount := 0
	for _, notification := range notifications {
		if notification.Get("Operation").String() == op {
			notificationsForOperationCount++
		}
	}
	require.LessOrEqual(t, minCount, notificationsForOperationCount)
}
