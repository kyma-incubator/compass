package testkit

import (
	"crypto/tls"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"net/http"
)

type ProvisionerClient interface {
	ProvisionRuntime(id schema.RuntimeIDInput, config schema.ProvisionRuntimeInput) schema.AsyncOperationID
	UpgradeRuntime(id schema.RuntimeIDInput, config schema.UpgradeRuntimeInput) schema.AsyncOperationID
	DeprovisionRuntime(id schema.RuntimeIDInput) schema.AsyncOperationID
	ReconnectRuntimeAgent(id schema.RuntimeIDInput) schema.AsyncOperationID
	RuntimeStatus(id schema.RuntimeIDInput) schema.RuntimeStatus
	RuntimeOperationStatus(id schema.AsyncOperationIDInput) schema.OperationStatus
}

type client struct {
	graphQLClient *gcli.Client
	queryProvider queryProvider
	grahqlizer    graphqlizer
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
		grahqlizer:    graphqlizer{},
	}
}

func (c client) ProvisionRuntime(id schema.RuntimeIDInput, config schema.ProvisionRuntimeInput) schema.AsyncOperationID {
}
func (c client) UpgradeRuntime(id schema.RuntimeIDInput, config schema.UpgradeRuntimeInput) schema.AsyncOperationID {
}
func (c client) DeprovisionRuntime(id schema.RuntimeIDInput) schema.AsyncOperationID           {}
func (c client) ReconnectRuntimeAgent(id schema.RuntimeIDInput) schema.AsyncOperationID        {}
func (c client) RuntimeStatus(id schema.RuntimeIDInput) schema.RuntimeStatus                   {}
func (c client) RuntimeOperationStatus(id schema.AsyncOperationIDInput) schema.OperationStatus {}

type AsyncOperationIDResult struct {
	Result schema.AsyncOperationID `json:"result"`
}

type RuntimeStatusStatusResult struct {
	Result schema.RuntimeStatus `json:"result"`
}

type OperationStatusResult struct {
	Result schema.OperationStatus `json:"result"`
}
