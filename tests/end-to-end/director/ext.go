package director

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

// ApplicationExt contains Application with all it dependant objects
type ApplicationExt struct {
	graphql.Application
	Labels    graphql.Labels
	Webhooks  []graphql.Webhook
	Apis      graphql.APIDefinitionPage
	EventAPIs graphql.EventAPIDefinitionPage
	Documents graphql.DocumentPage
}

type RuntimeExt struct {
	graphql.Runtime
	Labels graphql.Labels
}

type ApplicationPageExt struct {
	Data       []*ApplicationExt `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type RuntimePageExt struct {
	Data       []*RuntimeExt     `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}
