package fixtures

import (
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func FixCreateAutomaticScenarioAssignmentRequest(automaticScenarioAssignmentInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createAutomaticScenarioAssignment(in: %s) {
						%s
					}
				}`,
			automaticScenarioAssignmentInput, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func FixAutomaticScenarioAssignmentsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignments {
						%s
					}
				}`,
			testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment())))
}

func FixAutomaticScenarioAssignmentsForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentsForSelector(selector: %s) {
						%s
					}
				}`,
			labelSelectorInput, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func FixAutomaticScenarioAssignmentForScenarioRequest(scenarioName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentForScenario(scenarioName: "%s") {
						%s
					}
				}`,
			scenarioName, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func removeDoubleQuotesFromJSONKeys(in string) string {
	var validRegex = regexp.MustCompile(`"(\w+|\$\w+)"\s*:`)
	return validRegex.ReplaceAllString(in, `$1:`)
}

func FixEventAPIDefinitionInput() graphql.EventDefinitionInput {
	data := graphql.CLOB("data")
	return graphql.EventDefinitionInput{Name: "name",
		Spec: &graphql.EventSpecInput{
			Data:   &data,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		},
	}
}

func FixAPIDefinitionInput() graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{
		Name:      "new-api-name",
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

func FixDocumentInput(t require.TestingT) graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       "Readme",
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
