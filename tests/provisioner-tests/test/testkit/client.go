package testkit

import (
	"context"
	"crypto/tls"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"net/http"
)

type ProvisionerClient interface {
	ProvisionRuntime(id schema.RuntimeIDInput, config schema.ProvisionRuntimeInput) (schema.AsyncOperationID, error)
	UpgradeRuntime(id schema.RuntimeIDInput, config schema.UpgradeRuntimeInput) (schema.AsyncOperationID, error)
	DeprovisionRuntime(id schema.RuntimeIDInput) (schema.AsyncOperationID, error)
	ReconnectRuntimeAgent(id schema.RuntimeIDInput) (schema.AsyncOperationID, error)
	RuntimeStatus(id schema.RuntimeIDInput) (schema.RuntimeStatus, error)
	RuntimeOperationStatus(id schema.AsyncOperationIDInput) (schema.OperationStatus, error)
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

func (c client) ProvisionRuntime(id schema.RuntimeIDInput, config schema.ProvisionRuntimeInput) (schema.AsyncOperationID, error) {
	runtimeIptGQL, err := c.graphqlizer.RuntimeIDInputToGraphQL(id)
	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to convert RuntimeID Input to query")
	}

	provisionRuntimeIptGQL, err := c.graphqlizer.ProvisionRuntimeInputToGraphQL(config)
	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to convert Provision Runtime Input to query")
	}

	query := c.queryProvider.provisionRuntime(runtimeIptGQL, provisionRuntimeIptGQL)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err = c.graphQLClient.Run(context.Background(), req, &response)

	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to provision runtime")
	}
	return response.Result, nil
}

func (c client) UpgradeRuntime(id schema.RuntimeIDInput, config schema.UpgradeRuntimeInput) (schema.AsyncOperationID, error) {
	runtimeIptGQL, err := c.graphqlizer.RuntimeIDInputToGraphQL(id)
	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to convert RuntimeID Input to query")
	}

	upgradeRuntimeIptGQL, err := c.graphqlizer.UpgradeRuntimeInputToGraphQL(config)
	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to convert Upgrade Runtime Input to query")
	}

	query := c.queryProvider.provisionRuntime(runtimeIptGQL, upgradeRuntimeIptGQL)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err = c.graphQLClient.Run(context.Background(), req, &response)

	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to upgrade runtime")
	}
	return response.Result, nil
}

func (c client) DeprovisionRuntime(id schema.RuntimeIDInput) (schema.AsyncOperationID, error) {
	runtimeIptGQL, err := c.graphqlizer.RuntimeIDInputToGraphQL(id)
	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to convert RuntimeID Input to query")
	}

	query := c.queryProvider.deprovisionRuntime(runtimeIptGQL)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err = c.graphQLClient.Run(context.Background(), req, &response)

	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to deprovision runtime")
	}
	return response.Result, nil
}

func (c client) ReconnectRuntimeAgent(id schema.RuntimeIDInput) (schema.AsyncOperationID, error) {
	runtimeIptGQL, err := c.graphqlizer.RuntimeIDInputToGraphQL(id)
	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to convert RuntimeID Input to query")
	}

	query := c.queryProvider.deprovisionRuntime(runtimeIptGQL)
	req := gcli.NewRequest(query)

	var response AsyncOperationIDResult

	err = c.graphQLClient.Run(context.Background(), req, &response)

	if err != nil {
		return schema.AsyncOperationID{}, errors.Wrap(err, "Failed to reconnect runtime agent")
	}
	return response.Result, nil
}

func (c client) RuntimeStatus(id schema.RuntimeIDInput) (schema.RuntimeStatus, error) {
	runtimeIptGQL, err := c.graphqlizer.RuntimeIDInputToGraphQL(id)
	if err != nil {
		return schema.RuntimeStatus{}, errors.Wrap(err, "Failed to convert RuntimeID Input to query")
	}

	query := c.queryProvider.deprovisionRuntime(runtimeIptGQL)
	req := gcli.NewRequest(query)

	var response RuntimeStatusStatusResult

	err = c.graphQLClient.Run(context.Background(), req, &response)

	if err != nil {
		return schema.RuntimeStatus{}, errors.Wrap(err, "Failed to get runtime status")
	}
	return response.Result, nil
}

func (c client) RuntimeOperationStatus(id schema.AsyncOperationIDInput) (schema.OperationStatus, error) {
	runtimeIptGQL, err := c.graphqlizer.AsyncOperationIDInputToGraphQL(id)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to convert RuntimeID Input to query")
	}

	query := c.queryProvider.deprovisionRuntime(runtimeIptGQL)
	req := gcli.NewRequest(query)

	var response OperationStatus

	err = c.graphQLClient.Run(context.Background(), req, &response)

	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to get runtime operation status")
	}
	return response.Result, nil
}

type AsyncOperationIDResult struct {
	Result schema.AsyncOperationID `json:"result"`
}

type RuntimeStatusStatusResult struct {
	Result schema.RuntimeStatus `json:"result"`
}

type OperationStatus struct {
	Result schema.OperationStatus `json:"result"`
}
