package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

type LabelDefinition struct {
	ID     string
	Tenant string
	Key    string
	Schema *interface{}
}

func (def *LabelDefinition) Validate() error {
	if def.ID == "" {
		return errors.New("missing ID field")
	}

	if def.Tenant == "" {
		return errors.New("missing Tenant field")
	}

	if def.Key == "" {
		return errors.New("missing Key field")
	}
	if def.Schema != nil {
		if _, err := jsonschema.NewValidatorFromRawSchema(*def.Schema); err != nil {
			return errors.Wrapf(err, "while validating schema: [%+v]", *def.Schema)
		}
	}
	if def.Key == ScenariosKey {
		if err := def.validateScenariosSchema(); err != nil {
			return errors.Wrapf(err, "while validating schema for key %s", ScenariosKey)
		}
	}

	return nil
}

// TODO: Move this method to LabelDefinitionInput
func (def *LabelDefinition) ValidateForUpdate() error {
	if def.Tenant == "" {
		return errors.New("missing Tenant field")
	}

	if def.Key == "" {
		return errors.New("missing Key field")
	}

	if def.Schema != nil {
		if _, err := jsonschema.NewValidatorFromRawSchema(*def.Schema); err != nil {
			return errors.Wrapf(err, "while validating schema: [%+v]", *def.Schema)
		}
	}

	if def.Key == ScenariosKey {
		if err := def.validateScenariosSchema(); err != nil {
			return errors.Wrapf(err, "while validating schema for key %s", ScenariosKey)
		}
	}

	return nil
}

func (def *LabelDefinition) validateScenariosSchema() error {
	if def == nil || def.Schema == nil {
		return errors.New("schema can not be nil")
	}

	validator, err := jsonschema.NewValidatorFromRawSchema(SchemaForScenariosSchema)
	if err != nil {
		return errors.Wrap(err, "while compiling validator schema")
	}

	result, err := validator.ValidateRaw(*def.Schema)
	if err != nil {
		return errors.Wrap(err, "while validating new schema")
	}
	if !result.Valid {
		return result.Error
	}

	return nil
}
