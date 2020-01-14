package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
	TenantHeader        = "Tenant"
)

//go:generate mockery -name=DirectorClient
type DirectorClient interface {
	CreateRuntime(config *gqlschema.RuntimeInput, tenant string) (string, error)
	DeleteRuntime(id, tenant string) error
	GetConnectionToken(id, tenant string) (graphql.OneTimeToken, error)
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
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

func (cc *directorClient) CreateRuntime(config *gqlschema.RuntimeInput, tenant string) (string, error) {
	log.Infof("Registering Runtime on Director service")

	if config == nil {
		return "", errors.New("Cannot register runtime in Director: missing Runtime config")
	}

	if cc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := cc.getToken(); err != nil {
			return "", err
		}
	}

	runtimeInput, err := cc.graphqlizer.RuntimeInputToGraphQL(*config)
	if err != nil {
		log.Infof("Failed to create graphQLized Runtime input")
		return "", err
	}

	runtimeQuery := cc.queryProvider.createRuntimeMutation(runtimeInput)

	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantHeader, tenant)

	var response CreateRuntimeResponse
	err = cc.gqlClient.Do(req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to register runtime in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", errors.Errorf("Failed to register runtime in Director: Received nil response.")
	}

	log.Infof("Successfully registered Runtime %s in Director for tenant %s", config.Name, tenant)

	return response.Result.ID, nil
}

func (cc *directorClient) DeleteRuntime(id, tenant string) error {
	if cc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := cc.getToken(); err != nil {
			return err
		}
	}

	runtimeQuery := cc.queryProvider.deleteRuntimeMutation(id)
	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantHeader, tenant)

	var response DeleteRuntimeResponse
	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to unregister runtime %s in Director", id))
	}
	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.Errorf("Failed to register unregister runtime %s in Director: received nil response.", id)
	}

	if response.Result.ID != id {
		return errors.Errorf("Failed to unregister correctly the runtime %s in Director: Received bad Runtime id in response", id)
	}

	log.Infof("Successfully unregistered Runtime %s in Director for tenant %s", id, tenant)

	return nil
}

func (cc *directorClient) GetConnectionToken(id, tenant string) (graphql.OneTimeToken, error) {
	if cc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := cc.getToken(); err != nil {
			return graphql.OneTimeToken{}, err
		}
	}

	runtimeQuery := cc.queryProvider.requestOneTimeTokeneMutation(id)
	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantHeader, tenant)

	var response OneTimeTokenResponse
	err := cc.gqlClient.Do(req, &response)
	if err != nil {
		return graphql.OneTimeToken{}, errors.Wrap(err, fmt.Sprintf("Failed to get OneTimeToken for Runtime %s in Director", id))
	}

	if response.Result == nil {
		return graphql.OneTimeToken{}, errors.Errorf("Failed to get OneTimeToken for Runtime %s in Director: received nil response.", id)
	}

	log.Infof("Received OneTimeToken for Runtime %s in Director for tenant %s", id, tenant)

	return *response.Result, nil
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
