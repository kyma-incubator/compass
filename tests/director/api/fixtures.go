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
	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
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

func fixPackageCreateInputWithRelatedObjects(name string) graphql.PackageCreateInput {
	desc := "Foo bar"
	return graphql.PackageCreateInput{
		Name:        name,
		Description: &desc,
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
				DefaultAuth: fixBasicAuth(),
				Version:     fixDepracatedVersion1(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(`{"openapi":"3.0.2"}`),
				},
			},
			{
				Name:      "reviews-v1",
				TargetURL: "http://mywordpress.com/reviews",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatJSON,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/apis",
						Mode:   ptr.FetchMode(graphql.FetchModePackage),
						Filter: ptr.String("odata.json"),
						Auth:   fixBasicAuth(),
					},
				},
				DefaultAuth: &graphql.AuthInput{
					Credential: fixBasicCredential(),
					RequestAuth: &graphql.CredentialRequestAuthInput{
						Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
							Credential:       fixOAuthCredential(),
							TokenEndpointURL: "http://token.URL",
						},
					},
				},
			},
			{
				Name:      "xml",
				TargetURL: "http://mywordpress.com/xml",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatXML,
					Data:   ptr.CLOB("odata"),
				},
			},
		},
		EventDefinitions: []*graphql.EventDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("comments events"),
				Version:     fixDepracatedVersion1(),
				Group:       ptr.String("comments"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(`{"asyncapi":"1.2.0"}`),
				},
			},
			{
				Name:        "reviews-v1",
				Description: ptr.String("review events"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/events",
						Mode:   ptr.FetchMode(graphql.FetchModePackage),
						Filter: ptr.String("async.json"),
						Auth:   fixOauthAuth(),
					},
				},
			},
		},
		Documents: []*graphql.DocumentInput{
			{
				Title:       "Readme",
				Description: "Detailed description of project",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				FetchRequest: &graphql.FetchRequestInput{
					URL:    "kyma-project.io",
					Mode:   ptr.FetchMode(graphql.FetchModePackage),
					Filter: ptr.String("/docs/README.md"),
					Auth:   fixBasicAuth(),
				},
			},
			{
				Title:       "Troubleshooting",
				Description: "Troubleshooting description",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				Data:        ptr.CLOB("No problems, everything works on my machine"),
			},
		},
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

func fixApplicationRegisterInputWithPackages() graphql.ApplicationRegisterInput {
	pkg1 := fixPackageCreateInputWithRelatedObjects("foo")
	pkg2 := fixPackageCreateInputWithRelatedObjects("bar")
	return graphql.ApplicationRegisterInput{
		Name:         "create-application-with-documents",
		ProviderName: ptr.String("compass"),
		Packages: []*graphql.PackageCreateInput{
			&pkg1, &pkg2,
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
}
