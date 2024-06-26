package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	directordestinationcreator "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	esmdestinationcreator "github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	jsonutils "github.com/kyma-incubator/compass/tests/pkg/json"
	testpkg "github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	assignOperation                  = "assign"
	unassignOperation                = "unassign"
	createFormationOperation         = "createFormation"
	deleteFormationOperation         = "deleteFormation"
	emptyParentCustomerID            = "" // in the respective tests, the used GA tenant does not have customer parent, thus we assert that it is empty
	supportReset                     = true
	doesNotSupportReset              = false
	consumerType                     = "Integration System" // should be a valid consumer type
	exceptionSystemType              = "exception-type"
	eventuallyTimeoutForDestinations = 60 * time.Second
	eventuallyTickForDestinations    = 2 * time.Second
	eventuallyTimeout                = 8 * time.Second
	eventuallyTick                   = 50 * time.Millisecond
	readyAssignmentState             = "READY"
	createReadyAssignmentState       = "CREATE_READY"
	deleteReadyAssignmentState       = "DELETE_READY"
	initialAssignmentState           = "INITIAL"
	configPendingAssignmentState     = "CONFIG_PENDING"
	deletingAssignmentState          = "DELETING"
	draftFormationState              = "DRAFT"
	basicAuthType                    = "Basic"
	samlAuthType                     = "SAML2.0"
	oauth2AuthType                   = "bearer"
)

var (
	tenantAccessLevels = []string{"account", "global"} // should be a valid tenant access level
)

func assertFormationAssignments(t *testing.T, ctx context.Context, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.Assignment) {
	listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)

	for _, assignment := range assignments {
		targetAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q", assignment.Source)

		assignmentExpectation, ok := targetAssignmentsExpectations[assignment.Target]
		require.Truef(t, ok, "Could not find expectations for assignment with source %q and target %q", assignment.Source, assignment.Target)

		require.Equal(t, assignmentExpectation.AssignmentStatus.State, assignment.State)
		expectedAssignmentConfigStr := str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Config)
		assignmentConfiguration := str.PtrStrToStr(assignment.Configuration)
		if expectedAssignmentConfigStr != "" && expectedAssignmentConfigStr != "\"\"" && assignmentConfiguration != "" && assignmentConfiguration != "\"\"" {
			require.JSONEq(t, expectedAssignmentConfigStr, assignmentConfiguration)
		} else {
			require.Equal(t, expectedAssignmentConfigStr, assignmentConfiguration)
		}
		if str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Value) != "" && str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Value) != "\"\"" && str.PtrStrToStr(assignment.Value) != "" && str.PtrStrToStr(assignment.Value) != "\"\"" {
			require.JSONEq(t, str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Value), str.PtrStrToStr(assignment.Value))
		} else {
			require.Equal(t, expectedAssignmentConfigStr, assignmentConfiguration)
		}
		require.Equal(t, str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Error), str.PtrStrToStr(assignment.Error))
		// assert operations
		require.Equal(t, len(assignmentExpectation.Operations), len(assignment.AssignmentOperations.Data))
		for _, expectedOperation := range assignmentExpectation.Operations {
			require.Truef(t, testpkg.ContainsMatchingOperation(expectedOperation, assignment.AssignmentOperations.Data), "Could not find expected operation %v in assignment with ID %q", expectedOperation, assignment.ID)
		}
	}
}

