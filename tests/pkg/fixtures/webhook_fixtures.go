package fixtures

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

func FixFormationNotificationWebhookInput(webhookType graphql.WebhookType, mode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate, url string, headerTemplate *string) *graphql.WebhookInput {
	var inputTmpl *string
	if inputTemplate != "" {
		inputTmpl = &inputTemplate
	}

	var urlTpl *string
	if urlTemplate != "" {
		urlTpl = &urlTemplate
	}

	var urlPtr *string
	if url != "" {
		urlPtr = &url
	}

	spew.Dump("PTR", urlPtr)
	return &graphql.WebhookInput{
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		Mode:           &mode,
		URL:            urlPtr,
		URLTemplate:    urlTpl,
		InputTemplate:  inputTmpl,
		OutputTemplate: &outputTemplate,
		HeaderTemplate: headerTemplate,
	}
}

func FixNonFormationNotificationWebhookInput(webhookType graphql.WebhookType) *graphql.WebhookInput {
	return &graphql.WebhookInput{
		URL:            str.Ptr("http://new-webhook.url"),
		Type:           webhookType,
		OutputTemplate: str.Ptr("{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"),
	}
}
