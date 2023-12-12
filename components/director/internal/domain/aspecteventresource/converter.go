package aspecteventresource

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

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass Aspect Event Resource.
func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an Aspect Event Resource to the service-layer representation model.AspectEventResource.
func (c *converter) FromEntity(entity *Entity) *model.AspectEventResource {
	if entity == nil {
		return nil
	}

	return &model.AspectEventResource{
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		AspectID:                     entity.AspectID,
		OrdID:                        entity.OrdID,
		MinVersion:                   repo.StringPtrFromNullableString(entity.MinVersion),
		Subset:                       repo.JSONRawMessageFromNullableString(entity.Subset),
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

// ToEntity converts the provided service-layer representation of an Aspect Event Resource to the repository-layer one.
func (c *converter) ToEntity(aspectEventResourceModel *model.AspectEventResource) *Entity {
	if aspectEventResourceModel == nil {
		return nil
	}

	return &Entity{
		ApplicationID:                repo.NewNullableString(aspectEventResourceModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(aspectEventResourceModel.ApplicationTemplateVersionID),
		AspectID:                     aspectEventResourceModel.AspectID,
		OrdID:                        aspectEventResourceModel.OrdID,
		MinVersion:                   repo.NewNullableString(aspectEventResourceModel.MinVersion),
		Subset:                       repo.NewNullableStringFromJSONRawMessage(aspectEventResourceModel.Subset),
		BaseEntity: &repo.BaseEntity{
			ID:        aspectEventResourceModel.ID,
			Ready:     aspectEventResourceModel.Ready,
			CreatedAt: aspectEventResourceModel.CreatedAt,
			UpdatedAt: aspectEventResourceModel.UpdatedAt,
			DeletedAt: aspectEventResourceModel.DeletedAt,
			Error:     repo.NewNullableString(aspectEventResourceModel.Error),
		},
	}
}

// InputFromGraphQL converts the provided graphql-layer representation of an Aspect Event Definition to the service-layer one.
func (c *converter) InputFromGraphQL(in *graphql.AspectEventDefinitionInput) (*model.AspectEventResourceInput, error) {
	if in == nil {
		return nil, nil
	}

	subset, err := json.Marshal(in.Subset)
	if err != nil {
		return nil, errors.Wrap(err, "error while marshalling aspect event resource subset")
	}

	return &model.AspectEventResourceInput{
		OrdID:  in.OrdID,
		Subset: subset,
	}, nil
}

// MultipleInputFromGraphQL converts the provided graphql-layer representations of an Aspect Event Definition to the service-layer ones.
func (c *converter) MultipleInputFromGraphQL(in []*graphql.AspectEventDefinitionInput) ([]*model.AspectEventResourceInput, error) {
	inputs := make([]*model.AspectEventResourceInput, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}
		aspectEventIn, err := c.InputFromGraphQL(a)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, aspectEventIn)
	}

	return inputs, nil
}

// ToGraphQL converts the provided service-layer representation of an Aspect Event Resource to the graphql-layer one.
func (c *converter) ToGraphQL(in *model.AspectEventResource) (*graphql.AspectEventDefinition, error) {
	if in == nil {
		return nil, nil
	}

	var subset []*graphql.AspectEventDefinitionSubset
	if in.Subset != nil {
		if err := json.Unmarshal(in.Subset, &subset); err != nil {
			return nil, err
		}
	}

	return &graphql.AspectEventDefinition{
		OrdID:  in.OrdID,
		Subset: subset,
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

// MultipleToGraphQL converts the provided service-layer representations of an Aspect Event Resource to the graphql-layer ones.
func (c *converter) MultipleToGraphQL(in []*model.AspectEventResource) ([]*graphql.AspectEventDefinition, error) {
	aspectEvents := make([]*graphql.AspectEventDefinition, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}

		aspectEvent, err := c.ToGraphQL(a)
		if err != nil {
			return nil, err
		}

		aspectEvents = append(aspectEvents, aspectEvent)
	}

	return aspectEvents, nil
}

func timePtrToTimestampPtr(time *time.Time) *graphql.Timestamp {
	if time == nil {
		return nil
	}

	t := graphql.Timestamp(*time)
	return &t
}