func assertFormationAssignmentsAsynchronouslyWithEventually(t *testing.T, ctx context.Context, tenantID, formationID string, expectedAssignmentsCount int, expectedAssignments map[string]map[string]fixtures.Assignment, timeout, tick time.Duration) {
	t.Logf("Asserting formation assignments with eventually...")
	tOnce := testingx.NewOnceLogger(t)
	require.Eventually(t, func() (isOkay bool) {
		tOnce.Logf("Getting formation assignments...")
		listFormationAssignmentsRequest := fixtures.FixListFormationAssignmentRequest(formationID, 200)
		assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, tenantID, listFormationAssignmentsRequest)
		if expectedAssignmentsCount != assignmentsPage.TotalCount {
			tOnce.Logf("The expected assignments count: %d didn't match the actual: %d", expectedAssignmentsCount, assignmentsPage.TotalCount)
			tOnce.Logf("The actual assignments are: %s", *jsonutils.MarshalJSON(t, assignmentsPage))
			return
		}
		tOnce.Logf("There is/are: %d assignment(s), assert them with the expected ones...", assignmentsPage.TotalCount)

		assignments := assignmentsPage.Data
		for _, assignment := range assignments {
			sourceAssignmentsExpectations, ok := expectedAssignments[assignment.Source]
			if !ok {
				tOnce.Logf("Could not find expectations for assignment with ID: %q and source ID: %q", assignment.ID, assignment.Source)
				tOnce.Logf("The actual assignments are: %s", *jsonutils.MarshalJSON(t, assignmentsPage))
				return
			}
			assignmentExpectation, ok := sourceAssignmentsExpectations[assignment.Target]
			if !ok {
				tOnce.Logf("Could not find expectations for assignment with ID: %q, source ID: %q and target ID: %q", assignment.ID, assignment.Source, assignment.Target)
				tOnce.Logf("The actual assignments are: %s", *jsonutils.MarshalJSON(t, assignmentsPage))
				return
			}
			if assignmentExpectation.AssignmentStatus.State != assignment.State {
				tOnce.Logf("The expected assignment state: %s doesn't match the actual: %s for assignment ID: %s", assignmentExpectation.AssignmentStatus.State, assignment.State, assignment.ID)
				tOnce.Logf("The actual assignments are: %s", *jsonutils.MarshalJSON(t, assignmentsPage))
				return
			}
			if isEqual := jsonutils.AssertJSONStringEquality(tOnce, assignmentExpectation.AssignmentStatus.Error, assignment.Error); !isEqual {
				tOnce.Logf("The expected assignment state: %s doesn't match the actual: %s for assignment ID: %s", str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Error), str.PtrStrToStr(assignment.Error), assignment.ID)
				tOnce.Logf("The actual assignments are: %s", *jsonutils.MarshalJSON(t, assignmentsPage))
				return
			}
			if isEqual := jsonutils.AssertJSONStringEquality(tOnce, assignmentExpectation.AssignmentStatus.Config, assignment.Configuration); !isEqual {
				tOnce.Logf("The expected assignment config: %s doesn't match the actual: %s for assignment ID: %s", str.PtrStrToStr(assignmentExpectation.AssignmentStatus.Config), str.PtrStrToStr(assignment.Configuration), assignment.ID)
				tOnce.Logf("The actual assignments are: %s", *jsonutils.MarshalJSON(t, assignmentsPage))
				return
			}
			if len(assignmentExpectation.Operations) != len(assignment.AssignmentOperations.Data) {
				tOnce.Logf("The expected number of operations: %d doesn't match the actual number: %d", len(assignmentExpectation.Operations), len(assignment.AssignmentOperations.Data))
				return
			}
			for _, expectedOperation := range assignmentExpectation.Operations {
				if !testpkg.ContainsMatchingOperation(expectedOperation, assignment.AssignmentOperations.Data) {
					tOnce.Logf("Could not find expected operation %v in assignment with ID %q", expectedOperation, assignment.ID)
					return
				}
			}
		}

		tOnce.Logf("Successfully asserted formation asssignments asynchronously")
		return true
	}, timeout, tick)
}

func assertFormationStatus(t *testing.T, ctx context.Context, tenant, formationID string, expectedFormationStatus graphql.FormationStatus) {
	// Get the formation with its status
	t.Logf("Getting formation with ID: %q", formationID)
	var gotFormation graphql.FormationExt
	getFormationReq := fixtures.FixGetFormationRequest(formationID)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant, getFormationReq, &gotFormation)
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

func attachDestinationCreatorConstraints(t *testing.T, ctx context.Context, formationTemplate graphql.FormationTemplate, statusReturnedConstraintResourceType, sendNotificationConstraintResourceType graphql.ResourceType) []func() {
	deferredFunctions := make([]func(), 0, 4)
	firstConstraintInput := graphql.FormationConstraintInput{
		Name:            "e2e-destination-creator-notification-status-returned",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationNotificationStatusReturned,
		Operator:        formationconstraintpkg.DestinationCreator,
		ResourceType:    statusReturnedConstraintResourceType,
		ResourceSubtype: "ANY",
		InputTemplate:   "{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .NotificationStatusReport }}\\\"notification_status_report_memory_address\\\":{{ .NotificationStatusReport.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	firstConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraintInput)
	deferredFunctions = append([]func(){func() {
		fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraint.ID)
	}}, deferredFunctions...)
	require.NotEmpty(t, firstConstraint.ID)

	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, firstConstraint.ID, firstConstraint.Name, formationTemplate.ID, formationTemplate.Name)
	deferredFunctions = append([]func(){func() {
		fixtures.DetachConstraintFromFormationTemplate(t, ctx, certSecuredGraphQLClient, firstConstraint.ID, formationTemplate.ID)
	}}, deferredFunctions...)

	// second constraint
	secondConstraintInput := graphql.FormationConstraintInput{
		Name:            "e2e-destination-creator-send-notification",
		ConstraintType:  graphql.ConstraintTypePre,
		TargetOperation: graphql.TargetOperationSendNotification,
		Operator:        formationconstraintpkg.DestinationCreator,
		ResourceType:    sendNotificationConstraintResourceType,
		ResourceSubtype: "ANY",
		InputTemplate:   "{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
		ConstraintScope: graphql.ConstraintScopeFormationType,
	}

	secondConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraintInput)
	deferredFunctions = append([]func(){func() {
		fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraint.ID)
	}}, deferredFunctions...)
	require.NotEmpty(t, secondConstraint.ID)

	fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, secondConstraint.ID, secondConstraint.Name, formationTemplate.ID, formationTemplate.Name)
	deferredFunctions = append([]func(){func() {
		fixtures.DetachConstraintFromFormationTemplate(t, ctx, certSecuredGraphQLClient, secondConstraint.ID, formationTemplate.ID)
	}}, deferredFunctions...)
	return deferredFunctions
}

func cleanupNotificationsFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/formation-callback/cleanup", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func cleanupDestinationsFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/destinations/cleanup", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func cleanupDestnationCertificatesFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/destination-certificates/cleanup", nil)
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

func assertNoDestinationIsFound(t *testing.T, client *clients.DestinationClient, serviceURL, destinationName, token string) {
	_ = client.FindDestinationByName(t, serviceURL, destinationName, token, "", http.StatusNotFound)
}

func assertNoDestinationCertificateIsFound(t *testing.T, client *clients.DestinationClient, serviceURL, certificateName, instanceID, token string) {
	_ = client.GetDestinationCertificateByName(t, serviceURL, certificateName, instanceID, token, http.StatusNotFound)
}

func assertNoAuthDestination(t *testing.T, client *clients.DestinationClient, serviceURL, noAuthDestinationName, noAuthDestinationURL, instanceID, ownerSubaccountID, authToken string) {
	noAuthDestBytes := client.FindDestinationByName(t, serviceURL, noAuthDestinationName, authToken, "", http.StatusOK)
	var noAuthDest esmdestinationcreator.DestinationSvcNoAuthenticationDestResponse
	err := json.Unmarshal(noAuthDestBytes, &noAuthDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, noAuthDest.Owner.SubaccountID)
	require.Equal(t, instanceID, noAuthDest.Owner.InstanceID)
	require.Equal(t, noAuthDestinationName, noAuthDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, noAuthDest.DestinationConfiguration.Type)
	require.Equal(t, noAuthDestinationURL, noAuthDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeNoAuth, noAuthDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, noAuthDest.DestinationConfiguration.ProxyType)
}

func assertBasicDestination(t *testing.T, client *clients.DestinationClient, serviceURL, basicDestinationName, basicDestinationURL, instanceID, ownerSubaccountID, authToken string, expectedNumberOfAuthTokens int) {
	basicDestBytes := client.FindDestinationByName(t, serviceURL, basicDestinationName, authToken, "", http.StatusOK)
	var basicDest esmdestinationcreator.DestinationSvcBasicDestResponse
	err := json.Unmarshal(basicDestBytes, &basicDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, basicDest.Owner.SubaccountID)
	require.Equal(t, instanceID, basicDest.Owner.InstanceID)
	require.Equal(t, basicDestinationName, basicDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, basicDest.DestinationConfiguration.Type)
	require.Equal(t, basicDestinationURL, basicDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeBasic, basicDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, basicDest.DestinationConfiguration.ProxyType)

	for i := 0; i < expectedNumberOfAuthTokens; i++ {
		require.NotEmpty(t, basicDest.AuthTokens)
		require.NotEmpty(t, basicDest.AuthTokens[i].Type)
		require.Equal(t, basicAuthType, basicDest.AuthTokens[i].Type)
		require.NotEmpty(t, basicDest.AuthTokens[i].Value)
	}
}

