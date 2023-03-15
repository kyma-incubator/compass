package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
)

func FixProviderRuntimeWithWebhookInput(name string, webhookType graphql.WebhookType, mode graphql.WebhookMode, urlTemplate, inputTemplate, outputTemplate string, selfRegKey, selfRegValue string) graphql.RuntimeRegisterInput {
	return graphql.RuntimeRegisterInput{
		Name:        name,
		Description: ptr.String(fmt.Sprintf("%s-description", name)),
		Labels: graphql.Labels{
			selfRegKey: selfRegValue,
		},
		Webhooks: []*graphql.WebhookInput{
			{
				Type: webhookType,
				Auth: &graphql.AuthInput{
					AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
				},
				Mode:           &mode,
				URLTemplate:    &urlTemplate,
				InputTemplate:  &inputTemplate,
				OutputTemplate: &outputTemplate,
			},
		},
	}

}
