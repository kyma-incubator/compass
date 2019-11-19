package apptemplate

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

type AppConverter interface{}
type ApplicationConverter interface {
	CreateInputFromGraphQL(in graphql.ApplicationCreateInput) model.ApplicationCreateInput
}

type converter struct {
	applicationConverter ApplicationConverter
}

func NewConverter(applicationConverter ApplicationConverter) *converter {
	return &converter{
		applicationConverter: applicationConverter,
	}
}

func (c *converter) ToGraphQL(in *model.ApplicationTemplate) *graphql.ApplicationTemplate {
	if in == nil {
		return nil
	}
	appInput, err := c.applicationInputToString(in.ApplicationInput)
	if err != nil {
		return nil
	}

	return &graphql.ApplicationTemplate{
		ID:               in.ID,
		Name:             in.Name,
		Description:      in.Description,
		ApplicationInput: appInput,
		Placeholders:     c.placeholdersToGraphql(in.Placeholders),
		AccessLevel:      graphql.ApplicationTemplateAccessLevel(in.AccessLevel),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.ApplicationTemplate) []*graphql.ApplicationTemplate {
	var appTemplates []*graphql.ApplicationTemplate
	for _, r := range in {
		if r == nil {
			continue
		}

		appTemplates = append(appTemplates, c.ToGraphQL(r))
	}

	return appTemplates
}

func (c *converter) InputFromGraphQL(in graphql.ApplicationTemplateInput) model.ApplicationTemplateInput {
	appInput := c.applicationConverter.CreateInputFromGraphQL(*in.ApplicationInput)

	return model.ApplicationTemplateInput{
		Name:             in.Name,
		Description:      in.Description,
		ApplicationInput: &appInput,
		Placeholders:     c.placeholdersFromGraphql(in.Placeholders),
		AccessLevel:      model.ApplicationTemplateAccessLevel(in.AccessLevel),
	}
}

func (c *converter) ToEntity(in *model.ApplicationTemplate) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	placeholders, err := c.packPlaceholders(in.Placeholders)
	if err != nil {
		return nil, errors.Wrap(err, "while packing Placeholders")
	}

	appInput, err := c.packApplicationInput(in.ApplicationInput)
	if err != nil {
		return nil, errors.Wrap(err, "while packing Placeholders")
	}

	return &Entity{
		ID:               in.ID,
		Name:             in.Name,
		Description:      repo.NewNullableString(in.Description),
		ApplicationInput: appInput,
		Placeholders:     placeholders,
		AccessLevel:      string(in.AccessLevel),
	}, nil
}

func (c *converter) FromEntity(entity *Entity) (*model.ApplicationTemplate, error) {
	if entity == nil {
		return nil, nil
	}

	placeholders, err := c.unpackPlaceholders(entity.Placeholders)
	if err != nil {
		return nil, errors.Wrap(err, "while unpacking Placeholders")
	}

	appInput, err := c.unpackApplicationInput(entity.ApplicationInput)
	if err != nil {
		return nil, errors.Wrap(err, "while unpacking Application Create Input")
	}

	return &model.ApplicationTemplate{
		ID:               entity.ID,
		Name:             entity.Name,
		Description:      repo.StringPtrFromNullableString(entity.Description),
		ApplicationInput: appInput,
		Placeholders:     placeholders,
		AccessLevel:      model.ApplicationTemplateAccessLevel(entity.AccessLevel),
	}, nil
}

func (c *converter) unpackApplicationInput(in string) (*model.ApplicationCreateInput, error) {
	if in == "" {
		return nil, nil
	}

	var appInput model.ApplicationCreateInput
	err := json.Unmarshal([]byte(in), &appInput)
	if err != nil {
		return nil, err
	}

	return &appInput, nil
}

func (c *converter) packApplicationInput(in *model.ApplicationCreateInput) (string, error) {
	if in == nil {
		return "", nil
	}

	result, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshalling Application input")
	}

	return string(result), nil
}

func (c *converter) unpackPlaceholders(in sql.NullString) ([]model.ApplicationTemplatePlaceholder, error) {
	if !in.Valid || in.String == "" {
		return nil, nil
	}

	var placeholders []model.ApplicationTemplatePlaceholder
	err := json.Unmarshal([]byte(in.String), &placeholders)
	if err != nil {
		return nil, err
	}

	return placeholders, nil
}

func (c *converter) packPlaceholders(in []model.ApplicationTemplatePlaceholder) (sql.NullString, error) {
	result := sql.NullString{}

	if in == nil {
		return result, nil
	}

	placeholdersMarshalled, err := json.Marshal(in)
	if err != nil {
		return result, errors.Wrap(err, "while marshalling placeholders")
	}

	return repo.NewValidNullableString(string(placeholdersMarshalled)), nil
}

func (c *converter) applicationInputToString(in *model.ApplicationCreateInput) (string, error) {
	appInput, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling default auth")
	}
	return string(appInput), nil
}

func (c *converter) placeholdersFromGraphql(in []*graphql.PlaceholderDefinitionInput) []model.ApplicationTemplatePlaceholder {
	var placeholders []model.ApplicationTemplatePlaceholder
	for _, p := range in {
		np := model.ApplicationTemplatePlaceholder{
			Name:        p.Name,
			Description: p.Description,
		}
		placeholders = append(placeholders, np)
	}
	return placeholders
}

func (c *converter) placeholdersToGraphql(in []model.ApplicationTemplatePlaceholder) []*graphql.PlaceholderDefinition {
	var placeholders []*graphql.PlaceholderDefinition
	for _, p := range in {
		np := graphql.PlaceholderDefinition{
			Name:        p.Name,
			Description: p.Description,
		}
		placeholders = append(placeholders, &np)
	}

	return placeholders
}