func assertSAMLAssertionDestination(t *testing.T, client *clients.DestinationClient, serviceURL, samlAssertionDestinationName, samlAssertionCertName, samlAssertionDestinationURL, app1BaseURL, instanceID, ownerSubaccountID, authToken, userTokenHeader string, expectedCertNames map[string]bool) {
	samlAssertionDestBytes := client.FindDestinationByName(t, serviceURL, samlAssertionDestinationName, authToken, userTokenHeader, http.StatusOK)
	var samlAssertionDest esmdestinationcreator.DestinationSvcSAMLAssertionDestResponse
	err := json.Unmarshal(samlAssertionDestBytes, &samlAssertionDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, samlAssertionDest.Owner.SubaccountID)
	require.Equal(t, instanceID, samlAssertionDest.Owner.InstanceID)
	require.Equal(t, samlAssertionDestinationName, samlAssertionDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, samlAssertionDest.DestinationConfiguration.Type)
	require.Equal(t, samlAssertionDestinationURL, samlAssertionDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeSAMLAssertion, samlAssertionDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, samlAssertionDest.DestinationConfiguration.ProxyType)
	require.Equal(t, app1BaseURL, samlAssertionDest.DestinationConfiguration.Audience)
	require.Equal(t, samlAssertionCertName+directordestinationcreator.JavaKeyStoreFileExtension, samlAssertionDest.DestinationConfiguration.KeyStoreLocation)

	require.Equal(t, len(expectedCertNames), len(samlAssertionDest.CertificateDetails))
	for i := 0; i < len(expectedCertNames); i++ {
		require.True(t, expectedCertNames[samlAssertionDest.CertificateDetails[i].Name])
		require.NotEmpty(t, samlAssertionDest.CertificateDetails[i].Content)
	}

	require.NotEmpty(t, samlAssertionDest.AuthTokens)
	for _, token := range samlAssertionDest.AuthTokens {
		require.Equal(t, samlAuthType, token.Type)
		require.NotEmpty(t, token.Value)
	}
}

func assertClientCertAuthDestination(t *testing.T, client *clients.DestinationClient, serviceURL, clientCertAuthDestinationName, clientCertAuthCertName, clientCertAuthDestinationURL, instanceID, ownerSubaccountID, authToken string, expectedCertNames map[string]bool) {
	clientCertAuthDestBytes := client.FindDestinationByName(t, serviceURL, clientCertAuthDestinationName, authToken, "", http.StatusOK)
	var clientCertAuthDest esmdestinationcreator.DestinationSvcClientCertDestResponse
	err := json.Unmarshal(clientCertAuthDestBytes, &clientCertAuthDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, clientCertAuthDest.Owner.SubaccountID)
	require.Equal(t, instanceID, clientCertAuthDest.Owner.InstanceID)
	require.Equal(t, clientCertAuthDestinationName, clientCertAuthDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, clientCertAuthDest.DestinationConfiguration.Type)
	require.Equal(t, clientCertAuthDestinationURL, clientCertAuthDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeClientCertificate, clientCertAuthDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, clientCertAuthDest.DestinationConfiguration.ProxyType)
	require.Equal(t, clientCertAuthCertName+directordestinationcreator.JavaKeyStoreFileExtension, clientCertAuthDest.DestinationConfiguration.KeyStoreLocation)

	require.Equal(t, len(expectedCertNames), len(clientCertAuthDest.CertificateDetails))
	for i := 0; i < len(expectedCertNames); i++ {
		require.True(t, expectedCertNames[clientCertAuthDest.CertificateDetails[i].Name])
		require.NotEmpty(t, clientCertAuthDest.CertificateDetails[i].Content)
	}
}

func assertOAuth2ClientCredsDestination(t *testing.T, client *clients.DestinationClient, serviceURL, oauth2ClientCredsDestinationName, oauth2ClientCredsDestinationURL, instanceID, ownerSubaccountID, authToken string, expectedNumberOfAuthTokens int) {
	oauth2ClientCredsDestBytes := client.FindDestinationByName(t, serviceURL, oauth2ClientCredsDestinationName, authToken, "", http.StatusOK)
	var oauth2ClientCredsDest esmdestinationcreator.DestinationSvcOAuth2ClientCredsDestResponse
	err := json.Unmarshal(oauth2ClientCredsDestBytes, &oauth2ClientCredsDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, oauth2ClientCredsDest.Owner.SubaccountID)
	require.Equal(t, instanceID, oauth2ClientCredsDest.Owner.InstanceID)
	require.Equal(t, oauth2ClientCredsDestinationName, oauth2ClientCredsDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, oauth2ClientCredsDest.DestinationConfiguration.Type)
	require.Equal(t, oauth2ClientCredsDestinationURL, oauth2ClientCredsDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeOAuth2ClientCredentials, oauth2ClientCredsDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, oauth2ClientCredsDest.DestinationConfiguration.ProxyType)

	for i := 0; i < expectedNumberOfAuthTokens; i++ {
		require.NotEmpty(t, oauth2ClientCredsDest.AuthTokens)
		require.NotEmpty(t, oauth2ClientCredsDest.AuthTokens[i].Type)
		require.Equal(t, oauth2AuthType, oauth2ClientCredsDest.AuthTokens[i].Type)
		require.NotEmpty(t, oauth2ClientCredsDest.AuthTokens[i].Value)
	}
}

