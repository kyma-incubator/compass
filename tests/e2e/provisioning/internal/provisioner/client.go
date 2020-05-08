package provisioner

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"time"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	tenantHeader = "Tenant"
)

type Client struct {
	tenant             string
	graphQLClient      *gcli.Client
	queryProvider      queryProvider
	graphqlizer        graphqlizer
	componentsProvider *ComponentsListProvider
}

func NewProvisionerClient(endpoint, tenant string, logger logrus.FieldLogger, client *http.Client) *Client {

	gqlClinet := gcli.NewClient(endpoint, gcli.WithHTTPClient(client))
	gqlClinet.Log = func(s string) { logger.Print(s) }

	return &Client{
		tenant:        tenant,
		graphQLClient: gqlClinet,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
	}
}

func (c Client) UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (schema.OperationStatus, error) {
	upgradeRuntimeIptGQL, err := c.graphqlizer.UpgradeRuntimeInputToGraphQL(config)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to convert Upgrade Runtime Input to query")
	}

	query := c.queryProvider.upgradeRuntime(runtimeID, upgradeRuntimeIptGQL)
	req := c.newRequest(query)

	var operationStatus schema.OperationStatus
	err = c.executeRequest(req, &operationStatus)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to upgrade Runtime")
	}
	return operationStatus, nil
}

func (c Client) RuntimeStatus(runtimeID string) (schema.RuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := c.newRequest(query)

	var response schema.RuntimeStatus
	err := c.executeRequest(req, &response)
	if err != nil {
		return schema.RuntimeStatus{}, errors.Wrap(err, "Failed to get Runtime status")
	}
	return response, nil
}

func (c Client) RuntimeOperationStatus(operationID string) (schema.OperationStatus, error) {
	query := c.queryProvider.runtimeOperationStatus(operationID)
	req := c.newRequest(query)

	var response schema.OperationStatus
	err := c.executeRequest(req, &response)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to get Runtime operation status")
	}
	return response, nil
}

func (c Client) newRequest(query string) *gcli.Request {
	req := gcli.NewRequest(query)

	req.Header.Add(tenantHeader, c.tenant)

	return req
}

type graphQLResponseWrapper struct {
	Result interface{} `json:"result"`
}

// executeRequest executes GraphQL request and unmarshal response to respDestination.
func (c Client) executeRequest(req *gcli.Request, respDestination interface{}) error {
	if reflect.ValueOf(respDestination).Kind() != reflect.Ptr {
		return errors.New("destination is not of pointer type")
	}

	wrapper := &graphQLResponseWrapper{Result: respDestination}

	err := retry.Do(func() error {
		return c.graphQLClient.Run(context.Background(), req, wrapper)
	}, retry.Delay(1*time.Second), retry.Attempts(5))

	if err != nil {
		return errors.Wrap(err, "Failed to execute request")
	}

	return nil
}
