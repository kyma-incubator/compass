package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixApplicationTemplate(name string) graphql.ApplicationTemplateInput {
	appTemplateDesc := "app-template-desc"
	placeholderDisplayName := "new-placeholder-display-name"
	appInputName := "app"
	providerName := "compass-tests"
	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &appTemplateDesc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:         "{{name}}",
			ProviderName: &providerName,
			Description:  ptr.String("test {{display-name}}"),
			Labels: graphql.Labels{
				"a": []string{"b", "c"},
				"d": []string{"e", "f"},
			},
			Webhooks: []*graphql.WebhookInput{{
				Type: graphql.WebhookTypeConfigurationChanged,
				URL:  ptr.String("http://url.com"),
			}},
			HealthCheckURL: ptr.String("http://url.valid"),
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name:        "name",
				Description: &appInputName,
			},
			{
				Name:        "display-name",
				Description: &placeholderDisplayName,
			},
		},
		Labels: graphql.Labels{
			"test": []interface{}{"test"},
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
	}
	return appTemplateInput
}

func FixApplicationTemplateWithoutWebhooks(name string) graphql.ApplicationTemplateInput {
	appTemplateDesc := "app-template-without-webhook-desc"
	placeholderDisplayName := "placeholder-display-name"
	appInputName := "app"
	providerName := "compass-tests"
	appNamespace := "compass.ns.test"

	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &appTemplateDesc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:         "{{name}}",
			ProviderName: &providerName,
			Description:  ptr.String("test {{display-name}}"),
			Labels: graphql.Labels{
				"a": []string{"b", "c"},
				"d": []string{"e", "f"},
			},
			HealthCheckURL: ptr.String("http://url.valid"),
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name:        "name",
				Description: &appInputName,
			},
			{
				Name:        "display-name",
				Description: &placeholderDisplayName,
			},
		},
		Labels: graphql.Labels{
			"test": []interface{}{"test"},
		},
		ApplicationNamespace: &appNamespace,
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
	}
	return appTemplateInput
}

func FixApplicationTemplateWithWebhook(name string) graphql.ApplicationTemplateInput {
	appTemplate := FixApplicationTemplate(name)
	appTemplate.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
		Auth: &graphql.AuthInput{
			Credential: &graphql.CredentialDataInput{
				Basic: &graphql.BasicCredentialDataInput{
					Username: "username",
					Password: "password",
				},
			},
		},
	}}
	return appTemplate
}

func FixApplicationTemplateWithORDWebhook(name, webhookURL string) graphql.ApplicationTemplateInput {
	appTemplate := FixApplicationTemplate(name)
	appTemplate.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeOpenResourceDiscovery,
		URL:  ptr.String(webhookURL),
	}}
	return appTemplate
}

func FixCreateApplicationTemplateRequest(applicationTemplateInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplicationTemplate(in: %s) {
					%s
				}
			}`,
			applicationTemplateInGQL, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixDeleteApplicationTemplateRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationTemplate(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixApplicationTemplateRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applicationTemplate(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixUpdateApplicationTemplateRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplicationTemplate(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}
