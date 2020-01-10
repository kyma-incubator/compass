package provisioner

import (
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"

	"fmt"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Client interface {
	ProvisionRuntime(runtimeID string, config schema.ProvisionRuntimeInput) (schema.OperationStatus, error)
	UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(runtimeID string) (string, error)
	ReconnectRuntimeAgent(runtimeID string) (string, error)
	GCPRuntimeStatus(runtimeID string) (GCPRuntimeStatus, error)
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

func (c *client) ProvisionRuntime(runtimeID string, config schema.ProvisionRuntimeInput) (schema.OperationStatus, error) {
	provisionRuntimeIptGQL, err := c.graphqlizer.ProvisionRuntimeInputToGraphQL(config)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to convert Provision Runtime Input to query")
	}

	query := c.queryProvider.provisionRuntime(runtimeID, provisionRuntimeIptGQL)
	req := gcli.NewRequest(query)

	fmt.Println(query)

	var operationId string
	err = c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to provision Runtime")
	}
	return schema.OperationStatus{ID: &operationId, RuntimeID: &runtimeID}, nil
}

func (c *client) UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error) {
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

func (c *client) DeprovisionRuntime(runtimeID string) (string, error) {
	query := c.queryProvider.deprovisionRuntime(runtimeID)
	req := gcli.NewRequest(query)

	var operationId string
	err := c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to deprovision Runtime")
	}
	return operationId, nil
}

func (c *client) ReconnectRuntimeAgent(runtimeID string) (string, error) {
	query := c.queryProvider.reconnectRuntimeAgent(runtimeID)
	req := gcli.NewRequest(query)

	var operationId string
	err := c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to reconnect Runtime agent")
	}
	return operationId, nil
}

type GCPRuntimeStatus struct {
	LastOperationStatus     *schema.OperationStatus         `json:"lastOperationStatus"`
	RuntimeConnectionStatus *schema.RuntimeConnectionStatus `json:"runtimeConnectionStatus"`
	RuntimeConfiguration    struct {
		ClusterConfig         *schema.GCPConfig  `json:"clusterConfig"`
		KymaConfig            *schema.KymaConfig `json:"kymaConfig"`
		Kubeconfig            *string            `json:"kubeconfig"`
		CredentialsSecretName *string            `json:"credentialsSecretName"`
	} `json:"runtimeConfiguration"`
}

func (c *client) GCPRuntimeStatus(runtimeID string) (GCPRuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := gcli.NewRequest(query)

	var response GCPRuntimeStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &GCPRuntimeStatus{})
	if err != nil {
		return GCPRuntimeStatus{}, errors.Wrap(err, "Failed to get Runtime status")
	}
	return response, nil
}

func (c *client) RuntimeOperationStatus(operationID string) (schema.OperationStatus, error) {
	query := c.queryProvider.runtimeOperationStatus(operationID)
	req := gcli.NewRequest(query)

	var response schema.OperationStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &schema.OperationStatus{})
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to get Runtime operation status")
	}
	return response, nil
}
