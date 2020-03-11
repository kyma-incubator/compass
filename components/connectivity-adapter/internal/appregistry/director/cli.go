package director

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	ForApplication(ctx ...graphqlizer.FieldCtx) string
	ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string
	ForEventDefinition() string
	ForLabel() string
	Page(item string) string
}

const nameKey = "name"

type gqlCreateApplicationResponse struct {
	Result graphql.Application `json:"result"`
}

type gqlGetApplicationResponse struct {
	Result *graphql.ApplicationExt `json:"result"`
}

type directorClient struct {
	cli               gqlcli.GraphQLClient
	graphqlizer       GraphQLizer
	gqlFieldsProvider GqlFieldsProvider
}

func NewClient(cli gqlcli.GraphQLClient, graphqlizer GraphQLizer, gqlFieldsProvider GqlFieldsProvider) *directorClient {
	return &directorClient{cli: cli, graphqlizer: graphqlizer, gqlFieldsProvider: gqlFieldsProvider}
}

func (r *directorClient) SetApplicationLabel(appID string, label graphql.LabelInput) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
					%s
				}
			}`,
			appID, label.Key, label.Value, r.gqlFieldsProvider.ForLabel()))

	err := r.cli.Run(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}

	return nil
}

func (r *directorClient) CreateAPIDefinition(appID string, apiDefinitionInput graphql.APIDefinitionInput) (string, error) {
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

func (r *directorClient) CreateEventDefinition(appID string, eventDefinitionInput graphql.EventDefinitionInput) (string, error) {
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

func (r *directorClient) GetApplicationsByNameRequest(appName string) *gcli.Request {
	return gcli.NewRequest(fmt.Sprintf(`query {
			result: applications(filter: {key:"%s", query: "\"%s\""}) {
					%s
			}
	}`, nameKey, appName, r.gqlFieldsProvider.Page(r.gqlFieldsProvider.ForApplication())))
}

func (r *directorClient) DeleteAPIDefinition(apiID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteAPIDefinition(id: "%s") {
			id
		}	
	}`, apiID))

	err := r.cli.Run(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}

func (r *directorClient) DeleteEventDefinition(eventID string) error {
	gqlRequest := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteEventDefinition(id: "%s") {
			id
		}	
	}`, eventID))

	err := r.cli.Run(context.Background(), gqlRequest, nil)
	if err != nil {
		return errors.Wrap(err, "while doing GraphQL request")
	}
	return nil
}
