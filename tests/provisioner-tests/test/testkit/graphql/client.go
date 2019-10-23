package graphql

import (
	"context"
	"crypto/tls"
	"net/http"
	"reflect"

	"github.com/sirupsen/logrus"

	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Client struct {
	graphQLClient *gcli.Client
}

func NewGraphQLClient(endpoint string, skipTLSVerify, queryLogging bool) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLSVerify,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	if queryLogging {
		logger := logrus.WithField("Client", "GraphQL")
		graphQlClient.Log = func(s string) {
			logger.Info(s)
		}
	}

	return &Client{
		graphQLClient: graphQlClient,
	}
}

type graphQLResponseWrapper struct {
	Result interface{} `json:"result"`
}

// ExecuteRequest executes GraphQL request and unmarshal response to respDestination.
// If empty object is provided and unmarshaled response is equal to it, the error is returned.
// Due to GraphQL client not checking response codes we need to relay on result being empty in case of failure.
func (c Client) ExecuteRequest(req *gcli.Request, respDestination interface{}, emptyObject ...interface{}) error {
	if reflect.ValueOf(respDestination).Kind() != reflect.Ptr {
		return errors.New("destination is not of pointer type")
	}

	wrapper := &graphQLResponseWrapper{Result: respDestination}
	err := c.graphQLClient.Run(context.Background(), req, wrapper)
	if err != nil {
		return errors.Wrap(err, "Failed to execute request")
	}

	if len(emptyObject) > 0 {
		if reflect.DeepEqual(respDestination, emptyObject[0]) {
			return errors.New("Failed to execute request: received empty object response")
		}
	}

	return nil
}
