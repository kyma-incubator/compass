package assignmentoperation

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// NewConverter creates a new assignment operation converter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.AssignmentOperation) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:                    in.ID,
		Type:                  string(in.Type),
		FormationAssignmentID: in.FormationAssignmentID,
		FormationID:           in.FormationID,
		TriggeredBy:           string(in.TriggeredBy),
		StartedAtTimestamp:    in.StartedAtTimestamp,
		FinishedAtTimestamp:   in.FinishedAtTimestamp,
	}
}

func (c *converter) FromEntity(e *Entity) *model.AssignmentOperation {
	if e == nil {
		return nil
	}

	return &model.AssignmentOperation{
		ID:                    e.ID,
		Type:                  model.AssignmentOperationType(e.Type),
		FormationAssignmentID: e.FormationAssignmentID,
		FormationID:           e.FormationID,
		TriggeredBy:           model.OperationTrigger(e.TriggeredBy),
		StartedAtTimestamp:    e.StartedAtTimestamp,
		FinishedAtTimestamp:   e.FinishedAtTimestamp,
	}
}

// ToGraphQL converts from internal model to graphql model
func (c *converter) ToGraphQL(in *model.AssignmentOperation) *gql.AssignmentOperation {
	if in == nil {
		return nil
	}

	return &gql.AssignmentOperation{
		ID:                    in.ID,
		OperationType:         gql.AssignmentOperationType(in.Type),
		FormationAssignmentID: in.FormationAssignmentID,
		FormationID:           in.FormationID,
		TriggeredBy:           gql.OperationTrigger(in.TriggeredBy),
		StartedAtTimestamp:    (*gql.Timestamp)(in.StartedAtTimestamp),
		FinishedAtTimestamp:   (*gql.Timestamp)(in.FinishedAtTimestamp),
	}
}

// MultipleToGraphQL converts multiple entities from internal models to graphql models
func (c *converter) MultipleToGraphQL(in []*model.AssignmentOperation) []*gql.AssignmentOperation {
	var assignmentOperations []*gql.AssignmentOperation
	for _, r := range in {
		if r == nil {
			continue
		}

		ao := c.ToGraphQL(r)

		assignmentOperations = append(assignmentOperations, ao)
	}

	return assignmentOperations
}
