package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
)

func CreateAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}

func FixAppTemplateInputWithDefaultDistinguishLabel(name, selfRegDistinguishLabelKey, selfRegDistinguishLabelValue string) graphql.ApplicationTemplateInput {
	input := FixApplicationTemplate(name)
	input.Labels[selfRegDistinguishLabelKey] = selfRegDistinguishLabelValue

	return input
}

func FixAppTemplateInputWithDefaultProductLabel(name, productLabelKey string, productLabelValue []string) graphql.ApplicationTemplateInput {
	input := FixApplicationTemplate(name)
	input.Labels[productLabelKey] = productLabelValue

	return input
}

func FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder string) graphql.ApplicationTemplateInput {
	return FixApplicationTemplateWithWebhookInput(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder, nil, graphql.ApplicationStatusConditionInitial)
}

func FixApplicationTemplateWithCompositeLabelWithoutWebhook(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder string) graphql.ApplicationTemplateInput {
	appTemplateInput := FixApplicationTemplateWithWebhookInput(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder, nil, graphql.ApplicationStatusConditionInitial)
	appTemplateInput.Labels = map[string]interface{}{
		"composite": map[string]interface{}{
			"key":  "value",
			"key2": "value2",
		},
	}
	return appTemplateInput
}

func FixApplicationTemplateWithWebhookInput(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder string, webhookInput *graphql.WebhookInput, condition graphql.ApplicationStatusCondition) graphql.ApplicationTemplateInput {
	var webhooks []*graphql.WebhookInput = nil
	if webhookInput != nil {
		webhooks = []*graphql.WebhookInput{webhookInput}
	}
	return graphql.ApplicationTemplateInput{
		Name:        applicationType,
		Description: &applicationType,
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:            fmt.Sprintf("{{%s}}", namePlaceholder),
			ProviderName:    str.Ptr("compass"),
			Description:     ptr.String(fmt.Sprintf("test {{%s}}", displayNamePlaceholder)),
			LocalTenantID:   &localTenantID,
			StatusCondition: &condition,
			Labels: graphql.Labels{
				"applicationType": applicationType,
				"region":          region,
				"displayName":     fmt.Sprintf("{{%s}}", displayNamePlaceholder),
			},
			Webhooks: webhooks,
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name: namePlaceholder,
			},
			{
				Name: displayNamePlaceholder,
			},
		},
		Labels:               map[string]interface{}{},
		ApplicationNamespace: &namespace,
		AccessLevel:          graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func FixApplicationFromTemplateInput(applicationType, namePlaceholder, namePlaceholderValue, displayNamePlaceholder, displayNamePlaceholderValue string) graphql.ApplicationFromTemplateInput {
	return graphql.ApplicationFromTemplateInput{
		TemplateName: applicationType,
		Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       namePlaceholderValue,
			},
			{
				Placeholder: displayNamePlaceholder,
				Value:       displayNamePlaceholderValue,
			},
		},
	}
}
