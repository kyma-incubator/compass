package aspect

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// AspectEventResourceConverter converts Aspect Event Resources between the model.AspectEventResource service-layer representation and the graphql-layer representation graphql.AspectEventDefinition.
//
//go:generate mockery --name=AspectEventResourceConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectEventResourceConverter interface {
	MultipleToGraphQL(in []*model.AspectEventResource) ([]*graphql.AspectEventDefinition, error)
	MultipleInputFromGraphQL(in []*graphql.AspectEventDefinitionInput) ([]*model.AspectEventResourceInput, error)
}

type converter struct {
	aspectEventResourceConverter AspectEventResourceConverter
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass Aspect.
func NewConverter(aspectEventResourceConverter AspectEventResourceConverter) *converter {
	return &converter{
		aspectEventResourceConverter: aspectEventResourceConverter,
	}
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
func (c *converter) ToGraphQL(in *model.Aspect, aspectEventResources []*model.AspectEventResource) (*graphql.Aspect, error) {
	if in == nil {
		return nil, nil
	}

	var apiResources []*graphql.AspectAPIDefinition
	if in.APIResources != nil {
		if err := json.Unmarshal(in.APIResources, &apiResources); err != nil {
			return nil, err
		}
	}

	eventResources, err := c.aspectEventResourceConverter.MultipleToGraphQL(aspectEventResources)
	if err != nil {
		return nil, err
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
			CreatedAt: graphql.TimePtrToGraphqlTimestampPtr(in.CreatedAt),
			UpdatedAt: graphql.TimePtrToGraphqlTimestampPtr(in.UpdatedAt),
			DeletedAt: graphql.TimePtrToGraphqlTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
	}, nil
}

// MultipleToGraphQL converts the provided service-layer representations of an Aspect to the graphql-layer ones.
func (c *converter) MultipleToGraphQL(in []*model.Aspect, aspectEventResourcesByAspectID map[string][]*model.AspectEventResource) ([]*graphql.Aspect, error) {
	aspects := make([]*graphql.Aspect, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}

		aspect, err := c.ToGraphQL(a, aspectEventResourcesByAspectID[a.ID])
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

	eventResources, err := c.aspectEventResourceConverter.MultipleInputFromGraphQL(in.EventResources)
	if err != nil {
		return nil, err
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

func getMandatory(inputMandatory *bool) *bool {
	m := false
	if inputMandatory == nil {
		inputMandatory = &m
	}
	return inputMandatory
}