func assertOAuth2mTLSDestination(t *testing.T, client *clients.DestinationClient, serviceURL, oauth2mTLSDestinationName, oauth2mTLSCertName, oauth2mTLSDestinationURL, instanceID, ownerSubaccountID, authToken string, expectedNumberOfAuthTokens int) {
	oauth2mTLSDestBytes := client.FindDestinationByName(t, serviceURL, oauth2mTLSDestinationName, authToken, "", http.StatusOK)

	var oauth2mTLSDest esmdestinationcreator.DestinationSvcOAuth2mTLSDestResponse
	err := json.Unmarshal(oauth2mTLSDestBytes, &oauth2mTLSDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, oauth2mTLSDest.Owner.SubaccountID)
	require.Equal(t, instanceID, oauth2mTLSDest.Owner.InstanceID)
	require.Equal(t, oauth2mTLSDestinationName, oauth2mTLSDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, oauth2mTLSDest.DestinationConfiguration.Type)
	require.Equal(t, oauth2mTLSDestinationURL, oauth2mTLSDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeOAuth2ClientCredentials, oauth2mTLSDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, oauth2mTLSDest.DestinationConfiguration.ProxyType)
	require.Equal(t, oauth2mTLSCertName+directordestinationcreator.JavaKeyStoreFileExtension, oauth2mTLSDest.DestinationConfiguration.KeyStoreLocation)

	for i := 0; i < expectedNumberOfAuthTokens; i++ {
		require.NotEmpty(t, oauth2mTLSDest.AuthTokens)
		require.NotEmpty(t, oauth2mTLSDest.AuthTokens[i].Type)
		require.Equal(t, oauth2AuthType, oauth2mTLSDest.AuthTokens[i].Type)
		require.NotEmpty(t, oauth2mTLSDest.AuthTokens[i].Value)
	}
}

func assertFormationAssignmentsNotificationWithItemsStructure(t *testing.T, notification gjson.Result, op, formationID, expectedAppID, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string) {
	assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, notification, op, formationID, expectedAppID, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID, nil)
}

func assertFormationAssignmentsNotification(t *testing.T, notification gjson.Result, op, formationID, expectedSourceAppID, expectedTargetAppID, expectedReceiverTenantState, expectedAssignedTenantState, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string) {
	assertFormationAssignmentsNotificationWithConfig(t, notification, op, formationID, expectedSourceAppID, expectedTargetAppID, expectedReceiverTenantState, expectedAssignedTenantState, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID, nil)
}

func assertFormationAssignmentsNotificationSubdomainWithItemsStructure(t *testing.T, notification gjson.Result, expectedSubdomain string) {
	notificationItems := notification.Get("RequestBody.items")
	require.True(t, notificationItems.Exists())
	require.Len(t, notificationItems.Array(), 1)

	app1FromNotification := notificationItems.Array()[0]
	require.Equal(t, expectedSubdomain, app1FromNotification.Get("subdomain").String())
}

func assertNoNotificationsAreSentForTenant(t *testing.T, client *http.Client, tenantID string) {
	assertNoNotificationsAreSent(t, client, tenantID)
}

func assertNoNotificationsAreSent(t *testing.T, client *http.Client, objectID string) {
	body := getNotificationsFromExternalSvcMock(t, client)
	notifications := gjson.GetBytes(body, objectID)
	require.False(t, notifications.Exists())
	require.Len(t, notifications.Array(), 0)
}

func assertNotificationsCountMoreThanForTenant(t *testing.T, body []byte, tenantID string, count int) {
	assertNotificationsCountMoreThan(t, body, tenantID, count)
}

func assertNotificationsCountForTenant(t *testing.T, body []byte, tenantID string, count int) {
	assertNotificationsCount(t, body, tenantID, count)
}

func assertNotificationsCountForFormationID(t *testing.T, body []byte, formationID string, count int) {
	assertNotificationsCount(t, body, formationID, count)
}

func assertNotificationsCountMoreThan(t *testing.T, body []byte, objectID string, count int) {
	notifications := gjson.GetBytes(body, objectID)
	if count > 0 {
		require.True(t, notifications.Exists())
		length := len(notifications.Array())
		require.GreaterOrEqual(t, length, count)
	} else {
		require.False(t, notifications.Exists())
	}
}

func assertNotificationsCount(t *testing.T, body []byte, objectID string, count int) {
	notifications := gjson.GetBytes(body, objectID)
	if count > 0 {
		require.True(t, notifications.Exists())
		require.Len(t, notifications.Array(), count)
	} else {
		require.False(t, notifications.Exists())
	}
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
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusOK, string(body)))
	return body
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

