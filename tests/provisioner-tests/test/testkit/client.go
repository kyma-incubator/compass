package testkit

import (
	"context"
	"crypto/tls"
	"net/http"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type ProvisionerClient interface {
	ProvisionRuntime(runtimeID string, config schema.ProvisionRuntimeInput) (string, error)
	UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(runtimeID string) (string, error)
	ReconnectRuntimeAgent(runtimeID string) (string, error)
	RuntimeStatus(runtimeID string) (schema.RuntimeStatus, error)
	RuntimeOperationStatus(operationID string) (schema.OperationStatus, error)
}

type client struct {
	graphQLClient *gcli.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
}

func NewProvisionerClient(endpoint string) ProvisionerClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))
	return &client{
		graphQLClient: graphQlClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer{},
	}
}

func (c client) ProvisionRuntime(runtimeID string, config schema.ProvisionRuntimeInput) (string, error) {
	provisionRuntimeIptGQL, err := c.graphqlizer.ProvisionRuntimeInputToGraphQL(config)
	if err != nil {
		return "", errors.Wrap(err, "Failed to convert Provision Runtime Input to query")
	}

	query := c.queryProvider.provisionRuntime(runtimeID, provisionRuntimeIptGQL)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err = c.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to provision runtime")
	}
	return response.Result, nil
}

func (c client) UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error) {
	upgradeRuntimeIptGQL, err := c.graphqlizer.UpgradeRuntimeInputToGraphQL(config)
	if err != nil {
		return "", errors.Wrap(err, "Failed to convert Upgrade Runtime Input to query")
	}

	query := c.queryProvider.upgradeRuntime(runtimeID, upgradeRuntimeIptGQL)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err = c.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to upgrade runtime")
	}
	return response.Result, nil
}

func (c client) DeprovisionRuntime(runtimeID string) (string, error) {
	query := c.queryProvider.deprovisionRuntime(runtimeID)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err := c.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to deprovision runtime")
	}
	return response.Result, nil
}

func (c client) ReconnectRuntimeAgent(runtimeID string) (string, error) {
	query := c.queryProvider.reconnectRuntimeAgent(runtimeID)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err := c.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to reconnect runtime agent")
	}
	return response.Result, nil
}

func (c client) RuntimeStatus(runtimeID string) (schema.RuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := gcli.NewRequest(query)

	var response RuntimeStatusStatusResult

	err := c.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.RuntimeStatus{}, errors.Wrap(err, "Failed to get runtime status")
	}
	return response.Result, nil
}

func (c client) RuntimeOperationStatus(operationID string) (schema.OperationStatus, error) {
	query := c.queryProvider.runtimeOperationStatus(operationID)
	req := gcli.NewRequest(query)

	var response OperationStatus

	err := c.graphQLClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to get runtime operation status")
	}
	return response.Result, nil
}

type AsyncOperationIDResult struct {
	Result string `json:"result"`
}

type RuntimeStatusStatusResult struct {
	Result schema.RuntimeStatus `json:"result"`
}

type OperationStatus struct {
	Result schema.OperationStatus `json:"result"`
}
