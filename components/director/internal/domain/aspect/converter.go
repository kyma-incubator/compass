package aspect

import (
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass Aspect.
func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an Aspect to the service-layer representation model.Aspect.
func (c *converter) FromEntity(entity *Entity) *model.Aspect {
	if entity == nil {
		return nil
	}

	return &model.Aspect{
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		IntegrationDependencyID:      entity.IntegrationDependencyID,
		Title:                        entity.Title,
		Description:                  repo.StringPtrFromNullableString(entity.Description),
		Mandatory:                    repo.BoolPtrFromNullableBool(entity.Mandatory),
		SupportMultipleProviders:     repo.BoolPtrFromNullableBool(entity.SupportMultipleProviders),
		APIResources:                 repo.JSONRawMessageFromNullableString(entity.APIResources),
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

// ToEntity converts the provided service-layer representation of an Aspect to the repository-layer one.
func (c *converter) ToEntity(aspectModel *model.Aspect) *Entity {
	if aspectModel == nil {
		return nil
	}

	return &Entity{
		ApplicationID:                repo.NewNullableString(aspectModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(aspectModel.ApplicationTemplateVersionID),
		IntegrationDependencyID:      aspectModel.IntegrationDependencyID,
		Title:                        aspectModel.Title,
		Description:                  repo.NewNullableString(aspectModel.Description),
		Mandatory:                    repo.NewNullableBool(aspectModel.Mandatory),
		SupportMultipleProviders:     repo.NewNullableBool(aspectModel.SupportMultipleProviders),
		APIResources:                 repo.NewNullableStringFromJSONRawMessage(aspectModel.APIResources),
		BaseEntity: &repo.BaseEntity{
			ID:        aspectModel.ID,
			Ready:     aspectModel.Ready,
			CreatedAt: aspectModel.CreatedAt,
			UpdatedAt: aspectModel.UpdatedAt,
			DeletedAt: aspectModel.DeletedAt,
			Error:     repo.NewNullableString(aspectModel.Error),
		},
	}
}

// ToGraphQL converts the provided service-layer representation of an Aspect to the graphql-layer one.
func (c *converter) ToGraphQL(in *model.Aspect) (*graphql.Aspect, error) {
	if in == nil {
		return nil, nil
	}

	var apiResources []*graphql.AspectAPIDefinition
	var eventResources []*graphql.AspectEventDefinition

	if in.APIResources != nil {
		if err := json.Unmarshal(in.APIResources, &apiResources); err != nil {
			return nil, err
		}
	}

	if in.EventResources != nil {
		if err := json.Unmarshal(in.EventResources, &eventResources); err != nil {
			return nil, err
		}
	}

	return &graphql.Aspect{
		Name:           in.Title,
		Description:    in.Description,
		Mandatory:      in.Mandatory,
		APIResources:   apiResources,
		EventResources: eventResources,
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: timePtrToTimestampPtr(in.CreatedAt),
			UpdatedAt: timePtrToTimestampPtr(in.UpdatedAt),
			DeletedAt: timePtrToTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
	}, nil
}

// MultipleToGraphQL converts the provided service-layer representations of an Aspect to the graphql-layer ones.
func (c *converter) MultipleToGraphQL(in []*model.Aspect) ([]*graphql.Aspect, error) {
	aspects := make([]*graphql.Aspect, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}

		aspect, err := c.ToGraphQL(a)
		if err != nil {
			return nil, err
		}

		aspects = append(aspects, aspect)
	}

	return aspects, nil
}

// InputFromGraphQL converts the provided graphql-layer representation of an Aspect to the service-layer one.
func (c *converter) InputFromGraphQL(in *graphql.AspectInput) (*model.AspectInput, error) {
	if in == nil {
		return nil, nil
	}

	apiResources, err := json.Marshal(in.APIResources)
	if err != nil {
		return nil, errors.Wrap(err, "error while marshalling aspect api resources")
	}

	eventResources, err := json.Marshal(in.EventResources)
	if err != nil {
		return nil, errors.Wrap(err, "error while marshalling aspect event resources")
	}

	return &model.AspectInput{
		Title:          in.Name,
		Description:    in.Description,
		Mandatory:      getMandatory(in.Mandatory),
		APIResources:   apiResources,
		EventResources: eventResources,
	}, nil
}

// MultipleInputFromGraphQL converts the provided graphql-layer representations of an Aspect to the service-layer ones.
func (c *converter) MultipleInputFromGraphQL(in []*graphql.AspectInput) ([]*model.AspectInput, error) {
	inputs := make([]*model.AspectInput, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}
		aspectIn, err := c.InputFromGraphQL(a)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, aspectIn)
	}

	return inputs, nil
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}

func getMandatory(inputMandatory *bool) *bool {
	m := false
	if inputMandatory == nil {
		inputMandatory = &m
	}
	return inputMandatory
}
