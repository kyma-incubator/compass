package director

import (
	"fmt"
	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
	TenantKey           = "tenant"
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
	tenant        string
	token         oauth.Token
	oauthClient   oauth.Client
}

func NewDirectorClient(gqlClient gql.Client, oauthClient oauth.Client, tenant string) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
		tenant:        tenant,
		token:         oauth.Token{},
	}
}

func (cc *directorClient) CreateRuntime(config *gqlschema.RuntimeInput) (string, error) {
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

	var response CreateRuntimeResponse

	graphQLized, err := cc.graphqlizer.RuntimeInputToGraphQL(*config)

	if err != nil {
		log.Infof("Failed to create graphQLized Runtime input")
		return "", err
	}

	runtimeQuery := cc.queryProvider.createRuntimeMutation(graphQLized)

	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantKey, cc.tenant)

	err = cc.gqlClient.Do(req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to register runtime in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", errors.Errorf("Failed to register runtime in Director: Received nil response.")
	}

	log.Infof("Successfully sent GraphQL mutation to Director to register runtime %s for tenant %s", config.Name, cc.tenant)

	return response.Result.ID, nil
}

func (cc *directorClient) DeleteRuntime(id string) error {
	if cc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := cc.getToken(); err != nil {
			return err
		}
	}

	var response DeleteRuntimeResponse

	runtimeQuery := cc.queryProvider.deleteRuntimeMutation(id)
	req := gcli.NewRequest(runtimeQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantKey, cc.tenant)

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

	log.Infof("Successfully sent GraphQL mutation to Director to delete runtime %s for tenant %s", id, cc.tenant)

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
