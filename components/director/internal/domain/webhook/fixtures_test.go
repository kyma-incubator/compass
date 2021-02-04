package webhook_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelWebhook(id, appID, tenant, url string) *model.Webhook {
	return &model.Webhook{
		ID:             id,
		ApplicationID:  &appID,
		TenantID:       tenant,
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            url,
		Auth:           &model.Auth{},
		Mode:           &modelWebhookMode,
		URLTemplate:    &template,
		InputTemplate:  &template,
		HeaderTemplate: &template,
		OutputTemplate: &template,
	}
}

func fixGQLWebhook(id, appID, url string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:             id,
		ApplicationID:  &appID,
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            url,
		Auth:           &graphql.Auth{},
		Mode:           &graphqlWebhookMode,
		URLTemplate:    &template,
		InputTemplate:  &template,
		HeaderTemplate: &template,
		OutputTemplate: &template,
	}
}

func fixModelWebhookInput(url string) *model.WebhookInput {
	return &model.WebhookInput{
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            url,
		Auth:           &model.AuthInput{},
		Mode:           &modelWebhookMode,
		URLTemplate:    &template,
		InputTemplate:  &template,
		HeaderTemplate: &template,
		OutputTemplate: &template,
	}
}

func fixGQLWebhookInput(url string) *graphql.WebhookInput {
	return &graphql.WebhookInput{
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            url,
		Auth:           &graphql.AuthInput{},
		Mode:           &graphqlWebhookMode,
		URLTemplate:    &template,
		InputTemplate:  &template,
		HeaderTemplate: &template,
		OutputTemplate: &template,
	}
}
