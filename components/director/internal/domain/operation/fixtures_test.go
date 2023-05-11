package operation_test

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"

	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const (
	ordOpType         = "ORD_AGGREGATION"
	scheduledOpStatus = "SCHEDULED"
	operationID       = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
)

func fixOperationInput(opType, opStatus string) *model.OperationInput {
	return &model.OperationInput{
		OpType:     opType,
		Status:     opStatus,
		Data:       json.RawMessage("[]"),
		Error:      json.RawMessage("[]"),
		Priority:   1,
		CreatedAt:  &time.Time{},
		FinishedAt: &time.Time{},
	}
}

func fixOperationModel(opType, status string) *model.Operation {
	return fixOperationModelWithID(operationID, opType, status)
}

func fixOperationModelWithID(id, opType, opStatus string) *model.Operation {
	return &model.Operation{
		ID:         id,
		OpType:     opType,
		Status:     opStatus,
		Data:       json.RawMessage("[]"),
		Error:      json.RawMessage("[]"),
		Priority:   1,
		CreatedAt:  &time.Time{},
		FinishedAt: &time.Time{},
	}
}

func fixEntityOperation(id, opType, opStatus string) *operation.Entity {
	return &operation.Entity{
		ID:         id,
		Type:       opType,
		Status:     opStatus,
		Data:       repo.NewValidNullableString("[]"),
		Error:      repo.NewValidNullableString("[]"),
		Priority:   1,
		CreatedAt:  &time.Time{},
		FinishedAt: &time.Time{},
	}
}

func fixOperationCreateArgs(op *model.Operation) []driver.Value {
	return []driver.Value{op.ID, op.OpType, op.Status, repo.NewNullableStringFromJSONRawMessage(op.Data), repo.NewNullableStringFromJSONRawMessage(op.Error), op.Priority, op.CreatedAt, op.FinishedAt}
}
