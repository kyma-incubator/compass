package tests

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func fixBasicAuth(t *testing.T) *graphql.AuthInput {
	additionalHeaders, err := graphql.NewHttpHeadersSerialized(map[string][]string{
		"header-A": []string{"ha1", "ha2"},
		"header-B": []string{"hb1", "hb2"},
	})
	require.NoError(t, err)

	additionalQueryParams, err := graphql.NewQueryParamsSerialized(map[string][]string{
		"qA": []string{"qa1", "qa2"},
		"qB": []string{"qb1", "qb2"},
	})
	require.NoError(t, err)

	return &graphql.AuthInput{
		Credential:                      fixBasicCredential(),
		AdditionalHeadersSerialized:     &additionalHeaders,
		AdditionalQueryParamsSerialized: &additionalQueryParams,
	}
}

func fixBasicAuthLegacy() *graphql.AuthInput {
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
		ForRemoval:      ptr.Bool(false),
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

func fixBundleCreateInput(name string) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name: name,
	}
}

func fixBundleCreateInputWithRelatedObjects(t *testing.T, name string) graphql.BundleCreateInput {
	desc := "Foo bar"
	return graphql.BundleCreateInput{
		Name:        name,
		Description: &desc,
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
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
						Mode:   ptr.FetchMode(graphql.FetchModeBundle),
						Filter: ptr.String("odata.json"),
						Auth:   fixBasicAuth(t),
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
						Mode:   ptr.FetchMode(graphql.FetchModeBundle),
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
					Mode:   ptr.FetchMode(graphql.FetchModeBundle),
					Filter: ptr.String("/docs/README.md"),
					Auth:   fixBasicAuth(t),
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

func fixBundleCreateInputWithDefaultAuth(name string, authInput *graphql.AuthInput) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name:                name,
		DefaultInstanceAuth: authInput,
	}
}

func fixBundleUpdateInput(name string) graphql.BundleUpdateInput {
	return graphql.BundleUpdateInput{
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

func fixDocumentInputWithName(t *testing.T, name string) graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       name,
		Description: "Detailed description of project",
		Format:      graphql.DocumentFormatMarkdown,
		DisplayName: "display-name",
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "kyma-project.io",
			Mode:   ptr.FetchMode(graphql.FetchModeBundle),
			Filter: ptr.String("/docs/README.md"),
			Auth:   fixBasicAuth(t),
		},
	}
}

func fixBundleInstanceAuthRequestInput(ctx, inputParams *graphql.JSON) graphql.BundleInstanceAuthRequestInput {
	return graphql.BundleInstanceAuthRequestInput{
		Context:     ctx,
		InputParams: inputParams,
	}
}

func fixBundleInstanceAuthSetInputSucceeded(auth *graphql.AuthInput) graphql.BundleInstanceAuthSetInput {
	return graphql.BundleInstanceAuthSetInput{
		Auth: auth,
	}
}

func fixApplicationRegisterInputWithBundles(t *testing.T) graphql.ApplicationRegisterInput {
	bndl1 := fixBundleCreateInputWithRelatedObjects(t, "foo")
	bndl2 := fixBundleCreateInputWithRelatedObjects(t, "bar")
	return graphql.ApplicationRegisterInput{
		Name:         "create-application-with-documents",
		ProviderName: ptr.String("compass"),
		Bundles: []*graphql.BundleCreateInput{
			&bndl1, &bndl2,
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
}
