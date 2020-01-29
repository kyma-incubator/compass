package service

import (
	"fmt"

	gcli "github.com/machinebox/graphql"
)

// TODO: Remove this file and migrate to GraphQLRequester

const nameKey = "name"

type gqlRequestBuilder struct {
	graphqlizer       GraphQLizer
	gqlFieldsProvider GqlFieldsProvider
}

func NewGqlRequestBuilder(graphqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *gqlRequestBuilder {
	return &gqlRequestBuilder{graphqlizer: graphqlizer, gqlFieldsProvider: gqlFieldsProvider}
}

func (b *gqlRequestBuilder) GetApplicationsByName(appName string) *gcli.Request {
	return gcli.NewRequest(fmt.Sprintf(`query {
			result: applications(filter: {key:"%s", query: "\"%s\""}) {
					%s
			}
	}`, nameKey, appName, b.gqlFieldsProvider.Page(b.gqlFieldsProvider.ForApplication())))
}
