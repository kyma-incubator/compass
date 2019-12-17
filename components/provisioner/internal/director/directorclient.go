package director

import (
	"fmt"
	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	AuthorizationHeader = "Authorization"
)

//go:generate mockery -name=DirectorClient
type DirectorClient interface {
	CreateRuntime(config *gqlschema.RuntimeInput) (string, error)
	DeleteRuntime(id string) error
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
	runtimeConfig string
	token         oauth.Token
	oauthClient   oauth.Client
}

func NewDirectorClient(gqlClient gql.Client, oauthClient oauth.Client) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
		token:         oauth.Token{},
	}
}

func (cc *directorClient) CreateRuntime(config *gqlschema.RuntimeInput) (string, error) {
	if config == nil {
		return "", errors.New("Cannot register register runtime in Director: missing Runtime config")
	}

	if cc.token.EmptyOrExpired() {
		if err := cc.getToken(); err != nil {
			return "", err
		}
	}

	var response CreateRuntimeResponse

	graphQLized, err := cc.graphqlizer.RuntimeInputToGraphQL(*config)

	if err != nil {
		return "", err
	}

	applicationsQuery := cc.queryProvider.createRuntimeMutation(graphQLized)
	req := gcli.NewRequest(applicationsQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))

	err = cc.gqlClient.Do(req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to register runtime in Director")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", errors.Errorf("Failed to register runtime in Director: received nil response.")
	}

	return response.Result.ID, nil
}

func (cc *directorClient) DeleteRuntime(id string) error {
	if cc.token.EmptyOrExpired() {
		if err := cc.getToken(); err != nil {
			return err
		}
	}

	var response DeleteRuntimeResponse

	applicationsQuery := cc.queryProvider.deleteRuntimeMutation(id)
	req := gcli.NewRequest(applicationsQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))

	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return errors.Wrap(err, "Failed to unregister runtime %s in Director")
	}
	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.Errorf("Failed to register unregister runtime %s in Director: received nil response.", id)
	}

	if response.Result.ID != id {
		return errors.Errorf("Failed to unregister correctly the runtime %s in Director: Received bad Runtime id in response", id)
	}

	return nil
}

func (cc *directorClient) getToken() error {
	token, err := cc.oauthClient.GetAuthorizationToken()

	if err != nil {
		return errors.Wrap(err, "Error while obtaining token")
	}

	if token.EmptyOrExpired() {
		return errors.New("Obtained empty or expired token")
	}

	cc.token = token
	return nil
}
