package formationassignment

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
