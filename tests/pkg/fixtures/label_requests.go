package fixtures

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

func FixCreateLabelDefinitionRequest(labelDefinitionInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createLabelDefinition(in: %s) {
						%s
					}
				}`,
			labelDefinitionInputGQL, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixUpdateLabelDefinitionRequest(ldInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: updateLabelDefinition(in: %s) {
						%s
					}
				}`, ldInputGQL, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixSetApplicationLabelRequest(appID, labelKey string, labelValue interface{}) *gcli.Request {
	jsonValue, err := json.Marshal(labelValue)
	if err != nil {
		panic(errors.New("label value can not be marshalled"))
	}
	value := removeDoubleQuotesFromJSONKeys(string(jsonValue))

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
					%s
				}
			}`,
			appID, labelKey, value, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixSetRuntimeLabelRequest(runtimeID, labelKey string, labelValue interface{}) *gcli.Request {
	jsonValue, err := json.Marshal(labelValue)
	if err != nil {
		panic(errors.New("label value can not be marshalled"))
	}
	value := removeDoubleQuotesFromJSONKeys(string(jsonValue))

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: %s) {
						%s
					}
				}`, runtimeID, labelKey, value, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixLabelDefinitionRequest(labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: labelDefinition(key: "%s") {
						%s
					}
				}`,
			labelKey, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixLabelDefinitionsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result:	labelDefinitions() {
					key
					schema
				}
			}`))
}

func FixDeleteRuntimeLabelRequest(runtimeID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
					%s
				}
			}`, runtimeID, labelKey, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixDeleteApplicationLabelRequest(applicationID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
					%s
				}
			}`, applicationID, labelKey, testctx.Tc.GQLFieldsProvider.ForLabel()))
}
