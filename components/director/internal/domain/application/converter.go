package application

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//
////go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
//type APIConverter interface {
//	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
//}

//go:generate mockery -name=EventAPIConverter -output=automock -outpkg=automock -case=underscore
type EventAPIConverter interface {
	MultipleInputFromGraphQL(in []*graphql.EventAPIDefinitionInput) []*model.EventAPIDefinitionInput
}

type converter struct {
	webhook  WebhookConverter
	api      APIConverter
	eventAPI EventAPIConverter
	document DocumentConverter
}

func NewConverter(webhook WebhookConverter, api APIConverter, eventAPI EventAPIConverter, document DocumentConverter) *converter {
	return &converter{webhook: webhook, api: api, eventAPI: eventAPI, document: document}
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
		Tenant:         graphql.Tenant(in.Tenant),
		Annotations:    in.Annotations,
		Labels:         in.Labels,
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
	var annotations map[string]interface{}
	if in.Annotations != nil {
		annotations = *in.Annotations
	}

	var labels map[string][]string
	if in.Labels != nil {
		labels = *in.Labels
	}

	return model.ApplicationInput{
		Name:           in.Name,
		Description:    in.Description,
		Annotations:    annotations,
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
