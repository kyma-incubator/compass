package director

import (
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/oauth"

	"github.com/pkg/errors"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
)

// TODO - align - either use interfaces or Structs

type Client struct {
	tenant        string
	graphQLClient *graphql.Client
	oauthClient   *oauth.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
}

// TODO - add some shared client to handle tokens (some abstraction for Compass in general)

func NewDirectorClient(endpoint, tenant string, oauthClient *oauth.Client, queryLogging bool) *Client {
	return &Client{
		tenant:        tenant,
		graphQLClient: graphql.NewGraphQLClient(endpoint, true, queryLogging),
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
	}
}

func (c *Client) RegisterRuntime(input directorSchema.RuntimeInput) (Runtime, error) {
	runtimeInput, err := c.graphqlizer.RuntimeInputToGQL(input)
	if err != nil {
		return Runtime{}, errors.Wrap(err, "Failed to convert Runtime Input to query")
	}

	query := c.queryProvider.registerRuntime(runtimeInput)
	req, err := c.newRequest(query)
	if err != nil {
		return Runtime{}, err
	}

	var runtime Runtime
	err = c.graphQLClient.ExecuteRequest(req, &runtime, &Runtime{})
	if err != nil {
		return Runtime{}, errors.Wrap(err, "Failed to register Runtime")
	}

	return runtime, nil
}

func (c *Client) DeleteRuntime(runtimeId string) (string, error) {
	query := c.queryProvider.deleteRuntime(runtimeId)
	req, err := c.newRequest(query)
	if err != nil {
		return "", err
	}

	var idResponse IdResponse
	err = c.graphQLClient.ExecuteRequest(req, &idResponse, &idResponse)
	if err != nil {
		return "", errors.Wrap(err, "Failed to delete Runtime")
	}

	return idResponse.Id, nil
}

// TODO - modify this part to not to repeat it
func (c *Client) newRequest(query string) (*gcli.Request, error) {
	accessToken, err := c.oauthClient.GetAccessToken()
	if err != nil {
		return nil, errors.Wrap(err, "Error while getting Access Token")
	}

	req := gcli.NewRequest(query)
	req.Header.Set(compass.TenantHeader, c.tenant)
	req.Header.Set(compass.AuthorizationHeader, accessToken.AccessToken)

	return req, nil
}

type IdResponse struct {
	Id string `json:"id"`
}
