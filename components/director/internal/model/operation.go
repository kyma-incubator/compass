package model

import (
	"encoding/json"
	"time"
)

// OperationStatus defines operation status
type OperationStatus string

const (
	// OperationStatusScheduled scheduled operation status
	OperationStatusScheduled OperationStatus = "SCHEDULED"
	// OperationStatusInProgress in progress operation status
	OperationStatusInProgress OperationStatus = "IN_PROGRESS"
	// OperationStatusCompleted completed operation status
	OperationStatusCompleted OperationStatus = "COMPLETED"
	// OperationStatusFailed failed operation status
	OperationStatusFailed OperationStatus = "FAILED"
)

// OperationType defines supported operation types
type OperationType string

const (
	// OperationTypeOrdAggregation specifies open resource discovery operation type
	OperationTypeOrdAggregation OperationType = "ORD_AGGREGATION"
)

// Operation represents an Operation
type Operation struct {
	ID        string
	OpType    OperationType
	Status    OperationStatus
	Data      json.RawMessage
	Error     json.RawMessage
	Priority  int
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// OperationInput represents an OperationInput
type OperationInput struct {
	OpType    OperationType
	Status    OperationStatus
	Data      json.RawMessage
	Error     json.RawMessage
	Priority  int
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// ToOperation converts OperationInput to Operation
func (i *OperationInput) ToOperation(id string) *Operation {
	if i == nil {
		return nil
	}

	return &Operation{
		ID:        id,
		OpType:    i.OpType,
		Status:    i.Status,
		Data:      i.Data,
		Error:     i.Error,
		Priority:  i.Priority,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}
