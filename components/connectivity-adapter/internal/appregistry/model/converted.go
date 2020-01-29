package model

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type GraphQLServiceDetailsInput struct {
	ID    string
	API   *graphql.APIDefinitionInput
	Event *graphql.EventDefinitionInput
}

type GraphQLServiceDetails struct {
	ID    string
	API   *graphql.APIDefinitionExt
	Event *graphql.EventAPIDefinitionExt
}
