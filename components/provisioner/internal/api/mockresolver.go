package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type MockResolver struct {
	cache cache.Cache
}

type externalMutationResolver struct {
	*MockResolver
}

type externalQueryResolver struct {
	*MockResolver
}

func (r *MockResolver) Mutation() gqlschema.MutationResolver {
	return &externalMutationResolver{r}
}
func (r *MockResolver) Query() gqlschema.QueryResolver {
	return &externalQueryResolver{r}
}

type currentOperation struct {
	lastOperation      gqlschema.OperationType
	status             gqlschema.OperationState
	operationID        string
	shouldStatusChange bool
}

func NewMockResolver(cache cache.Cache) *MockResolver {
	return &MockResolver{cache: cache}
}

func (r *MockResolver) ProvisionRuntime(ctx context.Context, id *gqlschema.RuntimeIDInput, config *gqlschema.ProvisionRuntimeInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}
	operationID := string(uuid.NewUUID())

	operation := currentOperation{
		lastOperation: gqlschema.OperationTypeProvision,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
	}

	r.changeStatus(operation, id)
	return &gqlschema.AsyncOperationID{ID: string(operationID)}, nil
}

func (r *MockResolver) UpgradeRuntime(ctx context.Context, id *gqlschema.RuntimeIDInput, config *gqlschema.UpgradeRuntimeInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := currentOperation{
		lastOperation: gqlschema.OperationTypeUpgrade,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
	}

	r.changeStatus(operation, id)
	return &gqlschema.AsyncOperationID{ID: string(uuid.NewUUID())}, nil
}

func (r *MockResolver) DeprovisionRuntime(ctx context.Context, id *gqlschema.RuntimeIDInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := currentOperation{
		lastOperation: gqlschema.OperationTypeDeprovision,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
	}

	r.changeStatus(operation, id)
	return &gqlschema.AsyncOperationID{ID: string(uuid.NewUUID())}, nil
}

func (r *MockResolver) ReconnectRuntimeAgent(ctx context.Context, id *gqlschema.RuntimeIDInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := currentOperation{
		lastOperation: gqlschema.OperationTypeReconnectRuntime,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
	}

	r.changeStatus(operation, id)

	r.changeStatus(operation, id)
	return &gqlschema.AsyncOperationID{ID: string(uuid.NewUUID())}, nil
}

func (r *MockResolver) RuntimeStatus(ctx context.Context, id *gqlschema.RuntimeIDInput) (*gqlschema.RuntimeStatus, error) {
	operation, _ := r.getStatus(id)

	return &gqlschema.RuntimeStatus{
		LastOperationStatus: &gqlschema.OperationStatus{
			Operation: operation.lastOperation,
			State:     operation.status,
		}}, nil
}

/* Runtime Operation Status always returns status set in operation call (usually In Progress) in first call after starting new operation
and status Succeeded in second and following calls until next operation is started.
*/

func (r *MockResolver) RuntimeOperationStatus(ctx context.Context, id *gqlschema.AsyncOperationIDInput) (*gqlschema.OperationStatus, error) {
	operation, runtimeID, exists := r.checkOperation(id)

	if !exists {
		return nil, errors.Errorf("Operation: %s does not exist", id)
	}

	if operation.shouldStatusChange {
		operation.status = gqlschema.OperationStateSucceeded
	} else {
		operation.shouldStatusChange = true
	}

	r.changeStatus(operation, &gqlschema.RuntimeIDInput{ID: runtimeID})

	return &gqlschema.OperationStatus{
		Operation: operation.lastOperation,
		State:     operation.status,
		Message:   "",
	}, nil
}

func (r *MockResolver) checkIfFinished(runtimeID *gqlschema.RuntimeIDInput) (string, bool) {
	item, exists := r.cache.Get(runtimeID.ID)

	if !exists {
		return "", true
	}

	operation, ok := item.(currentOperation)

	if !ok {
		return "", true
	}

	if operation.status == gqlschema.OperationStateSucceeded {
		return operation.operationID, true
	}
	return operation.operationID, false
}

func (r *MockResolver) changeStatus(operation currentOperation, runtimeID *gqlschema.RuntimeIDInput) {
	r.cache.Set(runtimeID.ID, operation, 0)
}

func (r *MockResolver) getStatus(runtimeID *gqlschema.RuntimeIDInput) (currentOperation, bool) {
	item, exists := r.cache.Get(runtimeID.ID)

	if !exists {
		return currentOperation{}, false
	}

	return item.(currentOperation), true
}

func (r *MockResolver) checkOperation(id *gqlschema.AsyncOperationIDInput) (currentOperation, string, bool) {
	for runtimeID, item := range r.cache.Items() {
		operation, ok := item.Object.(currentOperation)
		if ok && operation.operationID == id.ID {
			return operation, runtimeID, true
		}
	}
	return currentOperation{}, "", false
}
