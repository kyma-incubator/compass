package director

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/director/oauth"
	gql "github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
)

const (
	authorizationHeaderKey = "Authorization"
	tenantHeaderKey        = "Tenant"
)

type Client interface {
	GetRuntime(id string) (graphql.RuntimeExt, error)
}

type client struct {
	graphQLClient gql.Client
	oauthClient   oauth.Client
	queryProvider queryProvider
	token         oauth.Token
	tenant        string
	log           logrus.FieldLogger
}

type GetRuntimeResponse struct {
	Result *graphql.RuntimeExt `json:"result"`
}

func NewDirectorClient(oauthClient oauth.Client, gqlClient gql.Client, tenant string, log logrus.FieldLogger) Client {
	return &client{
		graphQLClient: gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		token:         oauth.Token{},
		tenant:        tenant,
		log:           log,
	}
}

func (dc *client) GetRuntime(id string) (graphql.RuntimeExt, error) {
	getRuntimeQuery := dc.queryProvider.getRuntimeQuery(id)
	var response GetRuntimeResponse
	if err := dc.executeDirectorGraphQLCall(getRuntimeQuery, dc.tenant, &response); err != nil {
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

func (dc *client) executeDirectorGraphQLCall(directorQuery string, tenant string, response interface{}) error {
	if dc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := dc.getToken(); err != nil {
			return err
		}
	}

	req := gcli.NewRequest(directorQuery)
	req.Header.Set(authorizationHeaderKey, fmt.Sprintf("Bearer %s", dc.token.AccessToken))
	req.Header.Set(tenantHeaderKey, tenant)

	if err := dc.graphQLClient.ExecuteRequest(req, response); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to execute GraphQL query with Director"))
	}

	return nil
}

func (dc *client) getToken() error {
	token, err := dc.oauthClient.GetAuthorizationToken()
	if err != nil {
		return errors.Wrap(err, "Error while obtaining token")
	}
	dc.token = token

	return nil
}
