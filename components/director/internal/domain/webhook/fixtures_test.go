package webhook_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixModelWebhook(id, appID, url string) *model.ApplicationWebhook {
	return &model.ApplicationWebhook{
		ID:            id,
		ApplicationID: appID,
		Type:          model.ApplicationWebhookTypeConfigurationChanged,
		URL:           url,
		Auth:          &model.Auth{},
	}
}

func fixGQLWebhook(id, appID, url string) *graphql.ApplicationWebhook {
	return &graphql.ApplicationWebhook{
		ID:            id,
		ApplicationID: appID,
		Type:          graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:           url,
		Auth:          &graphql.Auth{},
	}
}

func fixModelWebhookInput(url string) *model.ApplicationWebhookInput {
	return &model.ApplicationWebhookInput{
		Type: model.ApplicationWebhookTypeConfigurationChanged,
		URL:  url,
		Auth: &model.AuthInput{},
	}
}

func fixGQLWebhookInput(url string) *graphql.ApplicationWebhookInput {
	return &graphql.ApplicationWebhookInput{
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  url,
		Auth: &graphql.AuthInput{},
	}
}