func assertFormationAssignmentsNotificationWithConfig(t *testing.T, notification gjson.Result, op, formationID, expectedSourceAppID, expectedTargetAppID, expectedReceiverTenantState, expectedAssignedTenantState, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string, expectedConfig *string) {
	require.Equal(t, op, notification.Get("Operation").String())
	if op == unassignOperation {
		require.Equal(t, expectedSourceAppID, notification.Get("ApplicationID").String())
	}
	require.Equal(t, op, notification.Get("RequestBody.context.operation").String())
	require.Equal(t, formationID, notification.Get("RequestBody.context.uclFormationId").String())
	require.Equal(t, expectedTenant, notification.Get("RequestBody.context.globalAccountId").String())
	require.Equal(t, expectedCustomerID, notification.Get("RequestBody.context.crmId").String())

	require.Equal(t, expectedReceiverTenantState, notification.Get("RequestBody.receiverTenant.state").String())
	require.Equal(t, expectedTargetAppID, notification.Get("RequestBody.receiverTenant.uclSystemTenantId").String())
	require.Equal(t, expectedLocalTenantID, notification.Get("RequestBody.receiverTenant.applicationTenantId").String())
	require.Equal(t, expectedAppNamespace, notification.Get("RequestBody.receiverTenant.applicationNamespace").String())
	require.Equal(t, expectedAppRegion, notification.Get("RequestBody.receiverTenant.deploymentRegion").String())
	if expectedConfig != nil {
		require.Equal(t, *expectedConfig, notification.Get("RequestBody.receiverTenant.configuration").String())
	}

	require.Equal(t, expectedAssignedTenantState, notification.Get("RequestBody.assignedTenant.state").String())
}

func assertFormationNotificationFromCreationOrDeletion(t *testing.T, body []byte, formationID, formationName, formationOperation, tenantID, parentTenantID string) {
	t.Logf("Assert synchronous formation lifecycle notifications are sent for %q operation...", formationOperation)
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
	t.Logf("Synchronous formation lifecycle notifications are successfully validated for %q operation.", formationOperation)
}

func assertAsyncFormationNotificationFromCreationOrDeletionWithEventually(t *testing.T, ctx context.Context, body []byte, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID string, timeout, tick time.Duration) {
	var shouldExpectDeleted bool
	if formationOperation == createFormationOperation || formationState == "DELETE_ERROR" {
		shouldExpectDeleted = false
	} else {
		shouldExpectDeleted = true
	}
	assertAsyncFormationNotificationFromCreationOrDeletionExpectDeletedWithEventually(t, ctx, body, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID, shouldExpectDeleted, timeout, tick)
}
func assertAsyncFormationNotificationFromCreationOrDeletionExpectDeletedWithEventually(t *testing.T, ctx context.Context, body []byte, formationID, formationName, formationState, formationOperation, tenantID, parentTenantID string, shouldExpectDeleted bool, timeout, tick time.Duration) {
	t.Logf("Assert asynchronous formation lifecycle notifications are sent for %q operation...", formationOperation)
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

	t.Logf("Asserting formation with eventually...")
	tOnce := testingx.NewOnceLogger(t)
	require.Eventually(t, func() (isOkay bool) {
		tOnce.Log("Assert formation lifecycle notifications are successfully processed...")
		formationPage := fixtures.ListFormationsWithinTenant(t, ctx, tenantID, certSecuredGraphQLClient)
		if shouldExpectDeleted {
			if formationPage.TotalCount != 0 {
				tOnce.Logf("Formation lifecycle notification is expected to have deleted formation with ID %q, but it is still there", formationID)
				return
			}
			if formationPage.Data != nil && len(formationPage.Data) > 0 {
				tOnce.Logf("Formation lifecycle notification is expected to have deleted formation with ID %q, but it is still there", formationID)
				return
			}
		} else {
			if formationPage.TotalCount != 1 {
				tOnce.Log("Formation count does not match")
				return
			}
			if formationPage.Data[0].State != formationState {
				tOnce.Logf("Formation state for formation with ID %q is %q, expected: %q", formationID, formationPage.Data[0].State, formationState)
				return
			}
			if formationPage.Data[0].ID != formationID {
				tOnce.Logf("Formation ID is %q, expected: %q", formationPage.Data[0].ID, formationID)
				return
			}
			if formationPage.Data[0].Name != formationName {
				tOnce.Logf("Formation name is %q, expected: %q", formationPage.Data[0].Name, formationName)
				return
			}
		}

		tOnce.Logf("Asynchronous formation lifecycle notifications are successfully validated for %q operation.", formationOperation)
		return true
	}, timeout, tick)
}

