package formationassignment

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// NewConverter creates a new instance of gqlConverter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in *model.FormationAssignment) *graphql.FormationAssignment {
	if in == nil {
		return nil
	}
	marshalledValue, err := json.Marshal(in.Value)
	if err != nil {
		return nil
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
	}
}

// MultipleToGraphQL converts multiple internal models to GraphQL models
func (c *converter) MultipleToGraphQL(in []*model.FormationAssignment) []*graphql.FormationAssignment {
	if in == nil {
		return nil
	}
	formationTemplates := make([]*graphql.FormationAssignment, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		formationTemplates = append(formationTemplates, c.ToGraphQL(r))
	}

	return formationTemplates
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
