package webhook_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelWebhook(id, appID, tenant, url string) *model.Webhook {
	return &model.Webhook{
		ID:            id,
		ApplicationID: &appID,
		TenantID:      tenant,
		Type:          model.WebhookTypeConfigurationChanged,
		URL:           url,
		Auth:          &model.Auth{},
	}
}

func fixGQLWebhook(id, appID, url string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:            id,
		ApplicationID: &appID,
		Type:          graphql.WebhookTypeConfigurationChanged,
		URL:           url,
		Auth:          &graphql.Auth{},
	}
}

func fixModelWebhookInput(url string) *model.WebhookInput {
	return &model.WebhookInput{
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            url,
		Auth:           &model.AuthInput{},
		Mode:           model.WebhookModeSync,
		URLTemplate:    `{}`,
		InputTemplate:  `{}`,
		HeaderTemplate: `{}`,
		OutputTemplate: `{}`,
	}
}

func fixGQLWebhookInput(url string) *graphql.WebhookInput {
	return &graphql.WebhookInput{
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            url,
		Auth:           &graphql.AuthInput{},
		Mode:           graphql.WebhookModeSync,
		URLTemplate:    `{}`,
		InputTemplate:  `{}`,
		HeaderTemplate: `{}`,
		OutputTemplate: `{}`,
	}
}
