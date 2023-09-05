package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixCreateFormationTemplateRequest(createFormationTemplateGQLInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: createFormationTemplate(in: %s) {
    					%s
					}
				}`, createFormationTemplateGQLInput, testctx.Tc.GQLFieldsProvider.ForFormationTemplate()))
}

func FixUpdateFormationTemplateRequest(id string, createFormationTemplateGQLInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: updateFormationTemplate(id: "%s", in: %s) {
    					%s
					}
				}`, id, createFormationTemplateGQLInput, testctx.Tc.GQLFieldsProvider.ForFormationTemplate()))
}

func FixDeleteFormationTemplateRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: deleteFormationTemplate(id: "%s") {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForFormationTemplate()))
}

func FixQueryFormationTemplateRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formationTemplate(id: "%s") {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForFormationTemplate()))
}

func FixQueryFormationTemplateWithConstraintsRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formationTemplate(id: "%s") {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForFormationTemplateWithConstraints()))
}

func FixQueryFormationTemplatesRequestWithPageSize(pageSize int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formationTemplates(first:%d, after:"") {
    					%s
					}
				}`, pageSize, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormationTemplate())))
}

func FixQueryFormationTemplatesRequestWithNameAndPageSize(name string, pageSize int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formationTemplatesByName(name:"%s", first:%d, after:"") {
    					%s
					}
				}`, name, pageSize, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormationTemplate())))
}
