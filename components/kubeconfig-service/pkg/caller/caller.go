package caller

import (
	"context"
	"reflect"
	"time"

	"github.com/avast/retry-go"
	schema "github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TenantHeader = "Tenant"
)

//Caller wrapper structure for the graphQL Client
type Caller struct {
	tenant        string
	client        *graphql.Client
	queryProvider queryProvider
}

//NewCaller return a new Caller instance
func NewCaller(endpoint, tenant string) *Caller {
	gqlClient := graphql.NewClient(endpoint)

	return &Caller{
		tenant:        tenant,
		client:        gqlClient,
		queryProvider: queryProvider{},
	}
}

func (c Caller) newRequest(query string) *graphql.Request {
	req := graphql.NewRequest(query)
	req.Header.Add(TenantHeader, c.tenant)

	return req
}

//RuntimeStatus return schema.RuntimeStatus
func (c Caller) RuntimeStatus(runtimeID string) (schema.RuntimeStatus, error) {
	query := c.queryProvider.runtimeStatus(runtimeID)
	req := c.newRequest(query)

	var response schema.RuntimeStatus
	err := c.executeRequest(req, &response)
	if err != nil {
		return schema.RuntimeStatus{}, errors.Wrap(err, "Failed to get Runtime status")
	}
	return response, nil
}

type graphQLResponseWrapper struct {
	Result interface{} `json:"result"`
}

// executeRequest executes GraphQL request and unmarshal response to respDestination.
func (c Caller) executeRequest(req *graphql.Request, respDestination interface{}) error {
	if reflect.ValueOf(respDestination).Kind() != reflect.Ptr {
		return errors.New("destination is not of pointer type")
	}

	wrapper := &graphQLResponseWrapper{Result: respDestination}

	err := retry.Do(func() error {
		return c.client.Run(context.Background(), req, wrapper)
	}, retry.Delay(1*time.Second), retry.Attempts(5))

	if err != nil {
		return errors.Wrap(err, "Failed to execute request")
	}

	return nil
}
