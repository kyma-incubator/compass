package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

func FixFormationNotificationWebhookInput(webhookType graphql.WebhookType, mode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string) *graphql.WebhookInput {
	return &graphql.WebhookInput{
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		Mode:           &mode,
		URLTemplate:    &urlTemplate,
		InputTemplate:  &inputTemplate,
		OutputTemplate: &outputTemplate,
	}
}
