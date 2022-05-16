package runtime

import (
	"database/sql"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct {
	webhook WebhookConverter
}

// NewConverter missing godoc
func NewConverter(webhook WebhookConverter) *converter {
	return &converter{
		webhook: webhook,
	}
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *model.Runtime) *graphql.Runtime {
	if in == nil {
		return nil
	}

	return &graphql.Runtime{
		ID:          in.ID,
		Status:      c.statusToGraphQL(in.Status),
		Name:        in.Name,
		Description: in.Description,
		Metadata:    c.metadataToGraphQL(in.CreationTimestamp),
	}
}

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime {
	runtimes := make([]*graphql.Runtime, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		runtimes = append(runtimes, c.ToGraphQL(r))
	}

	return runtimes
}

// RegisterInputFromGraphQL missing godoc
func (c *converter) RegisterInputFromGraphQL(in graphql.RuntimeRegisterInput) (model.RuntimeRegisterInput, error) {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = in.Labels
	}

	webhooks, err := c.webhook.MultipleInputFromGraphQL(in.Webhooks)
	if err != nil {
		return model.RuntimeRegisterInput{}, err
	}

	return model.RuntimeRegisterInput{
		Name:            in.Name,
		Description:     in.Description,
		Labels:          labels,
		Webhooks:        webhooks,
		StatusCondition: c.statusConditionToModel(in.StatusCondition),
	}, nil
}

// UpdateInputFromGraphQL missing godoc
func (c *converter) UpdateInputFromGraphQL(in graphql.RuntimeUpdateInput) model.RuntimeUpdateInput {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = in.Labels
	}

	return model.RuntimeUpdateInput{
		Name:            in.Name,
		Description:     in.Description,
		Labels:          labels,
		StatusCondition: c.statusConditionToModel(in.StatusCondition),
	}
}

func (c *converter) statusToGraphQL(in *model.RuntimeStatus) *graphql.RuntimeStatus {
	if in == nil {
		return &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
		}
	}

	var condition graphql.RuntimeStatusCondition

	switch in.Condition {
	case model.RuntimeStatusConditionInitial:
		condition = graphql.RuntimeStatusConditionInitial
	case model.RuntimeStatusConditionProvisioning:
		condition = graphql.RuntimeStatusConditionProvisioning
	case model.RuntimeStatusConditionFailed:
		condition = graphql.RuntimeStatusConditionFailed
	case model.RuntimeStatusConditionConnected:
		condition = graphql.RuntimeStatusConditionConnected
	default:
		condition = graphql.RuntimeStatusConditionInitial
	}

	return &graphql.RuntimeStatus{
		Condition: condition,
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}

func (c *converter) metadataToGraphQL(creationTimestamp time.Time) *graphql.RuntimeMetadata {
	return &graphql.RuntimeMetadata{
		CreationTimestamp: graphql.Timestamp(creationTimestamp),
	}
}

func (c *converter) statusConditionToModel(in *graphql.RuntimeStatusCondition) *model.RuntimeStatusCondition {
	if in == nil {
		return nil
	}

	var condition model.RuntimeStatusCondition
	switch *in {
	case graphql.RuntimeStatusConditionConnected:
		condition = model.RuntimeStatusConditionConnected
	case graphql.RuntimeStatusConditionFailed:
		condition = model.RuntimeStatusConditionFailed
	case graphql.RuntimeStatusConditionProvisioning:
		condition = model.RuntimeStatusConditionProvisioning
	case graphql.RuntimeStatusConditionInitial:
		fallthrough
	default:
		condition = model.RuntimeStatusConditionInitial
	}

	return &condition
}

func (*converter) ToEntity(model *model.Runtime) (*Runtime, error) {
	if model == nil {
		return nil, nil
	}
	var nullDescription sql.NullString
	if model.Description != nil && len(*model.Description) > 0 {
		nullDescription = sql.NullString{
			String: *model.Description,
			Valid:  true,
		}
	}
	if model.Status == nil {
		return nil, apperrors.NewInternalError("invalid input model")
	}

	return &Runtime{
		ID:                model.ID,
		Name:              model.Name,
		Description:       nullDescription,
		StatusCondition:   string(model.Status.Condition),
		StatusTimestamp:   model.Status.Timestamp,
		CreationTimestamp: model.CreationTimestamp,
	}, nil
}

func (*converter) FromEntity(e *Runtime) *model.Runtime {
	if e == nil {
		return nil
	}

	var description *string
	if e.Description.Valid {
		description = new(string)
		*description = e.Description.String
	}

	return &model.Runtime{
		ID:          e.ID,
		Name:        e.Name,
		Description: description,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusCondition(e.StatusCondition),
			Timestamp: e.StatusTimestamp,
		},
		CreationTimestamp: e.CreationTimestamp,
	}
}
