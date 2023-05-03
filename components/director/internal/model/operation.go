package model

import (
	"encoding/json"
	"time"
)

// Operation represents an Operation
type Operation struct {
	ID         string
	OpType     string
	Status     string
	Data       json.RawMessage
	Error      json.RawMessage
	Priority   int
	CreatedAt  *time.Time
	FinishedAt *time.Time
}

// OperationInput represents an OperationInput
type OperationInput struct {
	OpType     string
	Status     string
	Data       json.RawMessage
	Error      json.RawMessage
	Priority   int
	CreatedAt  *time.Time
	FinishedAt *time.Time
}

// ToOperation converts OperationInput to Operation
func (i *OperationInput) ToOperation(id string) *Operation {
	if i == nil {
		return nil
	}

	return &Operation{
		ID:         id,
		OpType:     i.OpType,
		Status:     i.Status,
		Data:       i.Data,
		Error:      i.Error,
		Priority:   i.Priority,
		CreatedAt:  i.CreatedAt,
		FinishedAt: i.FinishedAt,
	}
}
