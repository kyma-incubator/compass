package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestOperation(t *testing.T) {
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "int-sys-ops")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	intSystemOAuthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	in := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks("test", "register-app", "http://test.test", nil)
	in.LocalTenantID = nil
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}

	t.Log("Registering Application with ORD Webhook")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, intSystemOAuthGraphQLClient, tenantId, createRequest, &app)
	defer fixtures.CleanupApplication(t, ctx, intSystemOAuthGraphQLClient, tenantId, &app)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotEmpty(t, app.ID)
	require.Equal(t, 1, len(app.Operations))
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, app.Operations[0].OperationType)

	t.Logf("Getting operation with ID: %s", app.Operations[0].ID)
	getOperationRequest := fixtures.FixGetOperationByIDRequest(app.Operations[0].ID)
	op := graphql.Operation{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, intSystemOAuthGraphQLClient, getOperationRequest, &op)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.NotEmpty(t, op.ID)
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, op.OperationType)

	example.SaveExample(t, getOperationRequest.Query(), "get operation by id")
}
