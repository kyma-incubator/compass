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

// NewConverter missing godoc
func NewConverter(webhook WebhookConverter, bndlConverter BundleConverter) *converter {
	return &converter{webhook: webhook, bndl: bndlConverter}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Application) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	if in.Status == nil {
		return nil, apperrors.NewInternalError("invalid input model")
	}

	return &Entity{
		Name:                  in.Name,
		ProviderName:          repo.NewNullableString(in.ProviderName),
		Description:           repo.NewNullableString(in.Description),
		StatusCondition:       string(in.Status.Condition),
		StatusTimestamp:       in.Status.Timestamp,
		HealthCheckURL:        repo.NewNullableString(in.HealthCheckURL),
		IntegrationSystemID:   repo.NewNullableString(in.IntegrationSystemID),
		ApplicationTemplateID: repo.NewNullableString(in.ApplicationTemplateID),
		BaseURL:               repo.NewNullableString(in.BaseURL),
		SystemNumber:          repo.NewNullableString(in.SystemNumber),
		OrdLabels:             repo.NewNullableStringFromJSONRawMessage(in.OrdLabels),
		CorrelationIDs:        repo.NewNullableStringFromJSONRawMessage(in.CorrelationIDs),
		SystemStatus:          repo.NewNullableString(in.SystemStatus),
		DocumentationLabels:   repo.NewNullableStringFromJSONRawMessage(in.DocumentationLabels),
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

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) *model.Application {
	if entity == nil {
		return nil
	}

	return &model.Application{
		ProviderName: repo.StringPtrFromNullableString(entity.ProviderName),
		Name:         entity.Name,
		SystemNumber: repo.StringPtrFromNullableString(entity.SystemNumber),
		Description:  repo.StringPtrFromNullableString(entity.Description),
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusCondition(entity.StatusCondition),
			Timestamp: entity.StatusTimestamp,
		},
		HealthCheckURL:        repo.StringPtrFromNullableString(entity.HealthCheckURL),
		IntegrationSystemID:   repo.StringPtrFromNullableString(entity.IntegrationSystemID),
		ApplicationTemplateID: repo.StringPtrFromNullableString(entity.ApplicationTemplateID),
		BaseURL:               repo.StringPtrFromNullableString(entity.BaseURL),
		OrdLabels:             repo.JSONRawMessageFromNullableString(entity.OrdLabels),
		CorrelationIDs:        repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		SystemStatus:          repo.StringPtrFromNullableString(entity.SystemStatus),
		DocumentationLabels:   repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
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

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *model.Application) *graphql.Application {
	if in == nil {
		return nil
	}

	return &graphql.Application{
		Status:                c.statusModelToGraphQL(in.Status),
		Name:                  in.Name,
		Description:           in.Description,
		HealthCheckURL:        in.HealthCheckURL,
		BaseURL:               in.BaseURL,
		IntegrationSystemID:   in.IntegrationSystemID,
		ApplicationTemplateID: in.ApplicationTemplateID,
		ProviderName:          in.ProviderName,
		SystemNumber:          in.SystemNumber,
		SystemStatus:          in.SystemStatus,
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

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.Application) []*graphql.Application {
	applications := make([]*graphql.Application, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		applications = append(applications, c.ToGraphQL(r))
	}

	return applications
}

// CreateInputFromGraphQL missing godoc
func (c *converter) CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error) {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = in.Labels
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

	var statusCondition *model.ApplicationStatusCondition
	if in.StatusCondition != nil {
		condition := model.ApplicationStatusCondition(*in.StatusCondition)
		statusCondition = &condition
	}
	return model.ApplicationRegisterInput{
		Name:                in.Name,
		Description:         in.Description,
		Labels:              labels,
		BaseURL:             in.BaseURL,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		StatusCondition:     statusCondition,
		ProviderName:        in.ProviderName,
		Webhooks:            webhooks,
		Bundles:             bundles,
	}, nil
}

// UpdateInputFromGraphQL missing godoc
func (c *converter) UpdateInputFromGraphQL(in graphql.ApplicationUpdateInput) model.ApplicationUpdateInput {
	var statusCondition *model.ApplicationStatusCondition
	if in.StatusCondition != nil {
		condition := model.ApplicationStatusCondition(*in.StatusCondition)
		statusCondition = &condition
	}
	return model.ApplicationUpdateInput{
		Description:         in.Description,
		HealthCheckURL:      in.HealthCheckURL,
		IntegrationSystemID: in.IntegrationSystemID,
		ProviderName:        in.ProviderName,
		StatusCondition:     statusCondition,
		BaseURL:             in.BaseURL,
	}
}

// CreateInputJSONToGQL missing godoc
func (c *converter) CreateInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error) {
	var appInput graphql.ApplicationRegisterInput
	err := json.Unmarshal([]byte(in), &appInput)
	if err != nil {
		return graphql.ApplicationRegisterInput{}, errors.Wrap(err, "while unmarshalling string to ApplicationRegisterInput")
	}

	return appInput, nil
}

// CreateInputJSONToModel converts a JSON input to an application model.
func (c *converter) CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error) {
	modelIn := model.ApplicationRegisterInput{}
	if err := json.Unmarshal([]byte(in), &modelIn); err != nil {
		return modelIn, errors.Wrap(err, "while unmarshalling application input JSON")
	}
	return modelIn, nil
}

// CreateInputGQLToJSON missing godoc
func (c *converter) CreateInputGQLToJSON(in *graphql.ApplicationRegisterInput) (string, error) {
	appInput, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling application input")
	}

	return string(appInput), nil
}

// GraphQLToModel missing godoc
func (c *converter) GraphQLToModel(obj *graphql.Application, tenantID string) *model.Application {
	if obj == nil {
		return nil
	}

	return &model.Application{
		ProviderName:        obj.ProviderName,
		Name:                obj.Name,
		Description:         obj.Description,
		Status:              c.statusGraphQLToModel(obj.Status),
		HealthCheckURL:      obj.HealthCheckURL,
		IntegrationSystemID: obj.IntegrationSystemID,
		SystemNumber:        obj.SystemNumber,
		BaseEntity: &model.BaseEntity{
			ID: obj.ID,
		},
	}
}

func (c *converter) statusModelToGraphQL(in *model.ApplicationStatus) *graphql.ApplicationStatus {
	if in == nil {
		return &graphql.ApplicationStatus{Condition: graphql.ApplicationStatusConditionInitial}
	}

	return &graphql.ApplicationStatus{
		Condition: graphql.ApplicationStatusCondition(in.Condition),
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}

func (c *converter) statusGraphQLToModel(in *graphql.ApplicationStatus) *model.ApplicationStatus {
	if in == nil {
		return &model.ApplicationStatus{Condition: model.ApplicationStatusConditionInitial}
	}

	return &model.ApplicationStatus{
		Condition: model.ApplicationStatusCondition(in.Condition),
		Timestamp: time.Time(in.Timestamp),
	}
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}
