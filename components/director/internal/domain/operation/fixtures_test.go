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
	testOpType            model.OperationType = "TEST_TYPE"
	operationID           string              = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	lowOperationPriority  int                 = 1
	highOperationPriority int                 = 2
	operationErrMsg       string              = "operation processing failed"
	queueLimit            int                 = 10
	data                  string              = "{\"applicationID\": \"1c992715-5066-491f-8cbf-9d0fab4cd61c\",  \"applicationTemplateID\": \"a24f7aea-4da1-4caa-9716-76f6f6551b01\"}"
)

var (
	fixColumns = []string{"id", "op_type", "status", "data", "error", "priority", "created_at", "updated_at"}
)

func fixOperationInput(opType model.OperationType, opStatus model.OperationStatus) *model.OperationInput {
	return &model.OperationInput{
		OpType:    opType,
		Status:    opStatus,
		Data:      json.RawMessage(data),
		Error:     json.RawMessage(data),
		Priority:  1,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}
}

func fixOperationModel(opType model.OperationType, status model.OperationStatus) *model.Operation {
	return fixOperationModelWithID(operationID, opType, status, lowOperationPriority)
}

func fixOperationModelWithPriority(opType model.OperationType, status model.OperationStatus, priority int) *model.Operation {
	return fixOperationModelWithID(operationID, opType, status, priority)
}

func fixOperationModelWithID(id string, opType model.OperationType, opStatus model.OperationStatus, priority int) *model.Operation {
	return &model.Operation{
		ID:        id,
		OpType:    opType,
		Status:    opStatus,
		Data:      json.RawMessage(data),
		Error:     json.RawMessage(data),
		Priority:  priority,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}
}

func fixEntityOperation(id string, opType model.OperationType, opStatus model.OperationStatus) *operation.Entity {
	return &operation.Entity{
		ID:        id,
		Type:      string(opType),
		Status:    string(opStatus),
		Data:      repo.NewValidNullableString(data),
		Error:     repo.NewValidNullableString(data),
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
