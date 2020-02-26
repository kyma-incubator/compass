package provisioner

import (
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// accountIDKey is a header key name for request send by graphQL client
const accountIDKey = "tenant"

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore

type Client interface {
	ProvisionRuntime(accountID string, config schema.ProvisionRuntimeInput) (schema.OperationStatus, error)
	UpgradeRuntime(accountID, runtimeID string, config schema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(accountID, runtimeID string) (string, error)
	ReconnectRuntimeAgent(accountID, runtimeID string) (string, error)
	GCPRuntimeStatus(accountID, runtimeID string) (GCPRuntimeStatus, error)
	RuntimeOperationStatus(accountID, operationID string) (schema.OperationStatus, error)
}

type client struct {
	graphQLClient *graphql.Client
	queryProvider queryProvider
	graphqlizer   Graphqlizer
}

func NewProvisionerClient(endpoint string, queryLogging bool) Client {
	return &client{
		graphQLClient: graphql.NewGraphQLClient(endpoint, true, queryLogging),
		queryProvider: queryProvider{},
		graphqlizer:   Graphqlizer{},
	}
}

func (c *client) ProvisionRuntime(accountID string, config schema.ProvisionRuntimeInput) (schema.OperationStatus, error) {
	provisionRuntimeIptGQL, err := c.graphqlizer.ProvisionRuntimeInputToGraphQL(config)
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to convert Provision Runtime Input to query")
	}

	query := c.queryProvider.provisionRuntime(provisionRuntimeIptGQL)
	req := gcli.NewRequest(query)
	req.Header.Add(accountIDKey, accountID)

	var response schema.OperationStatus
	err = c.graphQLClient.ExecuteRequest(req, &response, schema.OperationStatus{})
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to provision Runtime")
	}

	return response, nil
}

func (c *client) UpgradeRuntime(accountID, runtimeID string, config schema.UpgradeRuntimeInput) (string, error) {
	upgradeRuntimeIptGQL, err := c.graphqlizer.UpgradeRuntimeInputToGraphQL(config)
	if err != nil {
		return "", errors.Wrap(err, "Failed to convert Upgrade Runtime Input to query")
	}

	query := c.queryProvider.upgradeRuntime(runtimeID, upgradeRuntimeIptGQL)
	req := gcli.NewRequest(query)
	req.Header.Add(accountIDKey, accountID)

	var operationId string
	err = c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to upgrade Runtime")
	}
	return operationId, nil
}

func (c *client) DeprovisionRuntime(accountID, runtimeID string) (string, error) {
	query := c.queryProvider.deprovisionRuntime(runtimeID)
	req := gcli.NewRequest(query)
	req.Header.Add(accountIDKey, accountID)

	var operationId string
	err := c.graphQLClient.ExecuteRequest(req, &operationId, "")
	if err != nil {
		return "", errors.Wrap(err, "Failed to deprovision Runtime")
	}
	return operationId, nil
}

func (c *client) ReconnectRuntimeAgent(accountID, runtimeID string) (string, error) {
	query := c.queryProvider.reconnectRuntimeAgent(runtimeID)
	req := gcli.NewRequest(query)
	req.Header.Add(accountIDKey, accountID)

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

func (c *client) GCPRuntimeStatus(accountID, runtimeID string) (GCPRuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := gcli.NewRequest(query)
	req.Header.Add(accountIDKey, accountID)

	var response GCPRuntimeStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &GCPRuntimeStatus{})
	if err != nil {
		return GCPRuntimeStatus{}, errors.Wrap(err, "Failed to get Runtime status")
	}
	return response, nil
}

func (c *client) RuntimeOperationStatus(accountID, operationID string) (schema.OperationStatus, error) {
	query := c.queryProvider.runtimeOperationStatus(operationID)
	req := gcli.NewRequest(query)
	req.Header.Add(accountIDKey, accountID)

	var response schema.OperationStatus
	err := c.graphQLClient.ExecuteRequest(req, &response, &schema.OperationStatus{})
	if err != nil {
		return schema.OperationStatus{}, errors.Wrap(err, "Failed to get Runtime operation status")
	}
	return response, nil
}
