package director

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixBasicAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: fixBasicCredential(),
		AdditionalHeaders: &graphql.HttpHeaders{
			"headerA": []string{"ha1", "ha2"},
			"headerB": []string{"hb1", "hb2"},
		},
		AdditionalQueryParams: &graphql.QueryParams{
			"qA": []string{"qa1", "qa2"},
			"qB": []string{"qb1", "qb2"},
		},
	}
}

func fixOauthAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: fixOAuthCredential(),
	}
}

func fixBasicCredential() *graphql.CredentialDataInput {
	return &graphql.CredentialDataInput{
		Basic: &graphql.BasicCredentialDataInput{
			Username: "admin",
			Password: "secret",
		}}
}

func fixOAuthCredential() *graphql.CredentialDataInput {
	return &graphql.CredentialDataInput{
		Oauth: &graphql.OAuthCredentialDataInput{
			URL:          "url.net",
			ClientSecret: "grazynasecret",
			ClientID:     "clientid",
		}}
}

func fixDepracatedVersion1() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           "v1",
		Deprecated:      ptr.Bool(true),
		ForRemoval:      ptr.Bool(true),
		DeprecatedSince: ptr.String("v5"),
	}
}

func fixRuntimeInput(placeholder string) graphql.RuntimeInput {
	return graphql.RuntimeInput{
		Name:        placeholder,
		Description: ptr.String(fmt.Sprintf("%s-description", placeholder)),
		Labels:      &graphql.Labels{"placeholder": []interface{}{"placeholder"}},
	}
}

func fixApplicationTemplate(name string) graphql.ApplicationTemplateInput {
	appTemplateDesc := "app-template-desc"
	placeholderDesc := "new-placeholder-desc"
	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &appTemplateDesc,
		ApplicationInput: &graphql.ApplicationCreateInput{
			Name:        "app",
			Description: ptr.String("test {{new-placeholder}}"),
			Labels: &graphql.Labels{
				"a": []string{"b", "c"},
				"d": []string{"e", "f"},
			},
			Webhooks: []*graphql.WebhookInput{{
				Type: graphql.ApplicationWebhookTypeConfigurationChanged,
				URL:  "http://url.com",
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
