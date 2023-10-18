package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

func FixFormationNotificationWebhookInput(webhookType graphql.WebhookType, mode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string) *graphql.WebhookInput {
	var inputTmpl *string
	if inputTemplate != "" {
		inputTmpl = &inputTemplate
	}

	return &graphql.WebhookInput{
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		Mode:           &mode,
		URLTemplate:    &urlTemplate,
		InputTemplate:  inputTmpl,
		OutputTemplate: &outputTemplate,
	}
}

func FixNonFormationNotificationWebhookInput(webhookType graphql.WebhookType) *graphql.WebhookInput {
	return &graphql.WebhookInput{
		URL:            str.Ptr("http://new-webhook.url"),
		Type:           webhookType,
		OutputTemplate: str.Ptr("{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"),
	}
}
