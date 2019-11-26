package director

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// Auth Validation

func TestSetAPIAuth_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	app := createApplication(t, ctx, "validation-test-app")
	defer deleteApplication(t, app.ID)
	require.Len(t, app.Apis.Data, 1)
	rtm := createRuntime(t, ctx, "validation-test-rtm")
	defer deleteRuntime(t, rtm.ID)

	invalidAuthInput := graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "custom",
				Password: "auth",
			},
			Oauth: &graphql.OAuthCredentialDataInput{
				ClientID:     "v",
				ClientSecret: "v",
				URL:          "http://v.url",
			},
		},
	}
	authInStr, err := tc.graphqlizer.AuthInputToGQL(&invalidAuthInput)
	require.NoError(t, err)
	var apiRtmAuth graphql.APIRuntimeAuth
	request := fixSetAPIAuthRequest(app.Apis.Data[0].ID, rtm.ID, authInStr)

	// WHEN
	err = tc.RunOperation(ctx, request, &apiRtmAuth)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type AuthInput")
}

// Webhook Validation

func TestAddWebhook_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	app := createApplication(t, ctx, "validation-test-app")
	defer deleteApplication(t, app.ID)

	invalidWebhookInput := graphql.WebhookInput{
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  "invalid",
	}
	webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&invalidWebhookInput)
	require.NoError(t, err)
	var webhook graphql.Webhook
	request := fixAddWebhookRequest(app.ID, webhookInStr)

	// WHEN
	err = tc.RunOperation(ctx, request, &webhook)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type WebhookInput")
}

func TestUpdateWebhook_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	app := createApplication(t, ctx, "validation-test-app")
	defer deleteApplication(t, app.ID)
	require.Len(t, app.Webhooks, 1)

	invalidWebhookInput := graphql.WebhookInput{
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  "invalid",
	}
	webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&invalidWebhookInput)
	require.NoError(t, err)
	var webhook graphql.Webhook
	request := fixUpdateWebhookRequest(app.Webhooks[0].ID, webhookInStr)

	// WHEN
	err = tc.RunOperation(ctx, request, &webhook)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type WebhookInput")
}
