package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func FixAPIDefinitionInputWithName(name string) graphql.APIDefinitionInput {
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

func FixEventAPIDefinitionInputWithName(name string) graphql.EventDefinitionInput {
	data := graphql.CLOB("data")
	return graphql.EventDefinitionInput{Name: name,
		Spec: &graphql.EventSpecInput{
			Data:   &data,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		}}
}

func FixDocumentInputWithName(t require.TestingT, name string) graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       name,
		Description: "Detailed description of project",
		Format:      graphql.DocumentFormatMarkdown,
		DisplayName: "display-name",
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "kyma-project.io",
			Mode:   ptr.FetchMode(graphql.FetchModeBundle),
			Filter: ptr.String("/docs/README.md"),
			Auth:   FixBasicAuth(t),
		},
	}
}

func FixAPIDefinitionInBundleRequest(appID, bndlID, apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							apiDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, apiID, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixAddAPIToApplicationRequest(appID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinitionToApplication(appID: "%s", in: %s) {
				%s
			}
		}
		`, appID, APIInputGQL, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixUpdateAPIToApplicationRequest(id, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateAPIDefinitionForApplication(id: "%s", in: %s) {
				%s
			}
		}
		`, id, APIInputGQL, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixAPIForApplicationWithDefaultPaginationRequest(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
		result: apisForApplication(appID: %s) {
				%s
			}
		}
		`, appID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForAPIDefinition())))
}
