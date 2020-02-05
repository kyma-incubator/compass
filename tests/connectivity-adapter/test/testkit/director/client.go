package director

import (
	"context"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlTools "github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
	gcli "github.com/machinebox/graphql"
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
}

func NewClient(tenant string, scopes []string) Client {
	return client{
		scopes:       scopes,
		tenant:       tenant,
		graphqulizer: gqlTools.Graphqlizer{},
	}
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

func (c client) CreateApplication(in schema.ApplicationRegisterInput) (schema.ApplicationExt, error) {

	client, err := c.getClient()
	if err != nil {
		return schema.ApplicationExt{}, err
	}

	appGraphql, err := c.graphqulizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return schema.ApplicationExt{}, err
	}

	var result ApplicationResponse
	query := createApplicationMutation(appGraphql)

	err = c.execute(client, query, &result)
	if err != nil {
		return schema.ApplicationExt{}, err
	}

	return result.Result, nil
}

func (c client) DeleteApplication(appID string) error {

	return nil
}

func (c client) CreateRuntime(in schema.RuntimeInput) (string, error) {

	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	runtimeGraphQL, err := c.graphqulizer.RuntimeInputToGQL(in)

	var result RuntimeResponse

	query := createRuntimeMutation(runtimeGraphQL)

	err = c.execute(client, query, &result)
	if err != nil {
		return "", err
	}

	return result.Result.ID, nil
}

func (c client) DeleteRuntime(runtimeID string) error {
	return nil
}

func (c client) getClient() (*gcli.Client, error) {

	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	return gqlTools.NewAuthorizedGraphQLClient(token), nil
}

func (c client) getToken() (string, error) {
	token, err := jwtbuilder.Do(c.tenant, c.scopes)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (c client) SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error {

	client, err := c.getClient()
	if err != nil {
		return err
	}

	query := setEventBaseURLMutation(runtimeID, eventsBaseURL)

	{
		var response SetLabelResponse

		err = c.execute(client, query, &response)
		if err != nil {
			return err
		}
	}

	{
		query := setDefaultEventingForApplication(runtimeID, appID)

		var response SetDefaultAppEventingResponse

		err = c.execute(client, query, &response)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c client) GetOneTimeTokenUrl(appID string) (string, error) {
	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	query := getOneTimeTokenForApplication(appID)

	var response OneTimeTokenResponse

	err = c.execute(client, query, &response)
	if err != nil {
		return "", err
	}

	return response.Result.LegacyConnectorURL, nil
}

func (c client) execute(client *gcli.Client, query string, res interface{}) error {

	req := gcli.NewRequest(query)

	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	err := client.Run(context.Background(), req, res)

	return err
}