func assertSeveralFormationAssignmentsNotifications(t *testing.T, notificationsForConsumerTenant gjson.Result, rtCtx *graphql.RuntimeContextExt, formationID, region, operationType, expectedTenant, expectedCustomerID string, expectedNumberOfNotifications int) {
	actualNumberOfNotifications := 0
	for _, notification := range notificationsForConsumerTenant.Array() {
		rtCtxIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
		op := notification.Get("Operation").String()
		if rtCtxIDFromNotification == rtCtx.ID && op == operationType {
			t.Logf("Found notification about rtCtx %q", rtCtxIDFromNotification)
			actualNumberOfNotifications++
			err := verifyFormationNotificationForApplicationWithItemsStructure(notification, operationType, formationID, rtCtx.ID, rtCtx.Value, region, "", expectedTenant, expectedCustomerID)
			assert.NoError(t, err)
		}
	}
	require.Equal(t, expectedNumberOfNotifications, actualNumberOfNotifications)
}

type applicationFormationExpectations struct {
	op                                     string
	formationID                            string
	objectID                               string
	localTenantID                          string
	objectRegion                           string
	configuration                          string
	tenant                                 string
	customerID                             string
	receiverTenantState                    string
	assignedTenantState                    string
	shouldRemoveDestinationCertificateData bool
}

func assertExpectationsForApplicationNotificationsWithItemsStructure(t *testing.T, notifications []gjson.Result, expectations []*applicationFormationExpectations) {
	assert.Equal(t, len(expectations), len(notifications))
	for _, expectation := range expectations {
		found := false
		for _, notification := range notifications {
			err := verifyFormationNotificationForApplicationWithItemsStructure(notification, expectation.op, expectation.formationID, expectation.objectID, expectation.localTenantID, expectation.objectRegion, expectation.configuration, expectation.tenant, expectation.customerID)
			if err == nil {
				found = true
			}
		}
		assert.Truef(t, found, "Did not match expectations for notification %v", expectation)
	}
}

