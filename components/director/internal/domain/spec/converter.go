package spec

import (
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// FetchRequestConverter missing godoc
//go:generate mockery --name=FetchRequestConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error)
	InputFromGraphQL(in *graphql.FetchRequestInput) (*model.FetchRequestInput, error)
}

type converter struct {
	fr FetchRequestConverter
}

// NewConverter missing godoc
func NewConverter(fr FetchRequestConverter) *converter {
	return &converter{
		fr: fr,
	}
}

// ToGraphQLAPISpec missing godoc
func (c *converter) ToGraphQLAPISpec(in *model.Spec) (*graphql.APISpec, error) {
	if in == nil {
		return nil, nil
	}

	if in.ObjectType != model.APISpecReference || in.APIType == nil {
		return nil, fmt.Errorf("could not convert %s Spec to API Spec with APIType %v", in.ObjectType, in.APIType)
	}

	var data *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB(*in.Data)
		data = &tmp
	}

	return &graphql.APISpec{
		ID:           in.ID,
		Data:         data,
		Format:       graphql.SpecFormat(in.Format),
		Type:         graphql.APISpecType(*in.APIType),
		DefinitionID: in.ObjectID,
	}, nil
}

// ToGraphQLEventSpec missing godoc
func (c *converter) ToGraphQLEventSpec(in *model.Spec) (*graphql.EventSpec, error) {
	if in == nil {
		return nil, nil
	}

	if in.ObjectType != model.EventSpecReference || in.EventType == nil {
		return nil, fmt.Errorf("could not convert %s Spec to Event Spec with EventType %v", in.ObjectType, in.EventType)
	}

	var data *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB(*in.Data)
		data = &tmp
	}

	return &graphql.EventSpec{
		ID:           in.ID,
		Data:         data,
		Format:       graphql.SpecFormat(in.Format),
		Type:         graphql.EventSpecType(*in.EventType),
		DefinitionID: in.ObjectID,
	}, nil
}

// InputFromGraphQLAPISpec missing godoc
func (c *converter) InputFromGraphQLAPISpec(in *graphql.APISpecInput) (*model.SpecInput, error) {
	if in == nil {
		return nil, nil
	}

	fetchReq, err := c.fr.InputFromGraphQL(in.FetchRequest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting FetchRequest from GraphQL input")
	}

	apiType := model.APISpecType(in.Type)

	return &model.SpecInput{
		Data:         (*string)(in.Data),
		APIType:      &apiType,
		Format:       model.SpecFormat(in.Format),
		FetchRequest: fetchReq,
	}, nil
}

// InputFromGraphQLEventSpec missing godoc
func (c *converter) InputFromGraphQLEventSpec(in *graphql.EventSpecInput) (*model.SpecInput, error) {
	if in == nil {
		return nil, nil
	}

	fetchReq, err := c.fr.InputFromGraphQL(in.FetchRequest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting FetchRequest from GraphQL input")
	}

	eventType := model.EventSpecType(in.Type)

	return &model.SpecInput{
		Data:         (*string)(in.Data),
		EventType:    &eventType,
		Format:       model.SpecFormat(in.Format),
		FetchRequest: fetchReq,
	}, nil
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Spec) *Entity {
	refID := repo.NewValidNullableString(in.ObjectID)

	var apiDefID sql.NullString
	var apiSpecFormat sql.NullString
	var apiSpecType sql.NullString

	var eventAPIDefID sql.NullString
	var eventSpecFormat sql.NullString
	var eventSpecType sql.NullString

	switch in.ObjectType {
	case model.APISpecReference:
		apiDefID = refID
		apiSpecFormat = repo.NewValidNullableString(string(in.Format))
		apiSpecType = repo.NewValidNullableString(string(*in.APIType))
	case model.EventSpecReference:
		eventAPIDefID = refID
		eventSpecFormat = repo.NewValidNullableString(string(in.Format))
		eventSpecType = repo.NewValidNullableString(string(*in.EventType))
	}

	return &Entity{
		ID:              in.ID,
		APIDefID:        apiDefID,
		EventAPIDefID:   eventAPIDefID,
		SpecData:        repo.NewNullableString(in.Data),
		APISpecFormat:   apiSpecFormat,
		APISpecType:     apiSpecType,
		EventSpecFormat: eventSpecFormat,
		EventSpecType:   eventSpecType,
		CustomType:      repo.NewNullableString(in.CustomType),
	}
}

// FromEntity missing godoc
func (c *converter) FromEntity(in *Entity) (*model.Spec, error) {
	objectID, objectType, err := c.objectReferenceFromEntity(*in)
	if err != nil {
		return nil, errors.Wrap(err, "while determining object reference")
	}

	var apiSpecFormat model.SpecFormat
	var apiSpecType *model.APISpecType

	var eventSpecFormat model.SpecFormat
	var eventSpecType *model.EventSpecType

	apiSpecFormatStr := repo.StringPtrFromNullableString(in.APISpecFormat)
	if apiSpecFormatStr != nil {
		apiSpecFormat = model.SpecFormat(*apiSpecFormatStr)
	}

	apiSpecTypeStr := repo.StringPtrFromNullableString(in.APISpecType)
	if apiSpecTypeStr != nil {
		apiType := model.APISpecType(*apiSpecTypeStr)
		apiSpecType = &apiType
	}

	eventSpecFormatStr := repo.StringPtrFromNullableString(in.EventSpecFormat)
	if eventSpecFormatStr != nil {
		eventSpecFormat = model.SpecFormat(*eventSpecFormatStr)
	}

	eventSpecTypeStr := repo.StringPtrFromNullableString(in.EventSpecType)
	if eventSpecTypeStr != nil {
		eventType := model.EventSpecType(*eventSpecTypeStr)
		eventSpecType = &eventType
	}

	specFormat := apiSpecFormat
	if objectType == model.EventSpecReference {
		specFormat = eventSpecFormat
	}

	return &model.Spec{
		ID:         in.ID,
		ObjectType: objectType,
		ObjectID:   objectID,
		Data:       repo.StringPtrFromNullableString(in.SpecData),
		Format:     specFormat,
		APIType:    apiSpecType,
		EventType:  eventSpecType,
		CustomType: repo.StringPtrFromNullableString(in.CustomType),
	}, nil
}

func (c *converter) objectReferenceFromEntity(in Entity) (string, model.SpecReferenceObjectType, error) {
	if in.APIDefID.Valid {
		return in.APIDefID.String, model.APISpecReference, nil
	}

	if in.EventAPIDefID.Valid {
		return in.EventAPIDefID.String, model.EventSpecReference, nil
	}

	return "", "", fmt.Errorf("incorrect Object Reference ID and its type for Entity with ID '%s'", in.ID)
}
