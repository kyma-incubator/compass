package fixtures

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
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

func FixDocumentInputWithName(t *testing.T, name string) graphql.DocumentInput {
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
