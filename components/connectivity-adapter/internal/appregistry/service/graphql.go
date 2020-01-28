package service

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/pkg/errors"

	gcli "github.com/machinebox/graphql"
)

//go:generate mockery -name=GraphQLizer -output=automock -outpkg=automock -case=underscore
type GraphQLizer interface {
	ApplicationRegisterInputToGQL(in graphql.ApplicationRegisterInput) (string, error)
}

const nameKey = "name"


//go:generate mockery -name=GqlFieldsProvider -output=automock -outpkg=automock -case=underscore
type GqlFieldsProvider interface {
	ForApplication(ctx ...gql.FieldCtx) string
	Page(item string) string
}

type gqlCreateApplicationResponse struct {
	Result graphql.Application `json:"result"`
}

type gqlGetApplicationResponse struct {
	Result *graphql.ApplicationExt `json:"result"`
}

type gqlRequestBuilder struct {
	graphqlizer       GraphQLizer
	gqlFieldsProvider GqlFieldsProvider
}

func NewGqlRequestBuilder(graphqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *gqlRequestBuilder {
	return &gqlRequestBuilder{graphqlizer: graphqlizer, gqlFieldsProvider: gqlFieldsProvider}
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


func (b *gqlRequestBuilder) GetApplicationsByName(appName string) *gcli.Request {
	return gcli.NewRequest(fmt.Sprintf(`query {
			result: applications(filter: {key:"%s", query: "\"%s\""}) {
					%s
			}
	}`, nameKey, appName, b.gqlFieldsProvider.Page(b.gqlFieldsProvider.ForApplication())))
}
