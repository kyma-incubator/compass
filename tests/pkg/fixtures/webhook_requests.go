package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixAddWebhookRequest(applicationID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`,
			applicationID, webhookInGQL, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}

func FixDeleteWebhookRequest(webhookID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteWebhook(webhookID: "%s") {
				%s
			}
		}`, webhookID, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}

func FixUpdateWebhookRequest(webhookID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`,
			webhookID, webhookInGQL, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}

func FixAddWebhookToTemplateRequest(applicationTemplateID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addWebhook(applicationTemplateID: "%s", in: %s) {
					%s
				}
			}`,
			applicationTemplateID, webhookInGQL, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}
