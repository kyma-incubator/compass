package director

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"

	machineGraph "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// accountIDKey is a header key name for request send by graphQL client
const accountIDKey = "tenant"

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

// NewDirectorClient returns new director client struct pointer
func NewDirectorClient(oauthClient OauthClient, gqlClient GraphQLClient) *Client {
	return &Client{
		graphQLClient: gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		token:         oauth.Token{},
	}
}

// GetConsoleURL fetches, validates and returns console URL from director component based on runtime ID
func (dc *Client) GetConsoleURL(accountID, runtimeID string) (string, error) {
	err := dc.setToken()
	if err != nil {
		return "", errors.Wrap(err, "while fetching token to director")
	}
	query := dc.queryProvider.Runtime(runtimeID)
	req := machineGraph.NewRequest(query)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dc.token.AccessToken))
	req.Header.Add(accountIDKey, accountID)

	var response graphql.RuntimeExt
	err = dc.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return "", TemporaryError{fmt.Sprintf("Failed to provision Runtime: %s", err)}
	}

	return dc.getURLFromRuntime(response)
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

func (dc Client) getURLFromRuntime(response graphql.RuntimeExt) (string, error) {
	if response.Status == nil {
		return "", TemporaryError{"response status from director is nil"}
	}
	if response.Status.Condition == graphql.RuntimeStatusConditionFailed {
		return "", fmt.Errorf("response status condition from director is %s", graphql.RuntimeStatusConditionFailed)
	}
	if response.Status.Condition != graphql.RuntimeStatusConditionReady {
		return "", TemporaryError{fmt.Sprintf("response status condition is not %q", graphql.RuntimeStatusConditionReady)}
	}
	if len(response.Labels) != 1 {
		return "", errors.New("response has more than one label")
	}

	value, ok := response.Labels[consoleURLLabelKey]
	if !ok {
		return "", fmt.Errorf("response label key is not equal to %q", consoleURLLabelKey)
	}

	var URL string
	switch value.(type) {
	case string:
		URL = value.(string)
	default:
		return "", errors.New("response label value is not string")
	}

	return URL, nil
}
