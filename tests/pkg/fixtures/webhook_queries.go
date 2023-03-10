package fixtures

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func AddWebhookToRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in *graphql.WebhookInput, tenant, runtimeID string) *graphql.Webhook {
	runtimeWebhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(in)
	require.NoError(t, err)

	addWebhookRequest := FixAddWebhookToRuntimeRequest(runtimeID, runtimeWebhookInStr)
	actualWebhook := graphql.Webhook{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addWebhookRequest, &actualWebhook)
	require.NoError(t, err)
	require.NotNil(t, actualWebhook.ID)
	return &actualWebhook
}

func AddWebhookToApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in *graphql.WebhookInput, tenant, applicationID string) *graphql.Webhook {
	applicationWebhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(in)
	require.NoError(t, err)

	addWebhookRequest := FixAddWebhookToApplicationRequest(applicationID, applicationWebhookInStr)
	actualWebhook := graphql.Webhook{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addWebhookRequest, &actualWebhook)
	require.NoError(t, err)
	require.NotNil(t, actualWebhook.ID)
	return &actualWebhook
}

func AddWebhookToApplicationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in *graphql.WebhookInput, tenant, applicationTemplateID string) *graphql.Webhook {
	applicationTemplateWebhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(in)
	require.NoError(t, err)

	addWebhookRequest := FixAddWebhookToTemplateRequest(applicationTemplateID, applicationTemplateWebhookInStr)
	actualWebhook := graphql.Webhook{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addWebhookRequest, &actualWebhook)
	require.NoError(t, err)
	require.NotNil(t, actualWebhook.ID)
	return &actualWebhook
}

func AddWebhookToFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in *graphql.WebhookInput, tenant, formationTemplateID string) *graphql.Webhook {
	applicationTemplateWebhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(in)
	require.NoError(t, err)

	addWebhookRequest := FixAddWebhookToFormationTemplateRequest(formationTemplateID, applicationTemplateWebhookInStr)
	actualWebhook := graphql.Webhook{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addWebhookRequest, &actualWebhook)
	require.NoError(t, err)
	require.NotNil(t, actualWebhook.ID)
	return &actualWebhook
}

func DeleteWebhook(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, webhookID string) *graphql.Webhook {
	deleteWebhookRequest := FixDeleteWebhookRequest(webhookID)
	actualWebhook := graphql.Webhook{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteWebhookRequest, &actualWebhook)
	require.NoError(t, err)
	require.NotNil(t, actualWebhook.ID)
	return &actualWebhook
}

func CleanupWebhook(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, webhookID string) *graphql.Webhook {
	deleteWebhookRequest := FixDeleteWebhookRequest(webhookID)
	actualWebhook := graphql.Webhook{}
	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteWebhookRequest, &actualWebhook))
	return &actualWebhook
}
