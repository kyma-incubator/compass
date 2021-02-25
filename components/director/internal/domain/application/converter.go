package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
	webhook WebhookConverter

	bndl BundleConverter
}

func NewConverter(webhook WebhookConverter, bndlConverter BundleConverter) *converter {
	return &converter{webhook: webhook, bndl: bndlConverter}
}

func (c *converter) ToEntity(in *model.Application) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	if in.Status == nil {
		return nil, apperrors.NewInternalError("invalid input model")
	}

	return &Entity{
		TenantID:            in.Tenant,
		Name:                in.Name,
		ProviderName:        repo.NewNullableString(in.ProviderName),
		Description:         repo.NewNullableString(in.Description),
		StatusCondition:     string(in.Status.Condition),
		StatusTimestamp:     in.Status.Timestamp,
		HealthCheckURL:      repo.NewNullableString(in.HealthCheckURL),
		IntegrationSystemID: repo.NewNullableString(in.IntegrationSystemID),
		BaseURL:             repo.NewNullableString(in.BaseURL),
		Labels:              repo.NewNullableStringFromJSONRawMessage(in.Labels),
		BaseEntity: &repo.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: in.CreatedAt,
			UpdatedAt: in.UpdatedAt,
			DeletedAt: in.DeletedAt,
			Error:     repo.NewNullableString(in.Error),
		},
	}, nil
}

func (c *converter) FromEntity(entity *Entity) *model.Application {
	if entity == nil {
		return nil
	}

	return &model.Application{
		ProviderName: repo.StringPtrFromNullableString(entity.ProviderName),
		Tenant:       entity.TenantID,
		Name:         entity.Name,
		Description:  repo.StringPtrFromNullableString(entity.Description),
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusCondition(entity.StatusCondition),
			Timestamp: entity.StatusTimestamp,
		},
		HealthCheckURL:      repo.StringPtrFromNullableString(entity.HealthCheckURL),
		IntegrationSystemID: repo.StringPtrFromNullableString(entity.IntegrationSystemID),
		BaseURL:             repo.StringPtrFromNullableString(entity.BaseURL),
		Labels:              repo.JSONRawMessageFromNullableString(entity.Labels),
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
	}
}

func (c *converter) ToGraphQL(in *model.Application) *graphql.Application {
	if in == nil {
		return nil
	}

	return &graphql.Application{
		Status:              c.statusToGraphQL(in.Status),
		Name:                in.Name,
		Description:         in.Description,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		ProviderName:        in.ProviderName,
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: timePtrToTimestampPtr(in.CreatedAt),
			UpdatedAt: timePtrToTimestampPtr(in.UpdatedAt),
			DeletedAt: timePtrToTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
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

func (c *converter) CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error) {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = *in.Labels
	}

	log.C(ctx).Debugf("Converting Webhooks from Application registration GraphQL input with name %s", in.Name)
	webhooks, err := c.webhook.MultipleInputFromGraphQL(in.Webhooks)
	if err != nil {
		return model.ApplicationRegisterInput{}, errors.Wrap(err, "while converting Webhooks")
	}

	log.C(ctx).Debugf("Converting Bundles from Application registration GraphQL input with name %s", in.Name)
	bundles, err := c.bndl.MultipleCreateInputFromGraphQL(in.Bundles)
	if err != nil {
		return model.ApplicationRegisterInput{}, errors.Wrap(err, "while converting Bundles")
	}

	return model.ApplicationRegisterInput{
		Name:                in.Name,
		Description:         in.Description,
		Labels:              labels,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		StatusCondition:     c.statusConditionToModel(in.StatusCondition),
		ProviderName:        in.ProviderName,
		Webhooks:            webhooks,
		Bundles:             bundles,
	}, nil
}

func (c *converter) UpdateInputFromGraphQL(in graphql.ApplicationUpdateInput) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Description:         in.Description,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		ProviderName:        in.ProviderName,
		StatusCondition:     c.statusConditionToModel(in.StatusCondition),
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

func (c *converter) GraphQLToModel(obj *graphql.Application, tenantID string) *model.Application {
	if obj == nil {
		return nil
	}

	return &model.Application{
		ProviderName:        obj.ProviderName,
		Tenant:              tenantID,
		Name:                obj.Name,
		Description:         obj.Description,
		Status:              c.statusToModel(obj.Status),
		HealthCheckURL:      obj.HealthCheckURL,
		IntegrationSystemID: obj.IntegrationSystemID,
		BaseEntity: &model.BaseEntity{
			ID: obj.ID,
		},
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
	case model.ApplicationStatusConditionConnected:
		condition = graphql.ApplicationStatusConditionConnected
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
	case graphql.ApplicationStatusConditionConnected:
		condition = model.ApplicationStatusConditionConnected
	default:
		condition = model.ApplicationStatusConditionInitial
	}
	return &model.ApplicationStatus{
		Condition: condition,
		Timestamp: time.Time(in.Timestamp),
	}
}

func (c *converter) statusConditionToModel(in *graphql.ApplicationStatusCondition) *model.ApplicationStatusCondition {
	if in == nil {
		return nil
	}

	var condition model.ApplicationStatusCondition
	switch *in {
	case graphql.ApplicationStatusConditionConnected:
		condition = model.ApplicationStatusConditionConnected
	case graphql.ApplicationStatusConditionFailed:
		condition = model.ApplicationStatusConditionFailed
	case graphql.ApplicationStatusConditionInitial:
		fallthrough
	default:
		condition = model.ApplicationStatusConditionInitial
	}

	return &condition
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}
