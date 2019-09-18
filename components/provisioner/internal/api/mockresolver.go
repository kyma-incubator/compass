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
	operationID        string
	operationType      gqlschema.OperationType
	status             gqlschema.OperationState
	runtimeID          string
	shouldStatusChange bool
	startTime          time.Time
}

func NewMockResolver(cache cache.Cache) *MockResolver {
	return &MockResolver{cache: cache}
}

func (r *MockResolver) ProvisionRuntime(ctx context.Context, id string, config *gqlschema.ProvisionRuntimeInput) (string, error) {
	return r.startNewOperation(ctx, id, gqlschema.OperationTypeProvision)
}

func (r *MockResolver) UpgradeRuntime(ctx context.Context, id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return r.startNewOperation(ctx, id, gqlschema.OperationTypeUpgrade)
}

func (r *MockResolver) DeprovisionRuntime(ctx context.Context, id string) (string, error) {
	return r.startNewOperation(ctx, id, gqlschema.OperationTypeDeprovision)
}

func (r *MockResolver) ReconnectRuntimeAgent(ctx context.Context, id string) (string, error) {
	return r.startNewOperation(ctx, id, gqlschema.OperationTypeReconnectRuntime)
}

func (r *MockResolver) startNewOperation(ctx context.Context, runtimeID string, operationType gqlschema.OperationType) (string, error) {
	if operationType != gqlschema.OperationTypeProvision {
		if !r.runtimeExists(runtimeID) {
			return "", errors.Errorf("Runtime %s does not exist", runtimeID)
		}
	}
	currentID, finished := r.checkIfLastOperationFinished(runtimeID)
	if !finished {
		return "", errors.Errorf("Cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := runtimeOperation{
		operationType: operationType,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
		runtimeID:     runtimeID,
		startTime:     time.Now(),
	}

	r.changeStatus(operation)
	return operationID, nil
}

func (r *MockResolver) RuntimeStatus(ctx context.Context, runtimeID string) (*gqlschema.RuntimeStatus, error) {
	operation, exists := r.getLastOperationStatus(runtimeID)

	if !exists {
		return nil, errors.Errorf("Runtime %s does not exist", runtimeID)
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

func (r *MockResolver) RuntimeOperationStatus(ctx context.Context, operationID string) (*gqlschema.OperationStatus, error) {
	operation, exists := r.checkOperation(operationID)

	if !exists {
		return nil, errors.Errorf("Operation: %s does not exist", operationID)
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

func (r *MockResolver) checkIfLastOperationFinished(runtimeID string) (string, bool) {
	for _, item := range r.cache.Items() {
		operation, ok := item.Object.(runtimeOperation)
		if !ok {
			continue
		}
		if operation.runtimeID == runtimeID && operation.status == gqlschema.OperationStateInProgress {
			return operation.operationID, false
		}
	}
	return "", true
}

func (r *MockResolver) changeStatus(operation runtimeOperation) {
	r.cache.Set(operation.operationID, operation, 0)
}

func (r *MockResolver) getLastOperationStatus(runtimeID string) (runtimeOperation, bool) {
	operationsMatchingRuntime := r.getRuntimeOperationsSorted(runtimeID)

	if len(operationsMatchingRuntime) != 0 {
		return operationsMatchingRuntime[0], true
	}

	return runtimeOperation{}, false
}

func (r *MockResolver) runtimeExists(runtimeID string) bool {
	operationsMatchingRuntime := r.getRuntimeOperationsSorted(runtimeID)

	for _, runtimeOpt := range operationsMatchingRuntime {
		if runtimeOpt.operationType == gqlschema.OperationTypeDeprovision && runtimeOpt.status == gqlschema.OperationStateSucceeded {
			return false
		}
		if runtimeOpt.operationType == gqlschema.OperationTypeProvision && runtimeOpt.status == gqlschema.OperationStateSucceeded {
			return true
		}
	}
	return false
}

func (r *MockResolver) getRuntimeOperationsSorted(runtimeID string) []runtimeOperation {
	operationsMatchingRuntime := make([]runtimeOperation, 0)
	for _, item := range r.cache.Items() {
		operation, ok := item.Object.(runtimeOperation)
		if !ok {
			continue
		}
		if operation.runtimeID == runtimeID {
			operationsMatchingRuntime = append(operationsMatchingRuntime, operation)
		}
	}
	sort.Slice(operationsMatchingRuntime, func(first, second int) bool {
		return operationsMatchingRuntime[first].startTime.After(operationsMatchingRuntime[second].startTime)
	})

	return operationsMatchingRuntime
}

func (r *MockResolver) checkOperation(operationID string) (runtimeOperation, bool) {
	item, exists := r.cache.Get(operationID)

	if exists {
		return item.(runtimeOperation), true
	}
	return runtimeOperation{}, false
}
