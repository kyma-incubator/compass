package provisioner

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type runtime struct {
	runtimeInput schema.ProvisionRuntimeInput
}

type fakeClient struct {
	mu         sync.Mutex
	runtimes   []runtime
	operations map[string]schema.OperationStatus
}

func NewFakeClient() *fakeClient {
	return &fakeClient{
		runtimes:   []runtime{},
		operations: make(map[string]schema.OperationStatus, 1),
	}
}

func (c *fakeClient) GetProvisionRuntimeInput(index int) schema.ProvisionRuntimeInput {
	c.mu.Lock()
	defer c.mu.Unlock()

	r := c.runtimes[index]
	return r.runtimeInput
}

func (c *fakeClient) FinishProvisionerOperation(id string, state schema.OperationState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	op := c.operations[id]
	op.State = state
	c.operations[id] = op
}

func (c *fakeClient) SetOperation(id string, operation schema.OperationStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.operations[id] = operation
}

// Provisioner Client methods

func (c *fakeClient) ProvisionRuntime(accountID, subAccountID string, config schema.ProvisionRuntimeInput) (schema.OperationStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rid := uuid.New().String()
	opId := uuid.New().String()
	c.runtimes = append(c.runtimes, runtime{
		runtimeInput: config,
	})
	c.operations = map[string]schema.OperationStatus{
		opId: {
			ID:        &opId,
			RuntimeID: &rid,
			Operation: schema.OperationTypeProvision,
			State:     schema.OperationStateInProgress,
		},
	}
	return schema.OperationStatus{
		RuntimeID: &rid,
		ID:        &opId,
	}, nil
}

func (c *fakeClient) DeprovisionRuntime(accountID, runtimeID string) (string, error) {
	return uuid.New().String(), nil
}

func (c *fakeClient) ReconnectRuntimeAgent(accountID, runtimeID string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *fakeClient) RuntimeOperationStatus(accountID, operationID string) (schema.OperationStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	o, found := c.operations[operationID]
	if !found {
		return schema.OperationStatus{}, fmt.Errorf("operation not found")
	}
	return o, nil
}
