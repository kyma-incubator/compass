package operation

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type converter struct {
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of an Operation.
func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an Operation to the service-layer representation model.Operation.
func (c *converter) FromEntity(entity *Entity) *model.Operation {
	return &model.Operation{
		ID:         entity.ID,
		OpType:     entity.Type,
		Status:     model.OperationStatus(entity.Status),
		Data:       repo.JSONRawMessageFromNullableString(entity.Data),
		Error:      repo.JSONRawMessageFromNullableString(entity.Error),
		Priority:   entity.Priority,
		CreatedAt:  entity.CreatedAt,
		FinishedAt: entity.FinishedAt,
	}
}

// ToEntity converts the provided service-layer representation of an Operation to the repository-layer one.
func (c *converter) ToEntity(operationModel *model.Operation) *Entity {
	return &Entity{
		ID:         operationModel.ID,
		Type:       operationModel.OpType,
		Status:     string(operationModel.Status),
		Data:       repo.NewNullableStringFromJSONRawMessage(operationModel.Data),
		Error:      repo.NewNullableStringFromJSONRawMessage(operationModel.Error),
		Priority:   operationModel.Priority,
		CreatedAt:  operationModel.CreatedAt,
		FinishedAt: operationModel.FinishedAt,
	}
}
