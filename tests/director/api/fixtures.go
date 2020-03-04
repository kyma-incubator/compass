package api

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

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
	providerName := "compass-tests"
	displayName := fmt.Sprintf("Display %s", name)
	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
		DisplayName: &displayName,
		Description: &appTemplateDesc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:         "app",
			ProviderName: &providerName,
			Description:  ptr.String("test {{new-placeholder}}"),
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

func fixPackageCreateInput(name string) graphql.PackageCreateInput {
	return graphql.PackageCreateInput{
		Name: name,
	}
}

func fixPackageCreateInputWithDefaultAuth(name string, authInput *graphql.AuthInput) graphql.PackageCreateInput {
	return graphql.PackageCreateInput{
		Name:                name,
		DefaultInstanceAuth: authInput,
	}
}

func fixPackageUpdateInput(name string) graphql.PackageUpdateInput {
	return graphql.PackageUpdateInput{
		Name: name,
	}
}

func fixAPIDefinitionInputWithName(name string) graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{
		Name:      name,
		TargetURL: "https://target.url",
		Spec: &graphql.APISpecInput{
			Format: graphql.SpecFormatJSON,
			Type:   graphql.APISpecTypeOpenAPI,
			FetchRequest: &graphql.FetchRequestInput{
				URL: "https://foo.bar",
			},
		},
	}
}

func fixEventAPIDefinitionInputWithName(name string) graphql.EventDefinitionInput {
	data := graphql.CLOB("data")
	return graphql.EventDefinitionInput{Name: name,
		Spec: &graphql.EventSpecInput{
			Data:   &data,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		}}
}

func fixDocumentInputWithName(name string) graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       name,
		Description: "Detailed description of project",
		Format:      graphql.DocumentFormatMarkdown,
		DisplayName: "display-name",
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "kyma-project.io",
			Mode:   ptr.FetchMode(graphql.FetchModePackage),
			Filter: ptr.String("/docs/README.md"),
			Auth:   fixBasicAuth(),
		},
	}
}

func fixPackageInstanceAuthRequestInput(ctx, inputParams *graphql.JSON) graphql.PackageInstanceAuthRequestInput {
	return graphql.PackageInstanceAuthRequestInput{
		Context:     ctx,
		InputParams: inputParams,
	}
}

func fixPackageInstanceAuthSetInputSucceeded(auth *graphql.AuthInput) graphql.PackageInstanceAuthSetInput {
	return graphql.PackageInstanceAuthSetInput{
		Auth: auth,
	}
}
