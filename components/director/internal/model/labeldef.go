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
		if _, err := jsonschema.NewValidatorFromRawSchema(def.Schema); err != nil {
			return errors.Wrapf(err, "while validating schema: [%v]", def.Schema)
		}
	}

	return nil
}

func (def *LabelDefinition) ValidateForUpdate() error {
	if def.Tenant == "" {
		return errors.New("missing Tenant field")
	}

	if def.Key == "" {
		return errors.New("missing Key field")
	}
	if def.Schema != nil {
		if _, err := jsonschema.NewValidatorFromRawSchema(def.Schema); err != nil {
			return errors.Wrapf(err, "while validating schema: [%v]", def.Schema)
		}
	}

	return nil
}
