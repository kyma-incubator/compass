package provisioner

import (
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"

	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Client interface {
	ProvisionRuntime(config schema.ProvisionRuntimeInput) (operationStatusID string, runtimeID string, err error)
	UpgradeRuntime(runtimeID string, config schema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(runtimeID string) (string, error)
	ReconnectRuntimeAgent(runtimeID string) (string, error)
	RuntimeStatus(runtimeID string) (RuntimeStatus, error)
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

func (c client) ProvisionRuntime(config schema.ProvisionRuntimeInput) (operationStatusID string, runtimeID string, err error) {
	provisionRuntimeIptGQL, err := c.graphqlizer.ProvisionRuntimeInputToGraphQL(config)
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to convert Provision Runtime Input to query")
	}

	query := c.queryProvider.provisionRuntime(provisionRuntimeIptGQL)
	req := gcli.NewRequest(query)

	var operationStatus schema.OperationStatus
	err = c.graphQLClient.ExecuteRequest(req, &operationStatus, "")
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to provision Runtime")
	}
	if operationStatus.ID == nil || operationStatus.RuntimeID == nil {
		return "", "", errors.New("Failed to receive proper Operation Status response")
	}
	return *operationStatus.ID, *operationStatus.RuntimeID, nil
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

func (c client) DeprovisionRuntime(runtimeID string) (string, error) {
	query := c.queryProvider.deprovisionRuntime(runtimeID)
	req := gcli.NewRequest(query)

	var operationId string
	err := c.graphQLClient.ExecuteRequest(req, &operationId, "")
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

type RuntimeStatus struct {
	LastOperationStatus     *schema.OperationStatus         `json:"lastOperationStatus"`
	RuntimeConnectionStatus *schema.RuntimeConnectionStatus `json:"runtimeConnectionStatus"`
	RuntimeConfiguration    RuntimeConfiguration            `json:"runtimeConfiguration"`
}

type RuntimeConfiguration struct {
	ClusterConfig         interface{}        `json:"clusterConfig"`
	KymaConfig            *schema.KymaConfig `json:"kymaConfig"`
	Kubeconfig            *string            `json:"kubeconfig"`
	CredentialsSecretName *string            `json:"credentialsSecretName"`
}

func (c client) RuntimeStatus(runtimeID string) (RuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := gcli.NewRequest(query)

	var response RuntimeStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &RuntimeStatus{})
	if err != nil {
		return RuntimeStatus{}, errors.Wrap(err, "Failed to get Runtime status")
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
