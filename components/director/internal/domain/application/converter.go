package application

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"time"
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
		ID:                  in.ID,
		TenantID:            in.Tenant,
		Name:                in.Name,
		Description:         repo.NewNullableString(in.Description),
		StatusCondition:     string(in.Status.Condition),
		StatusTimestamp:     in.Status.Timestamp,
		HealthCheckURL:      repo.NewNullableString(in.HealthCheckURL),
		IntegrationSystemID: repo.NewNullableString(in.IntegrationSystemID),
		ProviderName:        repo.NewNullableString(in.ProviderName),
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
		IntegrationSystemID: repo.StringPtrFromNullableString(entity.IntegrationSystemID),
		HealthCheckURL:      repo.StringPtrFromNullableString(entity.HealthCheckURL),
		ProviderName:        repo.StringPtrFromNullableString(entity.ProviderName),
	}
}

func (c *converter) ToGraphQL(in *model.Application) *graphql.Application {
	if in == nil {
		return nil
	}

	return &graphql.Application{
		ID:                  in.ID,
		Status:              c.statusToGraphQL(in.Status),
		Name:                in.Name,
		Description:         in.Description,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		ProviderName:        in.ProviderName,
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

func (c *converter) CreateInputFromGraphQL(in graphql.ApplicationRegisterInput) model.ApplicationRegisterInput {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = *in.Labels
	}

	return model.ApplicationRegisterInput{
		Name:                in.Name,
		Description:         in.Description,
		Labels:              labels,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		ProviderName:        in.ProviderName,
		Webhooks:            c.webhook.MultipleInputFromGraphQL(in.Webhooks),
		Documents:           c.document.MultipleInputFromGraphQL(in.Documents),
		EventDefinitions:    c.eventAPI.MultipleInputFromGraphQL(in.EventDefinitions),
		APIDefinitions:      c.api.MultipleInputFromGraphQL(in.APIDefinitions),
	}
}

func (c *converter) UpdateInputFromGraphQL(in graphql.ApplicationUpdateInput) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Name:                in.Name,
		Description:         in.Description,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		ProviderName:        in.ProviderName,
	}
}

func (c *converter) CreateInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error) {
	var appInput graphql.ApplicationRegisterInput
	err := json.Unmarshal([]byte(in), &appInput)
	if err != nil {
		return graphql.ApplicationRegisterInput{}, errors.Wrap(err, "while unmarshalling string to ApplicationRegisterInput")
	}

	return appInput, nil
}

func (c *converter) CreateInputGQLToJSON(in *graphql.ApplicationRegisterInput) (string, error) {
	appInput, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling application input")
	}

	return string(appInput), nil
}

func (c *converter) ToModel(obj *graphql.Application, tenantID string) *model.Application {
	if obj == nil {
		return nil
	}

	return &model.Application{
		ID:                  obj.ID,
		ProviderName:        obj.ProviderName,
		Tenant:              tenantID,
		Name:                obj.Name,
		Description:         obj.Description,
		Status:              c.statusToModel(obj.Status),
		HealthCheckURL:      obj.HealthCheckURL,
		IntegrationSystemID: obj.IntegrationSystemID,
	}
}

func (c *converter) statusToGraphQL(in *model.ApplicationStatus) *graphql.ApplicationStatus {
	if in == nil {
		return &graphql.ApplicationStatus{Condition: graphql.ApplicationStatusConditionInitial}
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

func (c *converter) statusToModel(in *graphql.ApplicationStatus) *model.ApplicationStatus {
	if in == nil {
		return &model.ApplicationStatus{Condition: model.ApplicationStatusConditionInitial}
	}

	var condition model.ApplicationStatusCondition

	switch in.Condition {
	case graphql.ApplicationStatusConditionInitial:
		condition = model.ApplicationStatusConditionInitial
	case graphql.ApplicationStatusConditionFailed:
		condition = model.ApplicationStatusConditionFailed
	case graphql.ApplicationStatusConditionReady:
		condition = model.ApplicationStatusConditionReady
	default:
		condition = model.ApplicationStatusConditionInitial
	}
	return &model.ApplicationStatus{
		Condition: condition,
		Timestamp: time.Time(in.Timestamp),
	}
}
