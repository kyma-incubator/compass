package formationassignment

import (
	"encoding/json"

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
	marshalledValue, err := json.Marshal(in.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting formation assignment to GraphQL")
	}
	strValue := string(marshalledValue)

	return &graphql.FormationAssignment{
		ID:         in.ID,
		Source:     in.Source,
		SourceType: in.SourceType,
		Target:     in.Target,
		TargetType: in.TargetType,
		State:      in.State,
		Value:      &strValue,
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

func (c *converter) ToInput(assignment *model.FormationAssignment) *model.FormationAssignmentInput {
	if assignment == nil {
		return nil
	}

	return &model.FormationAssignmentInput{
		FormationID: assignment.FormationID,
		Source:      assignment.Source,
		SourceType:  assignment.SourceType,
		Target:      assignment.Target,
		TargetType:  assignment.TargetType,
		State:       assignment.State,
		Value:       assignment.Value,
	}
}

func (c *converter) FromInput(in *model.FormationAssignmentInput) *model.FormationAssignment {
	if in == nil {
		return nil
	}

	return &model.FormationAssignment{
		FormationID: in.FormationID,
		Source:      in.Source,
		SourceType:  in.SourceType,
		Target:      in.Target,
		TargetType:  in.TargetType,
		State:       in.State,
		Value:       in.Value,
	}
}

func (c *converter) MultipleFromInput(in []*model.FormationAssignmentInput) []*model.FormationAssignment {
	if in == nil {
		return nil
	}
	formationAssignments := make([]*model.FormationAssignment, 0, len(in))
	for _, fai := range in {
		if fai == nil {
			continue
		}

		formationAssignments = append(formationAssignments, c.FromInput(fai))
	}
	return formationAssignments
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationAssignment) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:          in.ID,
		FormationID: in.FormationID,
		TenantID:    in.TenantID,
		Source:      in.Source,
		SourceType:  in.SourceType,
		Target:      in.Target,
		TargetType:  in.TargetType,
		State:       in.State,
		Value:       repo.NewNullableStringFromJSONRawMessage(in.Value),
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.FormationAssignment {
	if e == nil {
		return nil
	}

	return &model.FormationAssignment{
		ID:          e.ID,
		FormationID: e.FormationID,
		TenantID:    e.TenantID,
		Source:      e.Source,
		SourceType:  e.SourceType,
		Target:      e.Target,
		TargetType:  e.TargetType,
		State:       e.State,
		Value:       repo.JSONRawMessageFromNullableString(e.Value),
	}
}
