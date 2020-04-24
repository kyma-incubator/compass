package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	gql "github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"
)

const (
	AuthorizationHeader = "Authorization"
	TenantHeader        = "Tenant"
)

type Client interface {
	GetRuntime(id, tenant string) (graphql.RuntimeExt, error)
}

type client struct {
	graphQLClient *gql.Client
	queryProvider queryProvider
	tenant        string
}

func NewClient(endpoint, tenant string, queryLogging bool) Client {
	return &client{
		tenant:        tenant,
		graphQLClient: gql.NewGraphQLClient(endpoint, true, queryLogging),
		queryProvider: queryProvider{},
	}
}

type GetRuntimeResponse struct {
	Result *graphql.RuntimeExt `json:"result"`
}

func (dc *client) GetRuntime(id, tenant string) (graphql.RuntimeExt, error) {
	getRuntimeQuery := dc.queryProvider.getRuntimeQuery(id)
	var response GetRuntimeResponse
	err := dc.executeDirectorGraphQLCall(getRuntimeQuery, tenant, &response)
	if err != nil {
		return graphql.RuntimeExt{}, errors.Wrap(err, fmt.Sprintf("Failed to get runtime %s from Director", id))
	}
	if response.Result == nil {
		return graphql.RuntimeExt{}, errors.Errorf("Failed to get runtime %s get Director: received nil response.", id)
	}
	if response.Result.ID != id {
		return graphql.RuntimeExt{}, errors.Errorf("Failed to get correctly runtime %s in Director: Received wrong Runtime in the response", id)
	}
	return *response.Result, nil
}



func (dc *client) executeDirectorGraphQLCall(query, tenant string, response interface{}) error {
	if dc.token.EmptyOrExpired() {
		if err := dc.getToken(); err != nil {
			return err
		}
	}

	req := gcli.NewRequest(query)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", dc.token.AccessToken))
	req.Header.Set(TenantHeader, tenant)

	err := dc.graphQLClient.ExecuteRequest(req, response)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to execute GraphQL query with Director"))
	}

	return nil
}


func (dc *client) getToken() error {
	token, err := dc.oauthClient.GetAuthorizationToken()
	if err != nil {
		return errors.Wrap(err, "Error while obtaining token")
	}

	if token.EmptyOrExpired() {
		return errors.New("Obtained empty or expired token")
	}

	dc.token = token
	return nil
}
