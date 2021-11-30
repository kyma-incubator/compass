package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testPkg "github.com/kyma-incubator/compass/tests/pkg/webhook"
	"github.com/stretchr/testify/require"
)

func TestSyncAPIDeleteApplicationWithMTLSAppWebhook(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	appName := fmt.Sprintf("app-sync-del-%s", time.Now().Format("060102150405"))
	mtlsWebhook := testPkg.BuildMockedWebhook(testConfig.ExternalServicesMockMTLSSecuredURL, graphql.WebhookTypeUnregisterApplication)
	mtlsWebhook.Mode = testPkg.WebhookModePtr(graphql.WebhookModeSync)

	appInput := graphql.ApplicationRegisterInput{
		Name:         appName,
		ProviderName: ptr.String("compass"),
		Webhooks:     []*graphql.WebhookInput{mtlsWebhook},
	}

	t.Log(fmt.Sprintf("Registering application: %s", appName))
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)

	registerRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testConfig.DefaultTestTenant, registerRequest, &app)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, &app)
	require.NoError(t, err)

	operationFullPath := testPkg.BuildOperationFullPath(testConfig.ExternalServicesMockBaseURL)
	require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))
}
