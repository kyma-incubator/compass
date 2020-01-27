package model

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type ConvertedServiceDetails struct {
	ID string
	API   *graphql.APIDefinitionInput
	Event *graphql.EventDefinitionInput
}
