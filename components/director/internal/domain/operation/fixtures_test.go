package operation_test

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"

	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const (
	testOpType            model.OperationType = "TEST_TYPE"
	errorMsg              string              = "error message"
	operationID           string              = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	lowOperationPriority  int                 = 1
	highOperationPriority int                 = 2
	operationErrMsg       string              = "operation processing failed"
	queueLimit            int                 = 10
	applicationID         string              = "1c992715-5066-491f-8cbf-9d0fab4cd61c"
	applicationTemplateID string              = "a24f7aea-4da1-4caa-9716-76f6f6551b01"
)

var (
	fixColumns = []string{"id", "op_type", "status", "data", "error", "priority", "created_at", "updated_at"}
)

func fixOperationData(appID, appTemplateID string) interface{} {
	return ord.NewOrdOperationData(appID, appTemplateID)
}

func fixOperationDataAsString(appID, appTemplateID string) string {
	result, _ := json.Marshal(fixOperationData(appID, appTemplateID))
	return string(result)
}

func fixOperationInput(opType model.OperationType, opStatus model.OperationStatus) *model.OperationInput {
	return &model.OperationInput{
		OpType:    opType,
		Status:    opStatus,
		Data:      json.RawMessage(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:     json.RawMessage(errorMsg),
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
		Data:      json.RawMessage(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:     json.RawMessage(errorMsg),
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
		Data:      repo.NewValidNullableString(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:     repo.NewValidNullableString(errorMsg),
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
