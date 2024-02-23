package operations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const (
	formationIDPathParam           = "ucl-formation-id"
	formationAssignmentIDPathParam = "ucl-assignment-id"
)

type FormationAssignmentRequestBody struct {
	State         string          `json:"state"`
	Configuration json.RawMessage `json:"configuration"`
	Error         string          `json:"error"`
}

type ExecuteStatusReportOperation struct {
	externalServicesMockMtlsSecuredURL string
	client                             *http.Client
	tenant                             string
	assignmentSource                   string
	assignmentTarget                   string
	config                             string
	state                              string
	statusCode                         int
	asserters                          []asserters.Asserter
}

func NewExecuteStatusReportOperation() *ExecuteStatusReportOperation {
	return &ExecuteStatusReportOperation{
		statusCode: http.StatusOK,
		state:      "READY",
	}
}

func (o *ExecuteStatusReportOperation) WithExternalServicesMockMtlsSecuredURL(externalServicesMockMtlsSecuredURL string) *ExecuteStatusReportOperation {
	o.externalServicesMockMtlsSecuredURL = externalServicesMockMtlsSecuredURL
	return o
}

func (o *ExecuteStatusReportOperation) WithHTTPClient(client *http.Client) *ExecuteStatusReportOperation {
	o.client = client
	return o
}

func (o *ExecuteStatusReportOperation) WithFormationAssignment(source, target string) *ExecuteStatusReportOperation {
	o.assignmentSource = source
	o.assignmentTarget = target
	return o
}

func (o *ExecuteStatusReportOperation) WithStatusCode(statusCode int) *ExecuteStatusReportOperation {
	o.statusCode = statusCode
	return o
}

func (o *ExecuteStatusReportOperation) WithTenant(tenant string) *ExecuteStatusReportOperation {
	o.tenant = tenant
	return o
}

func (o *ExecuteStatusReportOperation) WithState(state string) *ExecuteStatusReportOperation {
	o.state = state
	return o
}

func (o *ExecuteStatusReportOperation) WithAsserters(asserters ...asserters.Asserter) *ExecuteStatusReportOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *ExecuteStatusReportOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)

	t.Logf("List formation assignments for formation with ID: %q", formationID)
	listFormationAssignmentsReq := fixtures.FixListFormationAssignmentRequest(formationID, 100)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, gqlClient, o.tenant, listFormationAssignmentsReq)

	formationAssignmentID := getFormationAssignmentIDBySourceAndTarget(t, assignmentsPage, o.assignmentSource, o.assignmentTarget)

	t.Logf("Calling FA status API for formation assignment ID %q", formationAssignmentID)
	faAsyncStatusAPIURL := strings.Replace(o.externalServicesMockMtlsSecuredURL, fmt.Sprintf("{%s}", formationIDPathParam), formationID, 1)
	faAsyncStatusAPIURL = strings.Replace(faAsyncStatusAPIURL, fmt.Sprintf("{%s}", formationAssignmentIDPathParam), formationAssignmentID, 1)
	reqBody := FormationAssignmentRequestBody{
		State: o.state,
	}
	if o.config != "" {
		reqBody.Configuration = json.RawMessage(o.config)
	}
	marshalBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPatch, faAsyncStatusAPIURL, bytes.NewBuffer(marshalBody))
	require.NoError(t, err)
	request.Header.Add("Content-Type", "application/json")
	response, err := o.client.Do(request)
	require.NoError(t, err)
	require.Equal(t, o.statusCode, response.StatusCode)
	// Log the unsuccessful HTTP responses
	if o.statusCode > 299 {
		bodyBytes, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		t.Logf("Response was: %s", string(bodyBytes))
	}
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *ExecuteStatusReportOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
}

func (o *ExecuteStatusReportOperation) Operation() Operation {
	return o
}

func getFormationAssignmentIDBySourceAndTarget(t *testing.T, assignmentsPage *graphql.FormationAssignmentPage, sourceID, targetID string) string {
	var formationAssignmentID string
	for _, a := range assignmentsPage.Data {
		if a.Source == sourceID && a.Target == targetID {
			formationAssignmentID = a.ID
		}
	}
	require.NotEmptyf(t, formationAssignmentID, "The formation assignment with ID %q should not be empty", formationAssignmentID)
	return formationAssignmentID
}
