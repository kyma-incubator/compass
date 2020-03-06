package director

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/director/oauth"

	machineGraph "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	log           logrus.FieldLogger
}

type successResponse struct {
	Result graphql.RuntimePageExt `json:"result"`
}

// NewDirectorClient returns new director client struct pointer
func NewDirectorClient(oauthClient OauthClient, gqlClient GraphQLClient, log logrus.FieldLogger) *Client {
	return &Client{
		graphQLClient: gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		token:         oauth.Token{},
		log:           log,
	}
}

// GetRuntimeID fetches runtime ID with given label name from director component
func (dc *Client) GetRuntimeID(accountID, instanceID string) (string, error) {
	dc.log.WithField("service", "director-client")
	dc.log.Info("Create request to director service")

	query := dc.queryProvider.Runtime(instanceID)
	req := machineGraph.NewRequest(query)
	req.Header.Add(tenantHeaderKey, accountID)

	dc.log.Info("Send request to director")
	response, err := dc.callDirector(req)
	if err != nil {
		return "", err
	}

	dc.log.Info("Extract the RuntimeID from the response")
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
			dc.log.Errorf("cannot set token to director client (attempt %d): %s", i, err)
			continue
		}
		req.Header.Add(authorizationKey, fmt.Sprintf("Bearer %s", dc.token.AccessToken))
		response, err = dc.call(req)
		if err != nil {
			lastError = err
			dc.token.AccessToken = ""
			req.Header.Del(authorizationKey)
			dc.log.Errorf("call to director failed (attempt %d): %s", i, err)
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
		return nil, errors.Wrap(err, "while requesting to director client")
	}
	return &response, nil
}

func (dc *Client) setToken() error {
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

func (dc *Client) getIDFromRuntime(response *graphql.RuntimePageExt) (string, error) {
	if response.Data == nil || len(response.Data) == 0 {
		return "", errors.New("got empty data from director response")
	}
	if len(response.Data) > 1 {
		return "", errors.Errorf("expected single runtime, got: %v", response.Data)
	}
	return response.Data[0].ID, nil
}
