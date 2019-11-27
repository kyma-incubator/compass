package apptemplate

import (
	"database/sql"
	"encoding/json"
	"strings"

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

func (c *converter) ToGraphQL(in *model.ApplicationTemplate) (*graphql.ApplicationTemplate, error) {
	if in == nil {
		return nil, nil
	}

	var gqlAppInput string
	var err error
	if in.ApplicationInput != "" {
		gqlAppInput, err = c.graphqliseApplicationCreateInput(in.ApplicationInput)
		if err != nil {
			return nil, errors.Wrapf(err, "while graphqlising application create input")
		}
	} else {
		return nil, errors.New("application input is empty")
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
		appCreateInput, err = c.applicationCreateInputGQLToJSON(in.ApplicationInput)
		if err != nil {
			return model.ApplicationTemplateInput{}, errors.Wrap(err, "while packing GQL application input")
		}
	}

	return model.ApplicationTemplateInput{
		Name:             in.Name,
		Description:      in.Description,
		ApplicationInput: appCreateInput,
		Placeholders:     c.placeholdersFromGraphql(in.Placeholders),
		AccessLevel:      model.ApplicationTemplateAccessLevel(in.AccessLevel),
	}, nil
}

func (c *converter) ToEntity(in *model.ApplicationTemplate) (*Entity, error) {
	if in == nil {
		return nil, nil
	}

	placeholders, err := c.placeholdersModelToJSON(in.Placeholders)
	if err != nil {
		return nil, errors.Wrap(err, "while packing PlaceholdersJSON")
	}

	return &Entity{
		ID:                   in.ID,
		Name:                 in.Name,
		Description:          repo.NewNullableString(in.Description),
		ApplicationInputJSON: in.ApplicationInput,
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
		return nil, errors.Wrap(err, "while unpacking PlaceholdersJSON")
	}

	return &model.ApplicationTemplate{
		ID:               entity.ID,
		Name:             entity.Name,
		Description:      repo.StringPtrFromNullableString(entity.Description),
		ApplicationInput: entity.ApplicationInputJSON,
		Placeholders:     placeholders,
		AccessLevel:      model.ApplicationTemplateAccessLevel(entity.AccessLevel),
	}, nil
}

func (c *converter) graphqliseApplicationCreateInput(applicationInput string) (string, error) {
	var jsonAppInput graphql.ApplicationCreateInput
	err := json.Unmarshal([]byte(applicationInput), &jsonAppInput)
	if err != nil {
		return "", errors.Wrap(err, "while unmarshaling application create input")
	}

	g := Graphqlizer{}
	gqlAppInput, err := g.ApplicationCreateInputToGQL(jsonAppInput)
	if err != nil {
		return "", errors.Wrap(err, "while graphqlising application create input")
	}
	gqlAppInput = strings.Replace(gqlAppInput, "\t", "", -1)
	gqlAppInput = strings.Replace(gqlAppInput, "\n", "", -1)
	return gqlAppInput, nil
}

func (c *converter) applicationCreateInputJSONToModel(in string) (*model.ApplicationCreateInput, error) {
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

func (c *converter) applicationCreateInputModelToJSON(in *model.ApplicationCreateInput) (string, error) {
	if in == nil {
		return "", nil
	}

	result, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshalling Application input")
	}

	return string(result), nil
}

func (c *converter) applicationCreateInputGQLToJSON(in *graphql.ApplicationCreateInput) (string, error) {
	appInput, err := json.Marshal(in)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling application input")
	}

	return string(appInput), nil
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
