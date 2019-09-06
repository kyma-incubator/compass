package application

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
	webhook  WebhookConverter
	api      APIConverter
	eventAPI EventAPIConverter
	document DocumentConverter
}

func NewConverter(webhook WebhookConverter, api APIConverter, eventAPI EventAPIConverter, document DocumentConverter) *converter {
	return &converter{webhook: webhook, api: api, eventAPI: eventAPI, document: document}
}

func (c *converter) ToEntity(in *model.Application) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	if in.Status == nil {
		return nil, errors.New("invalid input model")
	}

	return &Entity{
		ID:              in.ID,
		TenantID:        in.Tenant,
		Name:            in.Name,
		Description:     repo.NewNullableString(in.Description),
		StatusCondition: string(in.Status.Condition),
		StatusTimestamp: in.Status.Timestamp,
		HealthCheckURL:  repo.NewNullableString(in.HealthCheckURL),
	}, nil
}

func (c *converter) FromEntity(entity *Entity) *model.Application {
	if entity == nil {
		return nil
	}

	return &model.Application{
		ID:          entity.ID,
		Tenant:      entity.TenantID,
		Name:        entity.Name,
		Description: repo.StringPtrFromNullableString(entity.Description),
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusCondition(entity.StatusCondition),
			Timestamp: entity.StatusTimestamp,
		},
		HealthCheckURL: repo.StringPtrFromNullableString(entity.HealthCheckURL),
	}
}

func (c *converter) ToGraphQL(in *model.Application) *graphql.Application {
	if in == nil {
		return nil
	}

	return &graphql.Application{
		ID:             in.ID,
		Status:         c.statusToGraphQL(in.Status),
		Name:           in.Name,
		Description:    in.Description,
		HealthCheckURL: in.HealthCheckURL,
	}
}

func (c *converter) MultipleToGraphQL(in []*model.Application) []*graphql.Application {
	var runtimes []*graphql.Application
	for _, r := range in {
		if r == nil {
			continue
		}

		runtimes = append(runtimes, c.ToGraphQL(r))
	}

	return runtimes
}

func (c *converter) InputFromGraphQL(in graphql.ApplicationInput) model.ApplicationInput {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = *in.Labels
	}

	return model.ApplicationInput{
		Name:           in.Name,
		Description:    in.Description,
		Labels:         labels,
		HealthCheckURL: in.HealthCheckURL,
		Webhooks:       c.webhook.MultipleInputFromGraphQL(in.Webhooks),
		Documents:      c.document.MultipleInputFromGraphQL(in.Documents),
		EventAPIs:      c.eventAPI.MultipleInputFromGraphQL(in.EventAPIs),
		Apis:           c.api.MultipleInputFromGraphQL(in.Apis),
	}
}

func (c *converter) statusToGraphQL(in *model.ApplicationStatus) *graphql.ApplicationStatus {
	if in == nil {
		return &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
		}
	}

	var condition graphql.ApplicationStatusCondition

	switch in.Condition {
	case model.ApplicationStatusConditionInitial:
		condition = graphql.ApplicationStatusConditionInitial
	case model.ApplicationStatusConditionFailed:
		condition = graphql.ApplicationStatusConditionFailed
	case model.ApplicationStatusConditionReady:
		condition = graphql.ApplicationStatusConditionReady
	default:
		condition = graphql.ApplicationStatusConditionInitial
	}

	return &graphql.ApplicationStatus{
		Condition: condition,
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}
