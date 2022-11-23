package formationassignment

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// NewConverter creates a new formation assignment converter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in *model.FormationAssignment) (*graphql.FormationAssignment, error) {
	if in == nil {
		return nil, nil
	}

	var strValue *string
	if in.Value != nil {
		marshalledValue, err := json.Marshal(in.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while converting formation assignment to GraphQL")
		}
		strValue = str.Ptr(string(marshalledValue))
	}

	return &graphql.FormationAssignment{
		ID:                         in.ID,
		Source:                     in.Source,
		SourceType:                 graphql.FormationAssignmentType(in.SourceType),
		Target:                     in.Target,
		TargetType:                 graphql.FormationAssignmentType(in.TargetType),
		LastOperation:              string(in.LastOperation),
		LastOperationInitiator:     in.LastOperationInitiator,
		LastOperationInitiatorType: graphql.FormationAssignmentType(in.LastOperationInitiatorType),
		State:                      in.State,
		Value:                      strValue,
	}, nil
}

// MultipleToGraphQL converts multiple internal models to GraphQL models
func (c *converter) MultipleToGraphQL(in []*model.FormationAssignment) ([]*graphql.FormationAssignment, error) {
	if in == nil {
		return nil, nil
	}
	formationAssignment := make([]*graphql.FormationAssignment, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		fa, err := c.ToGraphQL(r)
		if err != nil {
			return nil, err
		}

		formationAssignment = append(formationAssignment, fa)
	}

	return formationAssignment, nil
}

// ToInput converts from internal model to internal model input
func (c *converter) ToInput(assignment *model.FormationAssignment) *model.FormationAssignmentInput {
	if assignment == nil {
		return nil
	}

	return &model.FormationAssignmentInput{
		FormationID:                assignment.FormationID,
		Source:                     assignment.Source,
		SourceType:                 assignment.SourceType,
		Target:                     assignment.Target,
		TargetType:                 assignment.TargetType,
		LastOperation:              assignment.LastOperation,
		LastOperationInitiator:     assignment.LastOperationInitiator,
		LastOperationInitiatorType: assignment.LastOperationInitiatorType,
		State:                      assignment.State,
		Value:                      assignment.Value,
	}
}

// FromInput converts from internal model input to internal model
func (c *converter) FromInput(in *model.FormationAssignmentInput) *model.FormationAssignment {
	if in == nil {
		return nil
	}

	return &model.FormationAssignment{
		FormationID:                in.FormationID,
		Source:                     in.Source,
		SourceType:                 in.SourceType,
		Target:                     in.Target,
		TargetType:                 in.TargetType,
		LastOperation:              in.LastOperation,
		LastOperationInitiator:     in.LastOperationInitiator,
		LastOperationInitiatorType: in.LastOperationInitiatorType,
		State:                      in.State,
		Value:                      in.Value,
	}
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationAssignment) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:                         in.ID,
		FormationID:                in.FormationID,
		TenantID:                   in.TenantID,
		Source:                     in.Source,
		SourceType:                 string(in.SourceType),
		Target:                     in.Target,
		TargetType:                 string(in.TargetType),
		LastOperation:              string(in.LastOperation),
		LastOperationInitiator:     in.LastOperationInitiator,
		LastOperationInitiatorType: string(in.LastOperationInitiatorType),
		State:                      in.State,
		Value:                      repo.NewNullableStringFromJSONRawMessage(in.Value),
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.FormationAssignment {
	if e == nil {
		return nil
	}

	return &model.FormationAssignment{
		ID:                         e.ID,
		FormationID:                e.FormationID,
		TenantID:                   e.TenantID,
		Source:                     e.Source,
		SourceType:                 model.FormationAssignmentType(e.SourceType),
		Target:                     e.Target,
		TargetType:                 model.FormationAssignmentType(e.TargetType),
		LastOperation:              model.FormationOperation(e.LastOperation),
		LastOperationInitiator:     e.LastOperationInitiator,
		LastOperationInitiatorType: model.FormationAssignmentType(e.LastOperationInitiatorType),
		State:                      e.State,
		Value:                      repo.JSONRawMessageFromNullableString(e.Value),
	}
}
