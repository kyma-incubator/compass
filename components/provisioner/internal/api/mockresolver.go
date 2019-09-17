package api

import (
	"context"
	"sort"
	"time"

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

type runtimeOperation struct {
	operationType      gqlschema.OperationType
	status             gqlschema.OperationState
	operationID        string
	shouldStatusChange bool
	runtimeID          string
	startTime          time.Time
}

func NewMockResolver(cache cache.Cache) *MockResolver {
	return &MockResolver{cache: cache}
}

func (r *MockResolver) ProvisionRuntime(ctx context.Context, id *gqlschema.RuntimeIDInput, config *gqlschema.ProvisionRuntimeInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfLastOperationFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}
	operationID := string(uuid.NewUUID())

	operation := runtimeOperation{
		operationType: gqlschema.OperationTypeProvision,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
		runtimeID:     id.ID,
		startTime:     time.Now(),
	}

	r.changeStatus(operation)
	return &gqlschema.AsyncOperationID{ID: string(operationID)}, nil
}

func (r *MockResolver) UpgradeRuntime(ctx context.Context, id *gqlschema.RuntimeIDInput, config *gqlschema.UpgradeRuntimeInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfLastOperationFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := runtimeOperation{
		operationType: gqlschema.OperationTypeUpgrade,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
		runtimeID:     id.ID,
		startTime:     time.Now(),
	}

	r.changeStatus(operation)
	return &gqlschema.AsyncOperationID{ID: operationID}, nil
}

func (r *MockResolver) DeprovisionRuntime(ctx context.Context, id *gqlschema.RuntimeIDInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfLastOperationFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := runtimeOperation{
		operationType: gqlschema.OperationTypeDeprovision,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
		runtimeID:     id.ID,
		startTime:     time.Now(),
	}

	r.changeStatus(operation)
	return &gqlschema.AsyncOperationID{ID: operationID}, nil
}

func (r *MockResolver) ReconnectRuntimeAgent(ctx context.Context, id *gqlschema.RuntimeIDInput) (*gqlschema.AsyncOperationID, error) {
	currentID, finished := r.checkIfLastOperationFinished(id)
	if !finished {
		return nil, errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := runtimeOperation{
		operationType: gqlschema.OperationTypeReconnectRuntime,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
		runtimeID:     id.ID,
		startTime:     time.Now(),
	}

	r.changeStatus(operation)
	return &gqlschema.AsyncOperationID{ID: operationID}, nil
}

func (r *MockResolver) RuntimeStatus(ctx context.Context, id *gqlschema.RuntimeIDInput) (*gqlschema.RuntimeStatus, error) {
	operation, exists := r.getStatus(id)

	if !exists {
		return nil, errors.Errorf("Runtime %s does not exist", id.ID)
	}

	return &gqlschema.RuntimeStatus{
		LastOperationStatus: &gqlschema.OperationStatus{
			Operation: operation.operationType,
			State:     operation.status,
		},
		RuntimeConnectionStatus: &gqlschema.RuntimeConnectionStatus{
			Status: gqlschema.RuntimeAgentConnectionStatusConnected,
		},
		RuntimeConnectionConfig: &gqlschema.RuntimeConnectionConfig{
			Kubeconfig: "kubeconfig",
		},
		RuntimeConfiguration: &gqlschema.RuntimeConfig{
			ClusterConfig: &gqlschema.ClusterConfig{},
			KymaConfig:    &gqlschema.KymaConfig{},
		},
	}, nil
}

/* Runtime Operation Status always returns status set in operation call (usually In Progress) in first call after starting new operation
and status Succeeded in second and following calls until next operation is started.
*/

func (r *MockResolver) RuntimeOperationStatus(ctx context.Context, id *gqlschema.AsyncOperationIDInput) (*gqlschema.OperationStatus, error) {
	operation, exists := r.checkOperation(id)

	if !exists {
		return nil, errors.Errorf("Operation: %s does not exist", id)
	}

	if operation.shouldStatusChange {
		operation.status = gqlschema.OperationStateSucceeded
	} else {
		operation.shouldStatusChange = true
	}

	r.changeStatus(operation)

	return &gqlschema.OperationStatus{
		Operation: operation.operationType,
		State:     operation.status,
		Message:   "",
	}, nil
}

func (r *MockResolver) checkIfLastOperationFinished(runtimeID *gqlschema.RuntimeIDInput) (string, bool) {
	for _, item := range r.cache.Items() {
		operation := item.Object.(runtimeOperation)
		if operation.runtimeID == runtimeID.ID && operation.status == gqlschema.OperationStateInProgress {
			return operation.operationID, false
		}
	}
	return "", true
}

func (r *MockResolver) changeStatus(operation runtimeOperation) {
	r.cache.Set(operation.operationID, operation, 0)
}

func (r *MockResolver) getStatus(runtimeID *gqlschema.RuntimeIDInput) (runtimeOperation, bool) {
	operationsMatchingRuntime := []runtimeOperation{}
	for _, item := range r.cache.Items() {
		operation := item.Object.(runtimeOperation)
		if operation.runtimeID == runtimeID.ID {
			operationsMatchingRuntime = append(operationsMatchingRuntime, operation)
		}
	}

	if len(operationsMatchingRuntime) != 0 {
		sort.Slice(operationsMatchingRuntime, func(first, second int) bool {
			return operationsMatchingRuntime[first].startTime.After(operationsMatchingRuntime[second].startTime)
		})
		return operationsMatchingRuntime[0], true
	}

	return runtimeOperation{}, false
}

func (r *MockResolver) checkOperation(id *gqlschema.AsyncOperationIDInput) (runtimeOperation, bool) {
	item, exists := r.cache.Get(id.ID)

	if exists {
		return item.(runtimeOperation), true
	}
	return runtimeOperation{}, false
}
