package operationsmanager

import "sync"

// OperationType defines supported operation types
type OperationType string

const (
	// OrdAggregationOpType specifies open resource discovery operation type
	OrdAggregationOpType OperationType = "ORD_AGGREGATION"
)

// OperationsManager provides methods for operations management
type OperationsManager struct {
	opType OperationType
	mutex  sync.Mutex
}

// NewOperationsManager creates new OperationsManager
func NewOperationsManager(opType OperationType) *OperationsManager {
	return &OperationsManager{
		opType: opType,
	}
}

// GetOperation retrieves one scheduled operation
func (om *OperationsManager) GetOperation() {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	//TODO implement me
}

// MarkOperationCompleted marks the operation with the given ID as completed
func (om *OperationsManager) MarkOperationCompleted(id string) {
	//TODO implement me
}

// MarkOperationFailed marks the operation with the given ID as failed
func (om *OperationsManager) MarkOperationFailed(id, err string) {
	//TODO implement me
}
