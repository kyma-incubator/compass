package operation_test

import (
	"database/sql/driver"
	"encoding/json"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/data"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"

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
	fixColumns = []string{"id", "op_type", "status", "data", "error", "error_severity", "priority", "created_at", "updated_at"}
)

func fixOperationData(appID, appTemplateID string) interface{} {
	return ord.NewOrdOperationData(appID, appTemplateID)
}

func fixOperationDataAsString(appID, appTemplateID string) string {
	result, _ := json.Marshal(fixOperationData(appID, appTemplateID))
	return string(result)
}

func fixOperationInput(opType model.OperationType, opStatus model.OperationStatus, errSeverity model.OperationErrorSeverity) *model.OperationInput {
	return &model.OperationInput{
		OpType:        opType,
		Status:        opStatus,
		Data:          json.RawMessage(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:         json.RawMessage(errorMsg),
		ErrorSeverity: errSeverity,
		Priority:      1,
		CreatedAt:     &time.Time{},
		UpdatedAt:     &time.Time{},
	}
}

func fixOperationModel(opType model.OperationType, status model.OperationStatus, errSeverity model.OperationErrorSeverity) *model.Operation {
	return fixOperationModelWithID(operationID, opType, status, lowOperationPriority, errSeverity)
}

func fixOperationModelWithPriority(opType model.OperationType, status model.OperationStatus, priority int, errSeverity model.OperationErrorSeverity) *model.Operation {
	return fixOperationModelWithID(operationID, opType, status, priority, errSeverity)
}

func fixOperationModelWithID(id string, opType model.OperationType, opStatus model.OperationStatus, priority int, errSeverity model.OperationErrorSeverity) *model.Operation {
	return &model.Operation{
		ID:            id,
		OpType:        opType,
		Status:        opStatus,
		Data:          json.RawMessage(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:         json.RawMessage(errorMsg),
		ErrorSeverity: errSeverity,
		Priority:      priority,
		CreatedAt:     &time.Time{},
		UpdatedAt:     &time.Time{},
	}
}

func fixOperationModelWithIDAndTimestamp(id string, opType model.OperationType, opStatus model.OperationStatus, errorMsg string, priority int, errSeverity model.OperationErrorSeverity, timestamp *time.Time) *model.Operation {
	return &model.Operation{
		ID:            id,
		OpType:        opType,
		Status:        opStatus,
		Data:          json.RawMessage(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:         json.RawMessage(errorMsg),
		ErrorSeverity: errSeverity,
		Priority:      priority,
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
	}
}

func fixOperationModelWithErrorSeverity(errorSeverity model.OperationErrorSeverity) *model.Operation {
	return &model.Operation{
		ID:            operationID,
		OpType:        testOpType,
		Status:        model.OperationStatusFailed,
		Data:          json.RawMessage(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:         json.RawMessage(errorMsg),
		ErrorSeverity: errorSeverity,
		Priority:      1,
		CreatedAt:     &time.Time{},
		UpdatedAt:     &time.Time{},
	}
}

func fixOperationGraphqlWithIDAndTimestamp(id string, opType graphql.ScheduledOperationType, opStatus graphql.OperationStatus, errorMsg string, errSeverity graphql.OperationErrorSeverity, timestamp *time.Time) *graphql.Operation {
	return &graphql.Operation{
		ID:            id,
		OperationType: opType,
		Status:        opStatus,
		Error:         &errorMsg,
		ErrorSeverity: errSeverity,
		CreatedAt:     graphql.TimePtrToGraphqlTimestampPtr(timestamp),
		UpdatedAt:     graphql.TimePtrToGraphqlTimestampPtr(timestamp),
	}
}

func fixEntityOperation(id string, opType model.OperationType, opStatus model.OperationStatus, opErrorSeverity model.OperationErrorSeverity) *operation.Entity {
	return &operation.Entity{
		ID:            id,
		Type:          string(opType),
		Status:        string(opStatus),
		Data:          repo.NewValidNullableString(fixOperationDataAsString(applicationID, applicationTemplateID)),
		Error:         repo.NewValidNullableString(errorMsg),
		ErrorSeverity: string(opErrorSeverity),
		Priority:      1,
		CreatedAt:     &time.Time{},
		UpdatedAt:     &time.Time{},
	}
}

func fixOperationCreateArgs(op *model.Operation) []driver.Value {
	return []driver.Value{op.ID, op.OpType, op.Status, repo.NewNullableStringFromJSONRawMessage(op.Data), repo.NewNullableStringFromJSONRawMessage(op.Error), string(op.ErrorSeverity), op.Priority, op.CreatedAt, op.UpdatedAt}
}

func fixOperationUpdateArgs(op *model.Operation) []driver.Value {
	return []driver.Value{op.Status, repo.NewNullableStringFromJSONRawMessage(op.Error), string(op.ErrorSeverity), op.Priority, op.UpdatedAt, op.ID}
}
