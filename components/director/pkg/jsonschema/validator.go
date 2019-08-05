package jsonschema

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

type validator struct {
	schema *gojsonschema.Schema
}

func NewValidatorFromStringSchema(jsonSchema string) (*validator, error) {
	var schema *gojsonschema.Schema
	var err error

	if jsonSchema != "" {
		sl := gojsonschema.NewStringLoader(jsonSchema)
		schema, err = gojsonschema.NewSchema(sl)
		if err != nil {
			return nil, err
		}
	}

	return &validator{
		schema: schema,
	}, nil
}

func NewValidatorFromRawSchema(jsonSchema interface{}) (*validator, error) {
	if jsonSchema == nil {
		return &validator{}, nil
	}

	var schema *gojsonschema.Schema
	var err error

	sl := gojsonschema.NewGoLoader(jsonSchema)
	schema, err = gojsonschema.NewSchema(sl)
	if err != nil {
		return nil, err
	}

	return &validator{
		schema: schema,
	}, nil
}

func (v *validator) ValidateString(json string) (bool, error) {
	if v.schema == nil {
		return true, nil
	}

	jsonLoader := gojsonschema.NewStringLoader(json)
	result, err := v.schema.Validate(jsonLoader)
	if err != nil {
		return false, err
	}

	return result.Valid(), nil
}

func (v *validator) ValidateRaw(value interface{}) (bool, error) {
	if v.schema == nil {
		return true, nil
	}

	valueString, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return v.ValidateString(string(valueString))
}
