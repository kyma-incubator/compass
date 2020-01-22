package service

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/pkg/errors"

	gcli "github.com/machinebox/graphql"
)

type gqlRequestBuilder struct {
	graphqlizer       gql.Graphqlizer
	gqlFieldsProvider gql.GqlFieldsProvider
}

func NewGqlRequestBuilder() *gqlRequestBuilder {
	return &gqlRequestBuilder{graphqlizer: gql.Graphqlizer{}, gqlFieldsProvider: gql.GqlFieldsProvider{}}
}

func (b *gqlRequestBuilder) RegisterApplicationRequest(input graphql.ApplicationRegisterInput) (*gcli.Request, error) {
	appInputGQL, err := b.graphqlizer.ApplicationRegisterInputToGQL(input)
	if err != nil {
		return nil, errors.Wrapf(err, "while constructing input")
	}

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
				id
			}	
		}`,
			appInputGQL)), nil
}

func (b *gqlRequestBuilder) UnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: unregisterApplication(id: "%s") {
			id
		}	
	}`, id))
}

func (b *gqlRequestBuilder) GetApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
			}
		}`, id, b.gqlFieldsProvider.ForApplication()))
}
