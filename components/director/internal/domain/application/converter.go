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
		ApplicationNamespace:  repo.NewNullableString(in.ApplicationNamespace),
		SystemNumber:          repo.NewNullableString(in.SystemNumber),
		LocalTenantID:         repo.NewNullableString(in.LocalTenantID),
		OrdLabels:             repo.NewNullableStringFromJSONRawMessage(in.OrdLabels),
		CorrelationIDs:        repo.NewNullableStringFromJSONRawMessage(in.CorrelationIDs),
		SystemStatus:          repo.NewNullableString(in.SystemStatus),
		Tags:                  repo.NewNullableStringFromJSONRawMessage(in.Tags),
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
		ProviderName:  repo.StringPtrFromNullableString(entity.ProviderName),
		Name:          entity.Name,
		SystemNumber:  repo.StringPtrFromNullableString(entity.SystemNumber),
		LocalTenantID: repo.StringPtrFromNullableString(entity.LocalTenantID),
		Description:   repo.StringPtrFromNullableString(entity.Description),
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusCondition(entity.StatusCondition),
			Timestamp: entity.StatusTimestamp,
		},
		HealthCheckURL:        repo.StringPtrFromNullableString(entity.HealthCheckURL),
		IntegrationSystemID:   repo.StringPtrFromNullableString(entity.IntegrationSystemID),
		ApplicationTemplateID: repo.StringPtrFromNullableString(entity.ApplicationTemplateID),
		BaseURL:               repo.StringPtrFromNullableString(entity.BaseURL),
		ApplicationNamespace:  repo.StringPtrFromNullableString(entity.ApplicationNamespace),
		OrdLabels:             repo.JSONRawMessageFromNullableString(entity.OrdLabels),
		CorrelationIDs:        repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		SystemStatus:          repo.StringPtrFromNullableString(entity.SystemStatus),
		Tags:                  repo.JSONRawMessageFromNullableString(entity.Tags),
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
		ApplicationNamespace:  in.ApplicationNamespace,
		IntegrationSystemID:   in.IntegrationSystemID,
		ApplicationTemplateID: in.ApplicationTemplateID,
		ProviderName:          in.ProviderName,
		SystemNumber:          in.SystemNumber,
		LocalTenantID:         in.LocalTenantID,
		SystemStatus:          in.SystemStatus,
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: graphql.TimePtrToGraphqlTimestampPtr(in.CreatedAt),
			UpdatedAt: graphql.TimePtrToGraphqlTimestampPtr(in.UpdatedAt),
			DeletedAt: graphql.TimePtrToGraphqlTimestampPtr(in.DeletedAt),
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
		Name:                 in.Name,
		Description:          in.Description,
		Labels:               labels,
		LocalTenantID:        in.LocalTenantID,
		BaseURL:              in.BaseURL,
		ApplicationNamespace: in.ApplicationNamespace,
		HealthCheckURL:       in.HealthCheckURL,
		IntegrationSystemID:  in.IntegrationSystemID,
		StatusCondition:      statusCondition,
		ProviderName:         in.ProviderName,
		Webhooks:             webhooks,
		Bundles:              bundles,
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
		Description:          in.Description,
		HealthCheckURL:       in.HealthCheckURL,
		IntegrationSystemID:  in.IntegrationSystemID,
		LocalTenantID:        in.LocalTenantID,
		ProviderName:         in.ProviderName,
		StatusCondition:      statusCondition,
		BaseURL:              in.BaseURL,
		ApplicationNamespace: in.ApplicationNamespace,
	}
}

// CreateRegisterInputJSONToGQL missing godoc
func (c *converter) CreateRegisterInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error) {
	var appInput graphql.ApplicationRegisterInput
	err := json.Unmarshal([]byte(in), &appInput)
	if err != nil {
		return graphql.ApplicationRegisterInput{}, errors.Wrap(err, "while unmarshalling string to ApplicationRegisterInput")
	}

	return appInput, nil
}

// CreateJSONInputJSONToGQL missing godoc
func (c *converter) CreateJSONInputJSONToGQL(in string) (graphql.ApplicationJSONInput, error) {
	var appInput graphql.ApplicationJSONInput
	err := json.Unmarshal([]byte(in), &appInput)
	if err != nil {
		return graphql.ApplicationJSONInput{}, errors.Wrap(err, "while unmarshalling string to ApplicationJSONInput")
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

// CreateRegisterInputGQLToJSON missing godoc
func (c *converter) CreateRegisterInputGQLToJSON(in *graphql.ApplicationRegisterInput) (string, error) {
	appInput, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling application input")
	}

	return string(appInput), nil
}

// CreateJSONInputGQLToJSON missing godoc
func (c *converter) CreateJSONInputGQLToJSON(in *graphql.ApplicationJSONInput) (string, error) {
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
		ProviderName:         obj.ProviderName,
		Name:                 obj.Name,
		Description:          obj.Description,
		LocalTenantID:        obj.LocalTenantID,
		Status:               c.statusGraphQLToModel(obj.Status),
		HealthCheckURL:       obj.HealthCheckURL,
		IntegrationSystemID:  obj.IntegrationSystemID,
		SystemNumber:         obj.SystemNumber,
		ApplicationNamespace: obj.ApplicationNamespace,
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

type appWithTenantsConverter struct {
	appConv    ApplicationConverter
	tenantConv TenantConverter
}

// NewAppWithTenantsConverter creates new application with tenant converter
func NewAppWithTenantsConverter(appConv ApplicationConverter, tenantConv TenantConverter) *appWithTenantsConverter {
	return &appWithTenantsConverter{appConv: appConv, tenantConv: tenantConv}
}

// MultipleToGraphQL converts multiple model objects to graphql objects
func (c *appWithTenantsConverter) MultipleToGraphQL(in []*model.ApplicationWithTenants) []*graphql.ApplicationWithTenants {
	applicationWithTenants := make([]*graphql.ApplicationWithTenants, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		appWithTenants := &graphql.ApplicationWithTenants{
			Application: c.appConv.ToGraphQL(&r.Application),
			Tenants:     c.tenantConv.MultipleToGraphQL(r.Tenants),
		}
		applicationWithTenants = append(applicationWithTenants, appWithTenants)
	}

	return applicationWithTenants
}
