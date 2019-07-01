package end_to_end

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

// ApplicationExt contains Application with all it dependant objects
type ApplicationExt struct {
	graphql.Application
	Webhooks  []graphql.ApplicationWebhook
	Apis      graphql.APIDefinitionPage
	EventAPIs graphql.EventAPIDefinitionPage
	Documents graphql.DocumentPage
}
