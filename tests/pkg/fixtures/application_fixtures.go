package fixtures

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
)

func FixApplicationTemplateWithWebhookNotifications(applicationType, localTenantID, region, namespace string, webhookType graphql.WebhookType, mode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string) graphql.ApplicationTemplateInput {
	webhookInput := &graphql.WebhookInput{
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		Mode:           &mode,
		URLTemplate:    &urlTemplate,
		InputTemplate:  &inputTemplate,
		OutputTemplate: &outputTemplate,
	}
	return fixApplicationTemplateWebhook(applicationType, localTenantID, region, namespace, webhookInput)
}

func FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, region, namespace string) graphql.ApplicationTemplateInput {
	return fixApplicationTemplateWebhook(applicationType, localTenantID, region, namespace, nil)
}

func fixApplicationTemplateWebhook(applicationType, localTenantID, region, namespace string, webhookInput *graphql.WebhookInput) graphql.ApplicationTemplateInput {
	var webhooks []*graphql.WebhookInput = nil
	if webhookInput != nil {
		webhooks = make([]*graphql.WebhookInput, 0, 1)
		webhooks[0] = webhookInput
	}
	return graphql.ApplicationTemplateInput{
		Name:        applicationType,
		Description: &applicationType,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:          "{{name}}",
			ProviderName:  str.Ptr("compass"),
			Description:   ptr.String("test {{display-name}}"),
			LocalTenantID: &localTenantID,
			Labels: graphql.Labels{
				"applicationType": applicationType,
				"region":          region,
			},
			Webhooks: webhooks,
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name: "name",
			},
			{
				Name: "display-name",
			},
		},
		ApplicationNamespace: &namespace,
		AccessLevel:          graphql.ApplicationTemplateAccessLevelGlobal,
	}
}
