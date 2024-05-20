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

// OperationErrorSeverity defines operation's error severity
type OperationErrorSeverity string

const (
	// OperationErrorSeverityError represents Error severity
	OperationErrorSeverityError OperationErrorSeverity = "ERROR"
	// OperationErrorSeverityWarning represents Warning severity
	OperationErrorSeverityWarning OperationErrorSeverity = "WARNING"
	// OperationErrorSeverityInfo represents Info severity
	OperationErrorSeverityInfo OperationErrorSeverity = "INFO"
	// OperationErrorSeverityNone represents missing severity
	OperationErrorSeverityNone OperationErrorSeverity = "NONE"
)

// OperationType defines supported operation types
type OperationType string

const (
	// OperationTypeOrdAggregation specifies open resource discovery operation type
	OperationTypeOrdAggregation OperationType = "ORD_AGGREGATION"
	// OperationTypeSystemFetching specifies system fetching operation type
	OperationTypeSystemFetching OperationType = "SYSTEM_FETCHING"
)

// Operation represents an Operation
type Operation struct {
	ID            string
	OpType        OperationType
	Status        OperationStatus
	Data          json.RawMessage
	Error         json.RawMessage
	ErrorSeverity OperationErrorSeverity
	Priority      int
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

// OperationInput represents an OperationInput
type OperationInput struct {
	OpType        OperationType
	Status        OperationStatus
	Data          json.RawMessage
	Error         json.RawMessage
	ErrorSeverity OperationErrorSeverity
	Priority      int
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

// ToOperation converts OperationInput to Operation
func (i *OperationInput) ToOperation(id string) *Operation {
	if i == nil {
		return nil
	}

	return &Operation{
		ID:            id,
		OpType:        i.OpType,
		Status:        i.Status,
		Data:          i.Data,
		Error:         i.Error,
		ErrorSeverity: i.ErrorSeverity,
		Priority:      i.Priority,
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
	}
}
