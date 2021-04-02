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
	placeholderDesc := "new-placeholder-desc"
	providerName := "compass-tests"
	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &appTemplateDesc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:         "app",
			ProviderName: &providerName,
			Description:  ptr.String("test {{new-placeholder}}"),
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
				Name:        "new-placeholder",
				Description: &placeholderDesc,
			},
		},
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
