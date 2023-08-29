package types

import "time"

const (
	// OperationStateSucceeded represents a successful operation
	OperationStateSucceeded OperationState = "succeeded"
	// OperationStateInProgress represents a operation that is in progress
	OperationStateInProgress OperationState = "in progress"
)

// OperationStatus holds info about the operation when a call is asynchronous
type OperationStatus struct {
	ID                  string               `json:"id"`
	Ready               bool                 `json:"ready"`
	Type                string               `json:"type"`
	State               OperationState       `json:"state"`
	ResourceID          string               `json:"resource_id"`
	ResourceType        string               `json:"resource_type"`
	PlatformId          string               `json:"platform_id"`
	CorrelationID       string               `json:"correlation_id"`
	Reschedule          bool                 `json:"reschedule"`
	RescheduleTimestamp time.Time            `json:"reschedule_timestamp"`
	DeletionScheduled   time.Time            `json:"deletion_scheduled"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	Errors              OperationStatusError `json:"errors"`
}

// OperationStatusError holds info about the operation error when a call is asynchronous
type OperationStatusError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

// OperationState represents the state of an Operation
type OperationState string
