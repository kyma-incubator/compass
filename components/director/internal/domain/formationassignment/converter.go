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

	var configurationStrValue *string
	if in.Value != nil {
		marshalledValue, err := json.Marshal(in.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal formation assignment configuration while converting formation assignment to GraphQL")
		}
		configurationStrValue = str.Ptr(string(marshalledValue))
	}

	var errorStrValue *string
	if in.Error != nil {
		marshalledValue, err := json.Marshal(in.Error)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal formation assignment error while converting formation assignment to GraphQL")
		}
		errorStrValue = str.Ptr(string(marshalledValue))
	}

	var value *string
	if errorStrValue != nil {
		value = errorStrValue
	} else {
		value = configurationStrValue
	}

	return &graphql.FormationAssignment{
		ID:                            in.ID,
		Source:                        in.Source,
		SourceType:                    graphql.FormationAssignmentType(in.SourceType),
		Target:                        in.Target,
		TargetType:                    graphql.FormationAssignmentType(in.TargetType),
		State:                         in.State,
		Value:                         value,
		Configuration:                 configurationStrValue,
		Error:                         errorStrValue,
		LastStateChangeTimestamp:      graphql.TimePtrToGraphqlTimestampPtr(in.LastStateChangeTimestamp),
		LastNotificationSentTimestamp: graphql.TimePtrToGraphqlTimestampPtr(in.LastNotificationSentTimestamp),
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

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationAssignment) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:                            in.ID,
		FormationID:                   in.FormationID,
		TenantID:                      in.TenantID,
		Source:                        in.Source,
		SourceType:                    string(in.SourceType),
		Target:                        in.Target,
		TargetType:                    string(in.TargetType),
		State:                         in.State,
		Value:                         repo.NewNullableStringFromJSONRawMessage(in.Value),
		Error:                         repo.NewNullableStringFromJSONRawMessage(in.Error),
		LastStateChangeTimestamp:      in.LastStateChangeTimestamp,
		LastNotificationSentTimestamp: in.LastNotificationSentTimestamp,
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.FormationAssignment {
	if e == nil {
		return nil
	}

	return &model.FormationAssignment{
		ID:                            e.ID,
		FormationID:                   e.FormationID,
		TenantID:                      e.TenantID,
		Source:                        e.Source,
		SourceType:                    model.FormationAssignmentType(e.SourceType),
		Target:                        e.Target,
		TargetType:                    model.FormationAssignmentType(e.TargetType),
		State:                         e.State,
		Value:                         repo.JSONRawMessageFromNullableString(e.Value),
		Error:                         repo.JSONRawMessageFromNullableString(e.Error),
		LastStateChangeTimestamp:      e.LastStateChangeTimestamp,
		LastNotificationSentTimestamp: e.LastNotificationSentTimestamp,
	}
}
