package director

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

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
	GetRuntime(id, tenant string) (graphql.Runtime, error)
	UpdateRuntime(id string, config *gqlschema.RuntimeInput, tenant string) error
	DeleteRuntime(id, tenant string) error
	SetRuntimeStatusCondition(id string, statusCondition gqlschema.RuntimeStatusCondition, tenant string) error
	GetConnectionToken(id, tenant string) (graphql.OneTimeTokenForRuntimeExt, error)
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer.Graphqlizer
	token         oauth.Token
	oauthClient   oauth.Client
}

func NewDirectorClient(gqlClient gql.Client, oauthClient oauth.Client) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer.Graphqlizer{},
		token:         oauth.Token{},
	}
}

func (cc *directorClient) CreateRuntime(config *gqlschema.RuntimeInput, tenant string) (string, error) {
	log.Infof("Registering Runtime on Director service")

	if config == nil {
		return "", errors.New("Cannot register runtime in Director: missing Runtime config")
	}

	var labels *graphql.Labels
	if config.Labels != nil {
		l := graphql.Labels(*config.Labels)
		labels = &l
	}

	directorInput := graphql.RuntimeInput{
		Name:        config.Name,
		Description: config.Description,
		Labels:      labels,
	}

	runtimeInput, err := cc.graphqlizer.RuntimeInputToGQL(directorInput)
	if err != nil {
		log.Infof("Failed to create graphQLized Runtime input")
		return "", err
	}

	runtimeQuery := cc.queryProvider.createRuntimeMutation(runtimeInput)

	var response CreateRuntimeResponse
	err = cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
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

func (cc *directorClient) GetRuntime(id, tenant string) (graphql.Runtime, error) {
	log.Infof("Getting Runtime on Director service")

	runtimeQuery := cc.queryProvider.getRuntimeMutation(id)

	var response GetRuntimeResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return graphql.Runtime{}, errors.Wrap(err, fmt.Sprintf("Failed to get runtime %s from Director", id))
	}
	if response.Result == nil {
		return graphql.Runtime{}, errors.Errorf("Failed to get runtime %s get Director: received nil response.", id)
	}
	if response.Result.ID != id {
		return graphql.Runtime{}, errors.Errorf("Failed to get correctly runtime %s in Director: Received wrong Runtime in the response", id)
	}

	log.Infof("Successfully got Runtime %s from Director for tenant %s", id, tenant)
	return *response.Result, nil
}

func (cc *directorClient) UpdateRuntime(id string, config *gqlschema.RuntimeInput, tenant string) error {
	log.Infof("Updating Runtime on Director service")

	if config == nil {
		return errors.New("Cannot update runtime in Director: missing Runtime config")
	}
	runtimeInput, err := cc.graphqlizer.RuntimeInputToGraphQL(*config)
	if err != nil {
		log.Infof("Failed to create graphQLized Runtime input")
		return err
	}
	runtimeQuery := cc.queryProvider.updateRuntimeMutation(id, runtimeInput)

	var response UpdateRuntimeResponse
	err = cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to update runtime %s in Director", id))
	}
	if response.Result == nil {
		return errors.Errorf("Failed to update runtime %s in Director: received nil response.", id)
	}
	if response.Result.ID != id {
		return errors.Errorf("Failed to update correctly the runtime %s in Director: Received bad Runtime id in response", id)
	}

	log.Infof("Successfully updated Runtime %s in Director for tenant %s", id, tenant)
	return nil
}

func (cc *directorClient) DeleteRuntime(id, tenant string) error {
	runtimeQuery := cc.queryProvider.deleteRuntimeMutation(id)

	var response DeleteRuntimeResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to unregister runtime %s in Director", id))
	}
	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.Errorf("Failed to unregister runtime %s in Director: received nil response.", id)
	}

	if response.Result.ID != id {
		return errors.Errorf("Failed to unregister correctly the runtime %s in Director: Received bad Runtime id in response", id)
	}

	log.Infof("Successfully unregistered Runtime %s in Director for tenant %s", id, tenant)

	return nil
}

func (cc *directorClient) SetRuntimeStatusCondition(id string, statusCondition gqlschema.RuntimeStatusCondition, tenant string) error {
	runtime, err := cc.GetRuntime(id, tenant)
	if err != nil {
		log.Errorf("Failed to get Runtime by ID: %s", err.Error())
		return errors.Wrap(err, "failed to get runtime by ID")
	}
	runtimeInput := &gqlschema.RuntimeInput{
		Name:            runtime.Name,
		Description:     runtime.Description,
		StatusCondition: &statusCondition,
	}
	err = cc.UpdateRuntime(id, runtimeInput, tenant)
	if err != nil {
		log.Errorf("Failed to update Runtime in Director: %s", err.Error())
		return errors.Wrap(err, "failed to update runtime in Director")
	}
	return nil
}

func (cc *directorClient) GetConnectionToken(id, tenant string) (graphql.OneTimeTokenForRuntimeExt, error) {
	runtimeQuery := cc.queryProvider.requestOneTimeTokeneMutation(id)

	var response OneTimeTokenResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return graphql.OneTimeTokenForRuntimeExt{}, errors.Wrap(err, fmt.Sprintf("Failed to get OneTimeToken for Runtime %s in Director", id))
	}

	if response.Result == nil {
		return graphql.OneTimeTokenForRuntimeExt{}, errors.Errorf("Failed to get OneTimeToken for Runtime %s in Director: received nil response.", id)
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

func (cc *directorClient) executeDirectorGraphQLCall(directorQuery string, tenant string, response interface{}) error {
	if cc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := cc.getToken(); err != nil {
			return err
		}
	}

	req := gcli.NewRequest(directorQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantHeader, tenant)

	err := cc.gqlClient.Do(req, response)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to execute GraphQL query with Director"))
	}

	return nil
}
