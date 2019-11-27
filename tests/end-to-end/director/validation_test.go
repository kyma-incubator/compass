package director

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runtime Validation

func TestCreateRuntime_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := tc.graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixCreateRuntimeRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type RuntimeInput")
}

func TestUpdateRuntime_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	rtm := createRuntime(t, ctx, "validation-test-rtm")
	defer deleteRuntime(t, rtm.ID)

	invalidInput := graphql.RuntimeInput{
		Name: "0invalid",
	}
	inputString, err := tc.graphqlizer.RuntimeInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixUpdateRuntimeRequest(rtm.ID, inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type RuntimeInput")
}

// Label Definition Validation

func TestCreateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	invalidInput := graphql.LabelDefinitionInput{
		Key: "",
	}
	inputString, err := tc.graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixCreateLabelDefinitionRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelDefinitionInput")
}

func TestUpdateLabelDefinition_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	key := "test-validation-ld"
	ld := createLabelDefinitionWithinTenant(t, ctx, key, map[string]string{"type": "string"}, defaultTenant)
	defer deleteLabelDefinitionWithinTenant(t, ctx, ld.Key, true, defaultTenant)
	invalidSchema := graphql.JSONSchema(`"{\"test\":}"`)
	invalidInput := graphql.LabelDefinitionInput{
		Key:    key,
		Schema: &invalidSchema,
	}
	inputString, err := tc.graphqlizer.LabelDefinitionInputToGQL(invalidInput)
	require.NoError(t, err)
	var result graphql.Runtime
	request := fixUpdateLabelDefinitionRequest(inputString)

	// WHEN
	err = tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelDefinitionInput")
}

// Label Validation

func TestSetApplicationLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	app := createApplication(t, ctx, "validation-test-app")
	defer deleteApplication(t, app.ID)

	request := fixSetApplicationLabelRequest(app.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err := tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelInput")
}

func TestSetRuntimeLabel_Validation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	rtm := createRuntime(t, ctx, "validation-test-rtm")
	defer deleteRuntime(t, rtm.ID)

	request := fixSetRuntimeLabelRequest(rtm.ID, strings.Repeat("x", 257), "")
	var result graphql.Label

	// WHEN
	err := tc.RunOperation(ctx, request, &result)

	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation error for type LabelInput")
}

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
