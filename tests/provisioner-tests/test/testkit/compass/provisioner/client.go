package provisioner

import (
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Client interface {
	ProvisionRuntime(runtimeID string, config schema.ProvisionRuntimeInput) (string, error)
	UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(runtimeID string, input schema.CredentialsInput) (string, error)
	ReconnectRuntimeAgent(runtimeID string) (string, error)
	RuntimeStatus(runtimeID string) (schema.RuntimeStatus, error)
	RuntimeOperationStatus(operationID string) (schema.OperationStatus, error)
}

type client struct {
	graphQLClient *graphql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer
}

func NewProvisionerClient(endpoint string, queryLogging bool) Client {
	return &client{
		graphQLClient: graphql.NewGraphQLClient(endpoint, true, queryLogging),
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

	var operationId string
	err = c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to provision Runtime")
	}
	return operationId, nil
}

func (c client) UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error) {
	upgradeRuntimeIptGQL, err := c.graphqlizer.UpgradeRuntimeInputToGraphQL(config)
	if err != nil {
		return "", errors.Wrap(err, "Failed to convert Upgrade Runtime Input to query")
	}

	query := c.queryProvider.upgradeRuntime(runtimeID, upgradeRuntimeIptGQL)
	req := gcli.NewRequest(query)

	var operationId string
	err = c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to upgrade Runtime")
	}
	return operationId, nil
}

func (c client) DeprovisionRuntime(runtimeID string, input schema.CredentialsInput) (string, error) {
	credentialsInput, err := c.graphqlizer.CredentialsInputToGraphQL(input)
	if err != nil {
		return "", errors.Wrap(err, "Failed to deprovision Runtime: Failed to convert Credentials Input to query")
	}

	query := c.queryProvider.deprovisionRuntime(runtimeID, credentialsInput)
	req := gcli.NewRequest(query)

	var operationId string
	err = c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to deprovision Runtime")
	}
	return operationId, nil
}

func (c client) ReconnectRuntimeAgent(runtimeID string) (string, error) {
	query := c.queryProvider.reconnectRuntimeAgent(runtimeID)
	req := gcli.NewRequest(query)

	var operationId string
	err := c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to reconnect Runtime agent")
	}
	return operationId, nil
}

func (c client) RuntimeStatus(runtimeID string) (schema.RuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := gcli.NewRequest(query)

	var response schema.RuntimeStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &schema.RuntimeStatus{})
	if err != nil {
		return schema.RuntimeStatus{}, errors.Wrap(err, "Failed to get Runtime status")
	}
	return response, nil
}

func (c client) RuntimeOperationStatus(operationID string) (schema.OperationStatus, error) {
	query := c.queryProvider.runtimeOperationStatus(operationID)
	req := gcli.NewRequest(query)

	var response schema.OperationStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &schema.OperationStatus{})
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to get Runtime operation status")
	}
	return response, nil
}
