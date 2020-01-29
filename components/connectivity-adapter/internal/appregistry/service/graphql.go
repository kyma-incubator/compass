package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=GraphQLizer -output=automock -outpkg=automock -case=underscore
type GraphQLizer interface {
	ApplicationRegisterInputToGQL(in graphql.ApplicationRegisterInput) (string, error)
	APIDefinitionInputToGQL(in graphql.APIDefinitionInput) (string, error)
	EventDefinitionInputToGQL(in graphql.EventDefinitionInput) (string, error)
}

//go:generate mockery -name=GqlFieldsProvider -output=automock -outpkg=automock -case=underscore
type GqlFieldsProvider interface {
	ForApplication(ctx ...gql.FieldCtx) string
	ForAPIDefinition(ctx ...gql.FieldCtx) string
	ForEventDefinition() string
	ForLabel() string
	Page(item string) string
}

type gqlCreateApplicationResponse struct {
	Result graphql.Application `json:"result"`
}

type gqlGetApplicationResponse struct {
	Result *graphql.ApplicationExt `json:"result"`
}

type gqlRequester struct {
	cli               gqlcli.GraphQLClient
	graphqlizer       GraphQLizer
	gqlFieldsProvider GqlFieldsProvider
}

func NewGqlRequester(cli gqlcli.GraphQLClient, graphqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *gqlRequester {
	return &gqlRequester{cli: cli, graphqlizer: graphqlizer, gqlFieldsProvider: gqlFieldsProvider}
}

func (r *gqlRequester) SetApplicationLabel(appID string, label graphql.LabelInput) error {
	jsonValue, err := json.Marshal(label.Value)
	if err != nil {
		return errors.Wrap(err, "while marshalling JSON value")
	}
	value := strconv.Quote(string(jsonValue))

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
					%s
				}
			}`,
			appID, label.Key, value, r.gqlFieldsProvider.ForLabel()))

	err = r.cli.Run(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}

	return nil
}

func (r *gqlRequester) CreateAPIDefinition(appID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error) {
	inStr, err := r.graphqlizer.APIDefinitionInputToGQL(apiDefinitionInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPIDefinition(applicationID: "%s", in: %s) {
					%s
				}
			}`, appID, inStr, r.gqlFieldsProvider.ForAPIDefinition()))

	var resp struct {
		Result graphql.APIDefinition `json:"result"`
	}

	err = r.cli.Run(context.Background(), gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}

func (r *gqlRequester) CreateEventDefinition(appID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error) {
	inStr, err := r.graphqlizer.EventDefinitionInputToGQL(eventDefinitionInput)
	if err != nil {
		return "", errors.Wrap(err, "while preparing GraphQL input")
	}

	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: addEventDefinition(applicationID: "%s", in: %s) {
						%s	
					}
				}`, appID, inStr, r.gqlFieldsProvider.ForEventDefinition()))

	var resp struct {
		Result graphql.EventDefinition `json:"result"`
	}

	err = r.cli.Run(context.Background(), gqlRequest, &resp)
	if err != nil {
		return "", errors.Wrap(err, "while doing GraphQL request")
	}

	return resp.Result.ID, nil
}
