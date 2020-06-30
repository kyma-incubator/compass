package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	gql "github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
	TenantHeader        = "Tenant"
)

//go:generate mockery -name=DirectorClient
type DirectorClient interface {
	CreateRuntime(config *gqlschema.RuntimeInput, tenant string) (string, apperrors.AppError)
	GetRuntime(id, tenant string) (graphql.RuntimeExt, apperrors.AppError)
	UpdateRuntime(id string, config *graphql.RuntimeInput, tenant string) apperrors.AppError
	DeleteRuntime(id, tenant string) apperrors.AppError
	SetRuntimeStatusCondition(id string, statusCondition graphql.RuntimeStatusCondition, tenant string) apperrors.AppError
	GetConnectionToken(id, tenant string) (graphql.OneTimeTokenForRuntimeExt, apperrors.AppError)
	RuntimeExists(gardenerClusterName, tenant string) (bool, apperrors.AppError)
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

func (cc *directorClient) CreateRuntime(config *gqlschema.RuntimeInput, tenant string) (string, apperrors.AppError) {
	log.Infof("Registering Runtime on Director service")

	if config == nil {
		return "", apperrors.BadRequest("Cannot register runtime in Director: missing Runtime config")
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
		return "", apperrors.Internal("Failed to create graphQLized Runtime input: %s", err.Error())
	}

	runtimeQuery := cc.queryProvider.createRuntimeMutation(runtimeInput)

	var response CreateRuntimeResponse
	appErr := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if appErr != nil {
		return "", appErr.Append("Failed to register runtime in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", apperrors.Internal("Failed to register runtime in Director: Received nil response.")
	}

	log.Infof("Successfully registered Runtime %s in Director for tenant %s", config.Name, tenant)

	return response.Result.ID, nil
}

func (cc *directorClient) GetRuntime(id, tenant string) (graphql.RuntimeExt, apperrors.AppError) {
	log.Infof("Getting Runtime from Director service")

	runtimeQuery := cc.queryProvider.getRuntimeQuery(id)

	var response GetRuntimeResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return graphql.RuntimeExt{}, err.Append("Failed to get runtime %s from Director", id)
	}
	if response.Result == nil {
		return graphql.RuntimeExt{}, apperrors.Internal("Failed to get runtime %s get Director: received nil response.", id)
	}
	if response.Result.ID != id {
		return graphql.RuntimeExt{}, apperrors.Internal("Failed to get correctly runtime %s in Director: Received wrong Runtime in the response", id)
	}

	log.Infof("Successfully got Runtime %s from Director for tenant %s", id, tenant)
	return *response.Result, nil
}

func (cc *directorClient) UpdateRuntime(id string, directorInput *graphql.RuntimeInput, tenant string) apperrors.AppError {
	log.Infof("Updating Runtime in Director service")

	if directorInput == nil {
		return apperrors.BadRequest("Cannot update runtime in Director: missing Runtime config")
	}

	runtimeInput, err := cc.graphqlizer.RuntimeInputToGQL(*directorInput)
	if err != nil {
		log.Infof("Failed to create graphQLized Runtime input")
		return apperrors.Internal("Failed to create graphQLized Runtime input: %s", err.Error())
	}
	runtimeQuery := cc.queryProvider.updateRuntimeMutation(id, runtimeInput)

	var response UpdateRuntimeResponse
	appErr := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if appErr != nil {
		return appErr.Append("Failed to update runtime %s in Director", id)
	}
	if response.Result == nil {
		return apperrors.Internal("Failed to update runtime %s in Director: received nil response.", id)
	}
	if response.Result.ID != id {
		return apperrors.Internal("Failed to update correctly the runtime %s in Director: Received bad Runtime id in response", id)
	}

	log.Infof("Successfully updated Runtime %s in Director for tenant %s", id, tenant)
	return nil
}

func (cc *directorClient) DeleteRuntime(id, tenant string) apperrors.AppError {
	runtimeQuery := cc.queryProvider.deleteRuntimeMutation(id)

	var response DeleteRuntimeResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return err.Append("Failed to unregister runtime %s in Director", id)
	}
	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return apperrors.Internal("Failed to unregister runtime %s in Director: received nil response.", id)
	}

	if response.Result.ID != id {
		return apperrors.Internal("Failed to unregister correctly the runtime %s in Director: Received bad Runtime id in response", id)
	}

	log.Infof("Successfully unregistered Runtime %s in Director for tenant %s", id, tenant)

	return nil
}

func (cc *directorClient) RuntimeExists(id, tenant string) (bool, apperrors.AppError) {
	runtimeQuery := cc.queryProvider.getRuntimeQuery(id)

	var response GetRuntimeResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return false, err.Append("Failed to get runtime %s from Director", id)
	}

	if response.Result == nil {
		return false, nil
	}

	return true, nil
}

func (cc *directorClient) SetRuntimeStatusCondition(id string, statusCondition graphql.RuntimeStatusCondition, tenant string) apperrors.AppError {
	// TODO: Set StatusCondition without getting the Runtime
	//       It'll be possible after this issue implementation:
	//       - https://github.com/kyma-incubator/compass/issues/1186
	runtime, err := cc.GetRuntime(id, tenant)
	if err != nil {
		log.Errorf("Failed to get Runtime by ID: %s", err.Error())
		return err.Append("failed to get runtime by ID")
	}
	runtimeInput := &graphql.RuntimeInput{
		Name:            runtime.Name,
		Description:     runtime.Description,
		StatusCondition: &statusCondition,
		Labels:          &runtime.Labels,
	}
	err = cc.UpdateRuntime(id, runtimeInput, tenant)
	if err != nil {
		log.Errorf("Failed to update Runtime in Director: %s", err.Error())
		return err.Append("failed to update runtime in Director")
	}
	return nil
}

func (cc *directorClient) GetConnectionToken(id, tenant string) (graphql.OneTimeTokenForRuntimeExt, apperrors.AppError) {
	runtimeQuery := cc.queryProvider.requestOneTimeTokeneMutation(id)

	var response OneTimeTokenResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, tenant, &response)
	if err != nil {
		return graphql.OneTimeTokenForRuntimeExt{}, err.Append("Failed to get OneTimeToken for Runtime %s in Director", id)
	}

	if response.Result == nil {
		return graphql.OneTimeTokenForRuntimeExt{}, apperrors.Internal("Failed to get OneTimeToken for Runtime %s in Director: received nil response.", id)
	}

	log.Infof("Received OneTimeToken for Runtime %s in Director for tenant %s", id, tenant)

	return *response.Result, nil
}

func (cc *directorClient) getToken() apperrors.AppError {
	token, err := cc.oauthClient.GetAuthorizationToken()
	if err != nil {
		return err.Append("Error while obtaining token")
	}

	if token.EmptyOrExpired() {
		return apperrors.Internal("Obtained empty or expired token")
	}

	cc.token = token
	return nil
}

func (cc *directorClient) executeDirectorGraphQLCall(directorQuery string, tenant string, response interface{}) apperrors.AppError {
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
		return apperrors.Internal("Failed to execute GraphQL query with Director: %s", err.Error())
	}

	return nil
}
