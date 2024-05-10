package assignmentOperation_test

import (
	assignmentOperation "github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"time"
)

const ()

var (
	operationID                 = "operationID"
	operationID2                = "operationID2"
	assignmentID                = "assignmentID"
	formationID                 = "formationID"
	operationType               = model.Assign
	operationTrigger            = model.AssignObject
	newOperationTrigger         = model.ResyncAssignment
	fixColumns                  = []string{"id", "type", "formation_assignment_id", "formation_id", "triggered_by", "started_at_timestamp", "finished_at_timestamp"}
	nilAssignmentOperationModel *model.AssignmentOperation
	defaultTime                 = time.Time{}
)

func fixAssignmentOperationInput() *model.AssignmentOperationInput {
	return &model.AssignmentOperationInput{
		Type:                  operationType,
		FormationAssignmentID: assignmentID,
		FormationID:           formationID,
		TriggeredBy:           operationTrigger,
	}
}

func fixAssignmentOperationModel() *model.AssignmentOperation {
	return fixAssignmentOperationModelWithAssignmentID(assignmentID)
}

func fixAssignmentOperationModelWithAssignmentID(id string) *model.AssignmentOperation {
	return &model.AssignmentOperation{
		ID:                    operationID,
		Type:                  operationType,
		FormationAssignmentID: id,
		FormationID:           formationID,
		TriggeredBy:           operationTrigger,
		StartedAtTimestamp:    &defaultTime,
		FinishedAtTimestamp:   &defaultTime,
	}
}

func fixAssignmentOperationModelWithoutFinishedAt() *model.AssignmentOperation {
	return &model.AssignmentOperation{
		ID:                    operationID,
		Type:                  operationType,
		FormationAssignmentID: assignmentID,
		FormationID:           formationID,
		TriggeredBy:           operationTrigger,
		StartedAtTimestamp:    &defaultTime,
	}
}

func fixAssignmentOperationEntity() *assignmentOperation.Entity {
	return fixAssignmentOperationEntityWithAssignmentID(assignmentID)
}

func fixAssignmentOperationEntityWithAssignmentID(id string) *assignmentOperation.Entity {
	return &assignmentOperation.Entity{
		ID:                    operationID,
		Type:                  string(model.Assign),
		FormationAssignmentID: id,
		FormationID:           formationID,
		TriggeredBy:           string(model.AssignObject),
		StartedAtTimestamp:    &defaultTime,
		FinishedAtTimestamp:   &defaultTime,
	}
}

func fixAssignmentOperationGQL() *graphql.AssignmentOperation {
	return &graphql.AssignmentOperation{
		ID:                    operationID,
		OperationType:         graphql.AssignmentOperationType(model.Assign),
		FormationAssignmentID: assignmentID,
		FormationID:           formationID,
		TriggeredBy:           graphql.OperationTrigger(model.AssignObject),
		StartedAtTimestamp:    graphql.TimePtrToGraphqlTimestampPtr(&defaultTime),
		FinishedAtTimestamp:   graphql.TimePtrToGraphqlTimestampPtr(&defaultTime),
	}
}

func fixUUIDService() *automock.UIDService {
	uidSvc := &automock.UIDService{}
	uidSvc.On("Generate").Return(operationID)
	return uidSvc
}

func fixAssignmentOperationPage() *model.AssignmentOperationPage {
	return &model.AssignmentOperationPage{
		Data: []*model.AssignmentOperation{
			fixAssignmentOperationModel(),
		},
	}
}
