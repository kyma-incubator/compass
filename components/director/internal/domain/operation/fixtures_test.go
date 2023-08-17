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
	ordOpType   = "ORD_AGGREGATION"
	operationID = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
)

var (
	fixColumns = []string{"id", "op_type", "status", "data", "error", "priority", "created_at", "updated_at"}
)

func fixOperationInput(opType string, opStatus model.OperationStatus) *model.OperationInput {
	return &model.OperationInput{
		OpType:    opType,
		Status:    opStatus,
		Data:      json.RawMessage("[]"),
		Error:     json.RawMessage("[]"),
		Priority:  1,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}
}

func fixOperationModel(opType string, status model.OperationStatus) *model.Operation {
	return fixOperationModelWithID(operationID, opType, status)
}

func fixOperationModelWithID(id, opType string, opStatus model.OperationStatus) *model.Operation {
	return &model.Operation{
		ID:        id,
		OpType:    opType,
		Status:    opStatus,
		Data:      json.RawMessage("[]"),
		Error:     json.RawMessage("[]"),
		Priority:  1,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}
}

func fixEntityOperation(id, opType string, opStatus model.OperationStatus) *operation.Entity {
	return &operation.Entity{
		ID:        id,
		Type:      opType,
		Status:    string(opStatus),
		Data:      repo.NewValidNullableString("[]"),
		Error:     repo.NewValidNullableString("[]"),
		Priority:  1,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}
}

func fixOperationCreateArgs(op *model.Operation) []driver.Value {
	return []driver.Value{op.ID, op.OpType, op.Status, repo.NewNullableStringFromJSONRawMessage(op.Data), repo.NewNullableStringFromJSONRawMessage(op.Error), op.Priority, op.CreatedAt, op.UpdatedAt}
}

func fixOperationUpdateArgs(op *model.Operation) []driver.Value {
	return []driver.Value{op.Status, repo.NewNullableStringFromJSONRawMessage(op.Error), op.Priority, op.UpdatedAt, op.ID}
}
