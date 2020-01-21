package provisioner

import (
	"fmt"
	"sync"

	"time"

	"github.com/google/uuid"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type runtime struct {
	id                 string
	runtimeInput       schema.ProvisionRuntimeInput
	provisionStartedAt time.Time
}

type fakeClient struct {
	mu         sync.Mutex
	runtimes   []runtime
	operations map[string]schema.OperationStatus
}

func NewFakeClient() *fakeClient {
	return &fakeClient{
		runtimes: []runtime{},
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

// Provisioner Client methods

func (c *fakeClient) ProvisionRuntime(accountID, id string, config schema.ProvisionRuntimeInput) (schema.OperationStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rid := uuid.New().String()
	opId := uuid.New().String()
	c.runtimes = append(c.runtimes, runtime{
		runtimeInput:       config,
		provisionStartedAt: time.Now(),
		id:                 rid,
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

func (c *fakeClient) UpgradeRuntime(accountID, runtimeID string, config schema.UpgradeRuntimeInput) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *fakeClient) DeprovisionRuntime(accountID, runtimeID string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *fakeClient) ReconnectRuntimeAgent(accountID, runtimeID string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *fakeClient) GCPRuntimeStatus(accountID, runtimeID string) (GCPRuntimeStatus, error) {
	return GCPRuntimeStatus{}, fmt.Errorf("not implemented")
}

func (c *fakeClient) RuntimeOperationStatus(accountID, operationID string) (schema.OperationStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	o, found := c.operations[operationID]
	for _, r := range c.runtimes {
		if r.id != *o.RuntimeID {
			continue
		}

		if time.Since(r.provisionStartedAt) > time.Minute {
			o.State = schema.OperationStateSucceeded
		}
	}
	if !found {
		return schema.OperationStatus{}, fmt.Errorf("operation not found")
	}
	return o, nil
}
