package tests

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestOperation(t *testing.T) {
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	in := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks("test", "register-app", "http://test.test", nil)
	in.LocalTenantID = nil
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}

	t.Log("Registering Application with ORD Webhook")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createRequest, &app)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotEmpty(t, app.ID)
	require.Equal(t, 1, len(app.Operations))
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, app.Operations[0].OperationType)

	t.Logf("Getting operation with ID: %s", app.Operations[0].ID)
	getOperationRequest := fixtures.FixGetOperationByIDRequest(app.Operations[0].ID)
	op := graphql.Operation{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, getOperationRequest, &op)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.NotEmpty(t, op.ID)
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, op.OperationType)

	example.SaveExample(t, getOperationRequest.Query(), "get operation by id")
}

func TestOperationSchedule(t *testing.T) {
	appInput := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks("app-name", "description", conf.ExternalServicesMockAbsoluteURL, nil)
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	ctx := context.Background()

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appInput)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	waitForOperationToFinish(t, ctx, app.Operations[0].ID)

	t.Logf("Rescheduling operation with ID: %s", app.Operations[0].ID)
	scheduleOperationRequest := fixtures.FixScheduleOperationByIDRequest(app.Operations[0].ID, 1)
	op := graphql.Operation{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, scheduleOperationRequest, &op)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.NotEmpty(t, op.ID)
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, op.OperationType)
	require.Equal(t, graphql.OperationStatusScheduled, op.Status)

	example.SaveExample(t, scheduleOperationRequest.Query(), "schedule operation")
}

func waitForOperationToFinish(t *testing.T, ctx context.Context, opID string) {
	require.Eventually(t, func() bool {
		getOperationRequest := fixtures.FixGetOperationByIDRequest(opID)
		currentOperation := graphql.Operation{}
		err := testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, getOperationRequest, &currentOperation)
		require.NoError(t, err)

		t.Logf("Operation with ID %q is found: %v.", opID, currentOperation)
		if currentOperation.Status == graphql.OperationStatusCompleted || currentOperation.Status == graphql.OperationStatusFailed {
			t.Logf("Operation with ID %q is in status %v.", opID, currentOperation.Status)
			return true
		}

		t.Logf("Operation with ID %q is still not completed. Current status is: %v.", opID, currentOperation.Status)
		return false
	}, time.Second*90, time.Second*1, "Waiting for operation to finish.")
}
