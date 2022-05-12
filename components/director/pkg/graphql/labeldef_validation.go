package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

// Validate missing godoc
func (i LabelDefinitionInput) Validate() error {
	return validation.Errors{
		"rule.validSchema": i.validateSchema(),
		"key":              validation.Validate(i.Key, validation.Required, validation.Match(scenarioRegExp)),
		"schema":           validation.Validate(i.Schema),
	}.Filter()
}

func (i LabelDefinitionInput) validateSchema() error {
	if i.Schema != nil {
		if _, err := jsonschema.NewValidatorFromStringSchema(string(*i.Schema)); err != nil {
			return errors.Wrapf(err, "while validating schema: [%+v]", *i.Schema)
		}
	}
	if i.Key == model.ScenariosKey {
		if err := i.validateScenariosSchema(); err != nil {
			return errors.Wrapf(err, "while validating schema for key %s", model.ScenariosKey)
		}
	}
	return nil
}

func (i LabelDefinitionInput) validateScenariosSchema() error {
	if i.Schema == nil {
		return apperrors.NewInternalError("schema can not be nil")
	}

	validator, err := jsonschema.NewValidatorFromRawSchema(model.SchemaForScenariosSchema)
	if err != nil {
		return errors.Wrap(err, "while compiling validator schema")
	}

	result, err := validator.ValidateString(string(*i.Schema))
	if err != nil {
		return errors.Wrap(err, "while validating new schema")
	}
	if !result.Valid {
		return result.Error
	}

	return nil
}
