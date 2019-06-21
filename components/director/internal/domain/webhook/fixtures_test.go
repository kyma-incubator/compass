package webhook_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

func fixModelWebhook(appID, id, url string) *model.ApplicationWebhook {
	return &model.ApplicationWebhook{
		ApplicationID: appID,
		ID:            id,
		Type:          model.ApplicationWebhookTypeConfigurationChanged,
		URL:           url,
		Auth:          &model.Auth{},
	}
}

func fixGQLWebhook(id, url string) *graphql.ApplicationWebhook {
	return &graphql.ApplicationWebhook{
		ID:   id,
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  url,
		Auth: &graphql.Auth{},
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
