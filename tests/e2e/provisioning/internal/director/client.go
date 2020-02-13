package director

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/director/oauth"

	machineGraph "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// tenantHeaderKey is a header key name for request send by graphQL client
	tenantHeaderKey = "tenant"

	// amount of request attempt to director service
	reqAttempt = 3
)

//go:generate mockery -name=GraphQLClient -output=automock
type GraphQLClient interface {
	Run(ctx context.Context, req *machineGraph.Request, resp interface{}) error
}

//go:generate mockery -name=OauthClient -output=automock
type OauthClient interface {
	GetAuthorizationToken() (oauth.Token, error)
}

type Client struct {
	graphQLClient GraphQLClient
	oauthClient   OauthClient
	queryProvider queryProvider
	token         oauth.Token
}

type successResponse struct {
	Result graphql.RuntimeExt `json:"result"`
}

var lock sync.Mutex

// NewDirectorClient returns new director client struct pointer
func NewDirectorClient(oauthClient OauthClient, gqlClient GraphQLClient) *Client {
	return &Client{
		graphQLClient: gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		token:         oauth.Token{},
	}
}

// GetRuntimeID fetches runtime ID with given label name from director component
func (dc *Client) GetRuntimeID(accountID, instanceID string) (string, error) {
	log.WithField("service", "director-client")
	log.Info("Create request to director service")

	query := dc.queryProvider.Runtime(instanceID)
	req := machineGraph.NewRequest(query)
	req.Header.Add(tenantHeaderKey, accountID)

	log.Info("Send request to director")
	response, err := dc.callDirector(req)
	if err != nil {
		// do not wrap error, because type of error (TemporaryError) is important
		return "", err
	}

	log.Info("Extract the RuntimeID from the response")
	return dc.getIDFromRuntime(&response.Result)
}

func (dc *Client) callDirector(req *machineGraph.Request) (*successResponse, error) {
	var response *successResponse
	var lastError error
	var success bool
	authorizationKey := "Authorization"

	for i := 0; i < reqAttempt; i++ {
		err := dc.setToken()
		if err != nil {
			lastError = err
			log.Errorf("cannot set token to director client (attempt %d): %s", i, err)
			continue
		}
		req.Header.Add(authorizationKey, fmt.Sprintf("Bearer %s", dc.token.AccessToken))
		response, err = dc.call(req)
		if err != nil {
			lastError = err
			dc.token.AccessToken = ""
			req.Header.Del(authorizationKey)
			log.Errorf("call to director failed (attempt %d): %s", i, err)
			continue
		}
		success = true
		break
	}

	if !success {
		return &successResponse{}, lastError
	}

	return response, nil
}

func (dc *Client) call(req *machineGraph.Request) (*successResponse, error) {
	var response successResponse
	err := dc.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return nil, TemporaryError{fmt.Sprintf("while requesting to director client: %s", err)}
	}
	return &response, nil
}

func (dc *Client) setToken() error {
	lock.Lock()
	defer lock.Unlock()
	if !dc.token.EmptyOrExpired() {
		return nil
	}

	token, err := dc.oauthClient.GetAuthorizationToken()
	if err != nil {
		return errors.Wrap(err, "Error while obtaining token")
	}
	dc.token = token

	return nil
}

func (dc *Client) getIDFromRuntime(response *graphql.RuntimeExt) (string, error) {
	if response.Status == nil {
		return "", TemporaryError{"response status from director is nil"}
	}
	if response.Status.Condition == graphql.RuntimeStatusConditionFailed {
		return "", fmt.Errorf("response status condition from director is %s", graphql.RuntimeStatusConditionFailed)
	}

	if response.Runtime.ID == "" {
		return "", fmt.Errorf("got empty runtime ID from director from runtime name")
	}

	return response.Runtime.ID, nil
}
