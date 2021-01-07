package apptemplate

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/pkg/errors"
)

//go:generate mockery -name=AppConverter -output=automock -outpkg=automock -case=underscore
type AppConverter interface {
	CreateInputGQLToJSON(in *graphql.ApplicationRegisterInput) (string, error)
}

type converter struct {
	appConverter AppConverter
}

func NewConverter(appConverter AppConverter) *converter {
	return &converter{appConverter: appConverter}
}

func (c *converter) ToGraphQL(in *model.ApplicationTemplate) (*graphql.ApplicationTemplate, error) {
	if in == nil {
		return nil, nil
	}

	if in.ApplicationInputJSON == "" {
		return nil, apperrors.NewInternalError("application input is empty")
	}

	gqlAppInput, err := c.graphqliseApplicationCreateInput(in.ApplicationInputJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "while graphqlising application create input")
	}

	return &graphql.ApplicationTemplate{
		ID:               in.ID,
		Name:             in.Name,
		Description:      in.Description,
		ApplicationInput: gqlAppInput,
		Placeholders:     c.placeholdersToGraphql(in.Placeholders),
		AccessLevel:      graphql.ApplicationTemplateAccessLevel(in.AccessLevel),
	}, nil
}

func (c *converter) MultipleToGraphQL(in []*model.ApplicationTemplate) ([]*graphql.ApplicationTemplate, error) {
	var appTemplates []*graphql.ApplicationTemplate
	for _, r := range in {
		if r == nil {
			continue
		}

		appTemplate, err := c.ToGraphQL(r)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting application template")
		}
		appTemplates = append(appTemplates, appTemplate)
	}

	return appTemplates, nil
}

func (c *converter) InputFromGraphQL(in graphql.ApplicationTemplateInput) (model.ApplicationTemplateInput, error) {
	var appCreateInput string
	var err error
	if in.ApplicationInput != nil {
		appCreateInput, err = c.appConverter.CreateInputGQLToJSON(in.ApplicationInput)
		if err != nil {
			return model.ApplicationTemplateInput{}, errors.Wrapf(err, "error occurred while converting GraphQL input to Application Template model with name %s", in.Name)
		}
	}

	return model.ApplicationTemplateInput{
		Name:                 in.Name,
		Description:          in.Description,
		ApplicationInputJSON: appCreateInput,
		Placeholders:         c.placeholdersFromGraphql(in.Placeholders),
		AccessLevel:          model.ApplicationTemplateAccessLevel(in.AccessLevel),
	}, nil
}

func (c *converter) ApplicationFromTemplateInputFromGraphQL(in graphql.ApplicationFromTemplateInput) model.ApplicationFromTemplateInput {
	var values []*model.ApplicationTemplateValueInput
	for _, value := range in.Values {
		valueInput := model.ApplicationTemplateValueInput{
			Placeholder: value.Placeholder,
			Value:       value.Value,
		}
		values = append(values, &valueInput)
	}

	return model.ApplicationFromTemplateInput{
		TemplateName: in.TemplateName,
		Values:       values,
	}
}

func (c *converter) ToEntity(in *model.ApplicationTemplate) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	placeholders, err := c.placeholdersModelToJSON(in.Placeholders)
	if err != nil {
		return nil, errors.Wrap(err, "while converting placeholders from model to JSON")
	}

	return &Entity{
		ID:                   in.ID,
		Name:                 in.Name,
		Description:          repo.NewNullableString(in.Description),
		ApplicationInputJSON: in.ApplicationInputJSON,
		PlaceholdersJSON:     placeholders,
		AccessLevel:          string(in.AccessLevel),
	}, nil
}

func (c *converter) FromEntity(entity *Entity) (*model.ApplicationTemplate, error) {
	if entity == nil {
		return nil, nil
	}

	placeholders, err := c.placeholdersJSONToModel(entity.PlaceholdersJSON)
	if err != nil {
		return nil, errors.Wrap(err, "while converting placeholders from JSON to model")
	}

	return &model.ApplicationTemplate{
		ID:                   entity.ID,
		Name:                 entity.Name,
		Description:          repo.StringPtrFromNullableString(entity.Description),
		ApplicationInputJSON: entity.ApplicationInputJSON,
		Placeholders:         placeholders,
		AccessLevel:          model.ApplicationTemplateAccessLevel(entity.AccessLevel),
	}, nil
}

func (c *converter) graphqliseApplicationCreateInput(jsonAppInput string) (string, error) {
	var gqlAppCreateInput graphql.ApplicationRegisterInput
	err := json.Unmarshal([]byte(jsonAppInput), &gqlAppCreateInput)
	if err != nil {
		return "", errors.Wrap(err, "while unmarshaling application create input")
	}

	g := graphqlizer.Graphqlizer{}
	gqlAppInput, err := g.ApplicationRegisterInputToGQL(gqlAppCreateInput)
	if err != nil {
		return "", errors.Wrap(err, "while graphqlising application create input")
	}
	gqlAppInput = strings.Replace(gqlAppInput, "\t", "", -1)
	gqlAppInput = strings.Replace(gqlAppInput, "\n", "", -1)
	return gqlAppInput, nil
}

func (c *converter) placeholdersJSONToModel(in sql.NullString) ([]model.ApplicationTemplatePlaceholder, error) {
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

func (c *converter) placeholdersModelToJSON(in []model.ApplicationTemplatePlaceholder) (sql.NullString, error) {
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
