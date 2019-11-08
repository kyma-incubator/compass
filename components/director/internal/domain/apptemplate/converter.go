package apptemplate

import (
	"database/sql"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

type AppConverter interface {}

type converter struct{}

func NewConverter() *converter {
	return &converter{}
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
		return nil, errors.Wrap(err, "while unpacking placeholders")
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
		Placeholders:    placeholders,
		AccessLevel:      model.ApplicationTemplateAccessLevel(entity.AccessLevel),
	}, nil
}

func (c *converter) unpackApplicationInput(in string) (*model.ApplicationCreateInput, error) {
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
	if !in.Valid {
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

	result.Valid = true
	result.String = string(placeholdersMarshalled)

	return result, nil
}
