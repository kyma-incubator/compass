package operation

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
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
		ID:            entity.ID,
		OpType:        model.OperationType(entity.Type),
		Status:        model.OperationStatus(entity.Status),
		Data:          repo.JSONRawMessageFromNullableString(entity.Data),
		Error:         repo.JSONRawMessageFromNullableString(entity.Error),
		ErrorSeverity: model.OperationErrorSeverity(repo.StringFromNullableString(entity.ErrorSeverity)),
		Priority:      entity.Priority,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}

// ToEntity converts the provided service-layer representation of an Operation to the repository-layer one.
func (c *converter) ToEntity(operationModel *model.Operation) *Entity {
	return &Entity{
		ID:            operationModel.ID,
		Type:          string(operationModel.OpType),
		Status:        string(operationModel.Status),
		Data:          repo.NewNullableStringFromJSONRawMessage(operationModel.Data),
		Error:         repo.NewNullableStringFromJSONRawMessage(operationModel.Error),
		ErrorSeverity: repo.NewValidNullableString(string(operationModel.ErrorSeverity)),
		Priority:      operationModel.Priority,
		CreatedAt:     operationModel.CreatedAt,
		UpdatedAt:     operationModel.UpdatedAt,
	}
}

// ToGraphQL converts the provided service-layer representation of an Operation to the graphql-layer one.
func (c *converter) ToGraphQL(in *model.Operation) (*graphql.Operation, error) {
	if in == nil {
		return nil, nil
	}

	opType, err := c.operationTypeModelToGraphQL(in.OpType)
	if err != nil {
		return nil, err
	}

	opStatus, err := c.operationStatusModelToGraphQL(in.Status)
	if err != nil {
		return nil, err
	}

	opErrorSeverity := c.operationErrorSeverityToGraphQL(in.ErrorSeverity)

	return &graphql.Operation{
		ID:            in.ID,
		OperationType: opType,
		Status:        opStatus,
		Error:         str.StringifyJSONRawMessage(in.Error),
		ErrorSeverity: &opErrorSeverity,
		CreatedAt:     graphql.TimePtrToGraphqlTimestampPtr(in.CreatedAt),
		UpdatedAt:     graphql.TimePtrToGraphqlTimestampPtr(in.UpdatedAt),
	}, nil
}

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.Operation) ([]*graphql.Operation, error) {
	operations := make([]*graphql.Operation, 0, len(in))
	for _, o := range in {
		if o == nil {
			continue
		}
		operation, err := c.ToGraphQL(o)
		if err != nil {
			return nil, errors.Wrap(err, "while converting Operation to GraphQL")
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

func (c *converter) operationTypeModelToGraphQL(in model.OperationType) (graphql.ScheduledOperationType, error) {
	switch in {
	case model.OperationTypeSystemFetching:
		return graphql.ScheduledOperationTypeSystemFetching, nil
	case model.OperationTypeOrdAggregation:
		return graphql.ScheduledOperationTypeOrdAggregation, nil
	default:
		return "", errors.Errorf("unknown operation type %v", in)
	}
}

func (c *converter) operationStatusModelToGraphQL(in model.OperationStatus) (graphql.OperationStatus, error) {
	switch in {
	case model.OperationStatusScheduled:
		return graphql.OperationStatusScheduled, nil
	case model.OperationStatusInProgress:
		return graphql.OperationStatusInProgress, nil
	case model.OperationStatusCompleted:
		return graphql.OperationStatusCompleted, nil
	case model.OperationStatusFailed:
		return graphql.OperationStatusFailed, nil
	default:
		return "", errors.Errorf("unknown operation status %v", in)
	}
}

func (c *converter) operationErrorSeverityToGraphQL(in model.OperationErrorSeverity) graphql.OperationErrorSeverity {
	switch in {
	case model.OperationErrorSeverityError:
		return graphql.OperationErrorSeverityError
	case model.OperationErrorSeverityWarning:
		return graphql.OperationErrorSeverityWarning
	case model.OperationErrorSeverityInfo:
		return graphql.OperationErrorSeverityInfo
	default:
		return ""
	}
}
