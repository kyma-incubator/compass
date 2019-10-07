package api

import (
	"context"
	"sort"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type MockResolver struct {
	repository map[string]RuntimeOperation
}

func (r *MockResolver) Mutation() gqlschema.MutationResolver {
	return &MockResolver{r.repository}
}
func (r *MockResolver) Query() gqlschema.QueryResolver {
	return &MockResolver{r.repository}
}

type RuntimeOperation struct {
	operationID   string
	operationType gqlschema.OperationType
	status        gqlschema.OperationState
	runtimeID     string
	succeeded     bool
	startTime     time.Time
}

func NewMockResolver(repository map[string]RuntimeOperation) *MockResolver {
	return &MockResolver{repository: repository}
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
			return "", errors.Errorf("runtime %s does not exist", runtimeID)
		}
	}
	currentID, finished := r.checkIfLastOperationFinished(runtimeID)
	if !finished {
		return "", errors.Errorf("cannot start new operation while previous one is not finished yet. Current operation: %s", currentID)
	}

	operationID := string(uuid.NewUUID())

	operation := RuntimeOperation{
		operationType: operationType,
		status:        gqlschema.OperationStateInProgress,
		operationID:   operationID,
		runtimeID:     runtimeID,
		startTime:     time.Now(),
	}

	r.save(operation)
	return operationID, nil
}

func (r *MockResolver) RuntimeStatus(ctx context.Context, runtimeID string) (*gqlschema.RuntimeStatus, error) {
	operation, exists := r.getLastOperationStatus(runtimeID)

	if !exists {
		return nil, errors.Errorf("runtime %s does not exist", runtimeID)
	}

	kubeconfig := "kubeconfig"

	return &gqlschema.RuntimeStatus{
		LastOperationStatus: &gqlschema.OperationStatus{
			Operation: operation.operationType,
			State:     operation.status,
		},
		RuntimeConnectionStatus: &gqlschema.RuntimeConnectionStatus{
			Status: gqlschema.RuntimeAgentConnectionStatusConnected,
		},
		RuntimeConfiguration: &gqlschema.RuntimeConfig{
			ClusterConfig: &gqlschema.GardenerConfig{},
			KymaConfig:    &gqlschema.KymaConfig{},
			Kubeconfig:    &kubeconfig,
		},
	}, nil
}

/* Runtime Operation Status always returns status set in operation call (usually In Progress) in first call after starting new operation
and status Succeeded in second and following calls until next operation is started.
*/

func (r *MockResolver) RuntimeOperationStatus(ctx context.Context, operationID string) (*gqlschema.OperationStatus, error) {
	operation, exists := r.load(operationID)

	if !exists {
		return nil, errors.Errorf("operation: %s does not exist", operationID)
	}

	if operation.succeeded {
		operation.status = gqlschema.OperationStateSucceeded
	} else {
		operation.succeeded = true
	}

	r.save(operation)

	return &gqlschema.OperationStatus{
		Operation: operation.operationType,
		State:     operation.status,
		RuntimeID: operation.runtimeID,
		Message:   "",
	}, nil
}

func (r *MockResolver) checkIfLastOperationFinished(runtimeID string) (string, bool) {
	for _, operation := range r.repository {
		if operation.runtimeID == runtimeID && operation.status == gqlschema.OperationStateInProgress {
			return operation.operationID, false
		}
	}
	return "", true
}

func (r *MockResolver) save(operation RuntimeOperation) {
	r.repository[operation.operationID] = operation
}

func (r *MockResolver) getLastOperationStatus(runtimeID string) (RuntimeOperation, bool) {
	operationsMatchingRuntime := r.getRuntimeOperationsSorted(runtimeID)

	if len(operationsMatchingRuntime) != 0 {
		return operationsMatchingRuntime[0], true
	}

	return RuntimeOperation{}, false
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

func (r *MockResolver) getRuntimeOperationsSorted(runtimeID string) []RuntimeOperation {
	operationsMatchingRuntime := make([]RuntimeOperation, 0)
	for _, operation := range r.repository {
		if operation.runtimeID == runtimeID {
			operationsMatchingRuntime = append(operationsMatchingRuntime, operation)
		}
	}
	sort.Slice(operationsMatchingRuntime, func(first, second int) bool {
		return operationsMatchingRuntime[first].startTime.After(operationsMatchingRuntime[second].startTime)
	})

	return operationsMatchingRuntime
}

func (r *MockResolver) load(operationID string) (RuntimeOperation, bool) {
	item, exists := r.repository[operationID]

	if exists {
		return item, true
	}
	return RuntimeOperation{}, false
}

func (r *MockResolver) CleanupRuntimeData(ctx context.Context, id string) (string, error) {
	return id, nil
}
