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
	CreateApplication(in schema.ApplicationRegisterInput) (schema.Application, error)
	DeleteApplication(appID string) error
	SetDefaultEventing(runtimeID string, appID string, eventsBaseURL string) error
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
	Result schema.Application `json:"result"`
}

type RuntimeResponse struct {
	Result schema.Runtime `json:"result"`
}

func (c client) CreateApplication(in schema.ApplicationRegisterInput) (schema.Application, error) {

	client, err := c.getClient()
	if err != nil {
		return schema.Application{}, err
	}

	appGraphql, err := c.graphqulizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return schema.Application{}, err
	}

	var result ApplicationResponse
	query := createApplicationMutation(appGraphql)

	err = c.execute(client, query, &result)
	if err != nil {
		return schema.Application{}, err
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
	return nil
}

func (c client) execute(client *gcli.Client, query string, res interface{}) error {

	req := gcli.NewRequest(query)

	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	err := client.Run(context.Background(), req, res)

	return err
}