func verifyFormationNotificationForApplicationWithItemsStructure(notification gjson.Result, op, formationID, expectedObjectID, expectedSubscribedTenantID, expectedObjectRegion, expectedConfiguration, expectedTenant, expectedCustomerID string) error {
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

func assertExpectationsForApplicationNotifications(t *testing.T, notifications []gjson.Result, expectations []*applicationFormationExpectations) {
	for _, expectation := range expectations {
		found := false
		for _, notification := range notifications {
			if err := verifyFormationAssignmentNotification(t, notification, expectation.op, expectation.formationID, expectation.objectID, expectation.localTenantID, expectation.objectRegion, expectation.configuration, expectation.tenant, expectation.customerID, expectation.shouldRemoveDestinationCertificateData); err != nil {
				t.Log(err)
				continue
			}
			found = true
		}
		assert.Truef(t, found, "Did not match expectations for notification %v", expectation)
	}
}

func verifyFormationAssignmentNotification(t *testing.T, notification gjson.Result, op, formationID, expectedObjectID, expectedAppLocalTenantID, expectedObjectRegion, expectedConfiguration, expectedTenant, expectedCustomerID string, shouldRemoveDestinationCertificateData bool) error {
	actualOp := notification.Get("Operation").String()
	if op != actualOp {
		return errors.Errorf("Operation does not match - expected: %q, but got: %q", op, actualOp)
	}

	if op == unassignOperation {
		if actualObjectIDExists := notification.Get("ApplicationID").Exists(); !actualObjectIDExists {
			return errors.New("ObjectID does not exist")
		}

		actualObjectID := notification.Get("ApplicationID").String()
		if expectedObjectID != actualObjectID {
			return errors.Errorf("ObjectID does not match - expected: %q, but got: %q", expectedObjectID, actualObjectID)
		}
	}

	actualFormationID := notification.Get("RequestBody.context.uclFormationId").String()
	if formationID != actualFormationID {
		return errors.Errorf("RequestBody.context.uclFormationId does not match - expected: %q, but got: %q", formationID, actualFormationID)
	}

	actualTenantID := notification.Get("RequestBody.context.globalAccountId").String()
	if expectedTenant != actualTenantID {
		return errors.Errorf("RequestBody.context.globalAccountId does not match - expected: %q, but got: %q", expectedTenant, actualTenantID)
	}

	actualCustomerID := notification.Get("RequestBody.context.crmId").String()
	if expectedCustomerID != actualCustomerID {
		return errors.Errorf("RequestBody.context.crmId does not match - expected: %q, but got: %q", expectedCustomerID, actualCustomerID)
	}

	actualAppTenantID := notification.Get("RequestBody.receiverTenant.applicationTenantId").String()
	if expectedAppLocalTenantID != actualAppTenantID {
		return errors.Errorf("RequestBody.receiverTenant.applicationTenantId does not match - expected: %q, but got: %q", expectedAppLocalTenantID, actualAppTenantID)
	}

	actualObjectRegion := notification.Get("RequestBody.receiverTenant.deploymentRegion").String()
	if expectedObjectRegion != actualObjectRegion {
		return errors.Errorf("RequestBody.receiverTenant.deploymentRegion does not match - expected: %q, but got: %q", expectedObjectRegion, actualObjectRegion)
	}

	if shouldRemoveDestinationCertificateData {
		notificationReceiverCfg := notification.Get("RequestBody.receiverTenant.configuration").String()
		notificationReceiverState := notification.Get("RequestBody.receiverTenant.state").String()
		if notificationReceiverCfg == "" && notificationReceiverState == "INITIAL" {
			return nil
		}

		modifiedNotification, err := sjson.Delete(notification.String(), "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.samlAssertion.certificate")
		if err != nil {
			return err
		}

		modifiedNotification, err = sjson.Delete(modifiedNotification, "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.clientCertificateAuthentication.certificate")
		if err != nil {
			return err
		}

		modifiedNotification, err = sjson.Delete(modifiedNotification, "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.oauth2mtls.certificate")
		if err != nil {
			return err
		}

		modifiedNotification, err = sjson.Delete(modifiedNotification, "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.samlAssertion.assertionIssuer")
		if err != nil {
			return err
		}

		modifiedConfig := gjson.Get(modifiedNotification, "RequestBody.receiverTenant.configuration").String()
		assert.JSONEq(t, expectedConfiguration, modifiedConfig, "RequestBody.receiverTenant.configuration does not match")
	} else {
		actualConfiguration := notification.Get("RequestBody.receiverTenant.configuration").String()
		if expectedConfiguration != "" && expectedConfiguration != actualConfiguration {
			return errors.Errorf("RequestBody.receiverTenant.configuration does not match - expected: %q, but got: %q", expectedConfiguration, actualConfiguration)
		}
	}

	return nil
}

func validateFormationNameInAssignmentNotification(t *testing.T, jsonResult gjson.Result, expectedFormationName string) {
	validateJSONStringProperty(t, jsonResult, "RequestBody.context.uclFormationName", expectedFormationName)
}

func validateJSONStringProperty(t *testing.T, jsonResult gjson.Result, path, expectedValue string) {
	require.Equal(t, expectedValue, jsonResult.Get(path).String())
}

func buildConsumerTokenURL(providerTokenURL, consumerSubdomain string) (string, error) {
	baseTokenURL, err := url.Parse(providerTokenURL)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to parse auth url '%s'", providerTokenURL)
	}
	parts := strings.Split(baseTokenURL.Hostname(), ".")
	if len(parts) < 2 {
		return "", errors.Errorf("Provider auth URL: '%s' should have a subdomain", providerTokenURL)
	}
	originalSubdomain := parts[0]

	tokenURL := strings.Replace(providerTokenURL, originalSubdomain, consumerSubdomain, 1)
	return tokenURL, nil
}

func executeFAStatusResetReqWithExpectedStatusCode(t *testing.T, certSecuredHTTPClient *http.Client, state, testConfig, tnt, formationID, formationAssignmentID string, expectedStatusCode int) {
	reqBody := FormationAssignmentRequestBody{
		State:         state,
		Configuration: json.RawMessage(testConfig),
	}
	marshalBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	formationAssignmentAsyncStatusAPIEndpoint := resolveFAAsyncStatusResetAPIURL(formationID, formationAssignmentID)
	request, err := http.NewRequest(http.MethodPatch, formationAssignmentAsyncStatusAPIEndpoint, bytes.NewBuffer(marshalBody))
	require.NoError(t, err)
	request.Header.Add("Content-Type", "application/json")
	// The Tenant header is needed in case we are simulating an application reporting status for its own formation assignment.
	request.Header.Add("Tenant", tnt)
	response, err := certSecuredHTTPClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, response.StatusCode)
}

func findConstraintByName(t *testing.T, name string, actualFormationConstraints []*graphql.FormationConstraint) *graphql.FormationConstraint {
	for _, constraint := range actualFormationConstraints {
		if constraint.Name == name {
			return constraint
		}
	}
	require.Failf(t, "Could not find constraint with name %q", name)
	return nil
}
