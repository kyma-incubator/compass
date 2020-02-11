package director

import (
	"context"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlTools "github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
	"time"
)

type Client interface {
	CreateRuntime(in schema.RuntimeInput) (string, error)
	DeleteRuntime(runtimeID string) error
	CreateApplication(in schema.ApplicationRegisterInput) (schema.ApplicationExt, error)
	DeleteApplication(appID string) error
	SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error
	GetOneTimeTokenUrl(appID string) (string, error)
}

type client struct {
	scopes       []string
	tenant       string
	graphqulizer gqlTools.Graphqlizer
	client       *gcli.Client
}

type ApplicationResponse struct {
	Result schema.ApplicationExt `json:"result"`
}

type RuntimeResponse struct {
	Result schema.Runtime `json:"result"`
}

type SetLabelResponse struct {
	Result schema.Label `json:"result"`
}

type SetDefaultAppEventingResponse struct {
	Result schema.ApplicationEventingConfiguration `json:"result"`
}

type OneTimeTokenResponse struct {
	Result schema.OneTimeTokenForApplication `json:"result"`
}

func NewClient(tenant string, scopes []string) (Client, error) {
	gqlClient, err := getClient(tenant, scopes)
	if err != nil {
		return nil, err
	}

	return client{
		scopes:       scopes,
		tenant:       tenant,
		graphqulizer: gqlTools.Graphqlizer{},
		client:       gqlClient,
	}, nil
}

func (c client) CreateApplication(in schema.ApplicationRegisterInput) (schema.ApplicationExt, error) {

	appGraphql, err := c.graphqulizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return schema.ApplicationExt{}, err
	}

	var result ApplicationResponse
	query := createApplicationQuery(appGraphql)

	err = c.execute(query, &result)
	if err != nil {
		return schema.ApplicationExt{}, err
	}

	return result.Result, nil
}

func (c client) DeleteApplication(appID string) error {

	return nil
}

func (c client) CreateRuntime(in schema.RuntimeInput) (string, error) {

	runtimeGraphQL, err := c.graphqulizer.RuntimeInputToGQL(in)

	var result RuntimeResponse

	query := createRuntimeQuery(runtimeGraphQL)

	err = c.execute(query, &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c client) DeleteRuntime(runtimeID string) error {
	return nil
}

func (c client) SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error {

	{
		query := setEventBaseURLQuery(runtimeID, eventsBaseURL)
		var response SetLabelResponse

		err := c.execute(query, &response)
		if err != nil {
			return err
		}
	}

	{
		query := setDefaultEventingQuery(runtimeID, appID)
		var response SetDefaultAppEventingResponse

		err := c.execute(query, &response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c client) GetOneTimeTokenUrl(appID string) (string, error) {

	query := getOneTimeTokenQuery(appID)

	var response OneTimeTokenResponse

	err := c.execute(query, &response)
	if err != nil {
		return "", err
	}

	return response.Result.LegacyConnectorURL, nil
}

func (c client) execute(query string, res interface{}) error {

	req := gcli.NewRequest(query)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.client.Run(ctx, req, res)

	return err
}

func getClient(tenant string, scopes []string) (*gcli.Client, error) {

	token, err := getToken(tenant, scopes)
	if err != nil {
		return nil, err
	}

	return gqlTools.NewAuthorizedGraphQLClient(token), nil
}

func getToken(tenant string, scopes []string) (string, error) {
	token, err := jwtbuilder.Do(tenant, scopes)
	if err != nil {
		return "", err
	}

	return token, nil
}
